use anyhow::{Context, Result};
use parking_lot::Mutex;
use rusqlite::{params, Connection};
use std::path::Path;
use std::sync::Arc;

use super::crypto::LocalCrypto;
use super::event::Event;

pub struct LocalStore {
    conn: Mutex<Connection>,
    crypto: Arc<LocalCrypto>,
}

impl LocalStore {
    pub fn open(path: impl AsRef<Path>, crypto: Arc<LocalCrypto>) -> Result<Self> {
        let conn = Connection::open(path)?;
        conn.execute_batch("PRAGMA journal_mode = WAL; PRAGMA synchronous = NORMAL;")?;
        Self::migrate(&conn)?;
        Ok(Self {
            conn: Mutex::new(conn),
            crypto,
        })
    }

    fn migrate(conn: &Connection) -> Result<()> {
        let current: i64 = conn.query_row("PRAGMA user_version", [], |r| r.get(0))?;
        let migrations: &[(i64, &str)] = &[
            (
                1,
                r#"
                CREATE TABLE IF NOT EXISTS event_buffer (
                    id INTEGER PRIMARY KEY AUTOINCREMENT,
                    payload TEXT NOT NULL,
                    created_at INTEGER NOT NULL,
                    lease_until INTEGER
                );
                CREATE INDEX IF NOT EXISTS idx_event_buffer_lease
                    ON event_buffer(lease_until);
                "#,
            ),
            (
                2,
                r#"
                DROP TABLE IF EXISTS event_buffer;
                CREATE TABLE event_buffer (
                    id INTEGER PRIMARY KEY AUTOINCREMENT,
                    payload BLOB NOT NULL,
                    created_at INTEGER NOT NULL,
                    lease_until INTEGER
                );
                CREATE INDEX IF NOT EXISTS idx_event_buffer_lease
                    ON event_buffer(lease_until);
                "#,
            ),
        ];
        for (version, sql) in migrations {
            if *version > current {
                conn.execute_batch(sql)?;
                conn.execute_batch(&format!("PRAGMA user_version = {}", version))?;
                tracing::info!(version, "applied store migration");
            }
        }
        Ok(())
    }

    pub fn push(&self, event: &Event) -> Result<()> {
        let conn = self.conn.lock();
        let json = serde_json::to_vec(event).context("serialize event")?;
        let payload = self.crypto.seal(&json).context("seal event payload")?;
        let now = chrono::Utc::now().timestamp();
        conn.execute(
            "INSERT INTO event_buffer (payload, created_at) VALUES (?1, ?2)",
            params![payload, now],
        )?;
        Ok(())
    }

    pub fn lease_batch(&self, limit: usize, lease_secs: i64) -> Result<Vec<(i64, Event)>> {
        let conn = self.conn.lock();
        let now = chrono::Utc::now().timestamp();
        let lease_until = now + lease_secs;

        let mut stmt = conn.prepare(
            "SELECT id, payload FROM event_buffer
             WHERE lease_until IS NULL OR lease_until < ?1
             ORDER BY id ASC
             LIMIT ?2",
        )?;
        let rows: Vec<(i64, Vec<u8>)> = stmt
            .query_map(params![now, limit as i64], |r| Ok((r.get(0)?, r.get(1)?)))?
            .collect::<rusqlite::Result<_>>()?;

        let ids: Vec<i64> = rows.iter().map(|(id, _)| *id).collect();
        if !ids.is_empty() {
            let placeholders = ids.iter().map(|_| "?").collect::<Vec<_>>().join(",");
            let sql = format!(
                "UPDATE event_buffer SET lease_until = {} WHERE id IN ({})",
                lease_until, placeholders
            );
            let params: Vec<Box<dyn rusqlite::ToSql>> = ids
                .iter()
                .map(|i| Box::new(*i) as Box<dyn rusqlite::ToSql>)
                .collect();
            let refs: Vec<&dyn rusqlite::ToSql> = params.iter().map(|b| b.as_ref()).collect();
            conn.execute(&sql, refs.as_slice())?;
        }

        let mut out = Vec::with_capacity(rows.len());
        let mut corrupted: Vec<i64> = Vec::new();
        for (id, payload) in rows {
            let json = match self.crypto.open(&payload) {
                Ok(b) => b,
                Err(e) => {
                    tracing::warn!(id, error = %e, "decrypt failed, dropping row");
                    corrupted.push(id);
                    continue;
                }
            };
            match serde_json::from_slice::<Event>(&json) {
                Ok(ev) => out.push((id, ev)),
                Err(e) => {
                    tracing::warn!(id, error = %e, "deserialize failed after decrypt, dropping row");
                    corrupted.push(id);
                }
            }
        }
        if !corrupted.is_empty() {
            let placeholders = corrupted.iter().map(|_| "?").collect::<Vec<_>>().join(",");
            let sql = format!("DELETE FROM event_buffer WHERE id IN ({})", placeholders);
            let params: Vec<Box<dyn rusqlite::ToSql>> = corrupted
                .iter()
                .map(|i| Box::new(*i) as Box<dyn rusqlite::ToSql>)
                .collect();
            let refs: Vec<&dyn rusqlite::ToSql> = params.iter().map(|b| b.as_ref()).collect();
            let _ = conn.execute(&sql, refs.as_slice());
        }
        Ok(out)
    }

    pub fn commit(&self, ids: &[i64]) -> Result<()> {
        if ids.is_empty() {
            return Ok(());
        }
        let conn = self.conn.lock();
        let placeholders = ids.iter().map(|_| "?").collect::<Vec<_>>().join(",");
        let sql = format!("DELETE FROM event_buffer WHERE id IN ({})", placeholders);
        let params: Vec<Box<dyn rusqlite::ToSql>> = ids
            .iter()
            .map(|i| Box::new(*i) as Box<dyn rusqlite::ToSql>)
            .collect();
        let refs: Vec<&dyn rusqlite::ToSql> = params.iter().map(|b| b.as_ref()).collect();
        conn.execute(&sql, refs.as_slice())?;
        Ok(())
    }

    pub fn release(&self, ids: &[i64]) -> Result<()> {
        if ids.is_empty() {
            return Ok(());
        }
        let conn = self.conn.lock();
        let placeholders = ids.iter().map(|_| "?").collect::<Vec<_>>().join(",");
        let sql = format!(
            "UPDATE event_buffer SET lease_until = NULL WHERE id IN ({})",
            placeholders
        );
        let params: Vec<Box<dyn rusqlite::ToSql>> = ids
            .iter()
            .map(|i| Box::new(*i) as Box<dyn rusqlite::ToSql>)
            .collect();
        let refs: Vec<&dyn rusqlite::ToSql> = params.iter().map(|b| b.as_ref()).collect();
        conn.execute(&sql, refs.as_slice())?;
        Ok(())
    }

    pub fn pending_count(&self) -> Result<i64> {
        let conn = self.conn.lock();
        let n: i64 = conn.query_row("SELECT COUNT(*) FROM event_buffer", [], |r| r.get(0))?;
        Ok(n)
    }

    pub fn gc(&self, older_than_secs: i64) -> Result<usize> {
        let conn = self.conn.lock();
        let cutoff = chrono::Utc::now().timestamp() - older_than_secs;
        let n = conn.execute(
            "DELETE FROM event_buffer WHERE created_at < ?1",
            params![cutoff],
        )?;
        Ok(n)
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::core::event::{Category, Event};

    fn tmp_store() -> (LocalStore, tempfile::NamedTempFile) {
        let f = tempfile::NamedTempFile::new().unwrap();
        let crypto = Arc::new(LocalCrypto::with_key([7u8; 32]));
        let s = LocalStore::open(f.path(), crypto).unwrap();
        (s, f)
    }

    fn ev(app: &str) -> Event {
        Event::os_focus(app, Category::Other, 1000)
    }

    #[test]
    fn migrations_set_user_version() {
        let f = tempfile::NamedTempFile::new().unwrap();
        let crypto = Arc::new(LocalCrypto::with_key([7u8; 32]));
        let s = LocalStore::open(f.path(), crypto).unwrap();
        let v: i64 = s
            .conn
            .lock()
            .query_row("PRAGMA user_version", [], |r| r.get(0))
            .unwrap();
        assert!(
            v >= 2,
            "user_version should be bumped after migrate, got {v}"
        );
    }

    #[test]
    fn payload_column_is_blob_not_text() {
        let (s, _f) = tmp_store();
        let conn = s.conn.lock();
        let type_: String = conn
            .query_row(
                "SELECT type FROM pragma_table_info('event_buffer') WHERE name = 'payload'",
                [],
                |r| r.get(0),
            )
            .unwrap();
        assert_eq!(type_, "BLOB");
    }

    #[test]
    fn push_and_pending_count() {
        let (s, _f) = tmp_store();
        assert_eq!(s.pending_count().unwrap(), 0);
        s.push(&ev("a")).unwrap();
        s.push(&ev("b")).unwrap();
        assert_eq!(s.pending_count().unwrap(), 2);
    }

    // Regression: со std::sync::Mutex паника внутри одного потока, удерживающего
    // lock, отравляла Mutex — все последующие .lock().unwrap() в других потоках
    // также паниковали и убивали watcher/tray/ingest loop. parking_lot::Mutex
    // не отравляется: паника одного потока не влияет на acquire'ы других.
    #[test]
    fn mutex_does_not_poison_on_panic() {
        use std::sync::Arc as StdArc;
        let (s, _f) = tmp_store();
        let s = StdArc::new(s);

        // Поток 1: берёт lock, паникует пока держит его.
        let s1 = StdArc::clone(&s);
        let join = std::thread::spawn(move || {
            let _guard = s1.conn.lock();
            panic!("intentional panic while holding the lock");
        });
        let _ = join.join(); // Ожидаем Err — поток упал, это норм.

        // Поток 2: после паники должен по-прежнему успешно работать со store.
        // Со std::sync::Mutex здесь .pending_count() паниковал бы;
        // с parking_lot — отдаёт реальный ответ.
        let n = s.pending_count().expect("pending_count after poisoned thread");
        assert_eq!(n, 0);
        s.push(&ev("post-panic")).expect("push after poisoned thread");
        assert_eq!(s.pending_count().unwrap(), 1);
    }

    #[test]
    fn stored_payload_is_ciphertext_not_plaintext() {
        let (s, _f) = tmp_store();
        s.push(&ev("com.apple.dt.Xcode")).unwrap();
        let conn = s.conn.lock();
        let blob: Vec<u8> = conn
            .query_row("SELECT payload FROM event_buffer LIMIT 1", [], |r| r.get(0))
            .unwrap();
        let as_str = String::from_utf8_lossy(&blob);
        assert!(
            !as_str.contains("com.apple.dt.Xcode"),
            "ciphertext must not leak app_bundle string"
        );
        assert!(
            !as_str.contains("app_bundle"),
            "ciphertext must not leak field names"
        );
    }

    #[test]
    fn lease_then_commit_clears_rows() {
        let (s, _f) = tmp_store();
        s.push(&ev("a")).unwrap();
        s.push(&ev("b")).unwrap();
        let batch = s.lease_batch(10, 60).unwrap();
        assert_eq!(batch.len(), 2);
        let ids: Vec<i64> = batch.iter().map(|(id, _)| *id).collect();
        s.commit(&ids).unwrap();
        assert_eq!(s.pending_count().unwrap(), 0);
    }

    #[test]
    fn lease_excludes_already_leased_rows() {
        let (s, _f) = tmp_store();
        s.push(&ev("a")).unwrap();
        s.push(&ev("b")).unwrap();
        let first = s.lease_batch(10, 60).unwrap();
        assert_eq!(first.len(), 2);
        let second = s.lease_batch(10, 60).unwrap();
        assert_eq!(second.len(), 0);
    }

    #[test]
    fn release_makes_rows_available_again() {
        let (s, _f) = tmp_store();
        s.push(&ev("a")).unwrap();
        let first = s.lease_batch(10, 60).unwrap();
        let ids: Vec<i64> = first.iter().map(|(id, _)| *id).collect();
        s.release(&ids).unwrap();
        let second = s.lease_batch(10, 60).unwrap();
        assert_eq!(second.len(), 1);
    }

    #[test]
    fn gc_removes_old_rows_only() {
        let (s, _f) = tmp_store();
        s.push(&ev("fresh")).unwrap();
        {
            let conn = s.conn.lock();
            let blob = s
                .crypto
                .seal(b"{\"fake\":\"old\"}")
                .expect("seal fake payload");
            conn.execute(
                "INSERT INTO event_buffer (payload, created_at) VALUES (?1, ?2)",
                params![blob, chrono::Utc::now().timestamp() - 99_999],
            )
            .unwrap();
        }
        let removed = s.gc(60_000).unwrap();
        assert_eq!(removed, 1);
        assert_eq!(s.pending_count().unwrap(), 1);
    }

    #[test]
    fn corrupted_row_is_dropped_not_blocking() {
        let (s, _f) = tmp_store();
        s.push(&ev("good")).unwrap();
        {
            let conn = s.conn.lock();
            conn.execute(
                "INSERT INTO event_buffer (payload, created_at) VALUES (?1, ?2)",
                params![vec![0u8; 50], chrono::Utc::now().timestamp()],
            )
            .unwrap();
        }
        assert_eq!(s.pending_count().unwrap(), 2);
        let batch = s.lease_batch(10, 60).unwrap();
        assert_eq!(batch.len(), 1);
        assert_eq!(s.pending_count().unwrap(), 1);
    }
}
