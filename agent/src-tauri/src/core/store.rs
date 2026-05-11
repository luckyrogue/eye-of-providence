// Local SQLite-буфер событий: WAL mode, retry-friendly.
// Pop возвращает события + lease_id, чтобы при ошибке отправки можно было
// поставить их обратно в очередь без потерь.

use anyhow::Result;
use rusqlite::{params, Connection};
use std::path::Path;
use std::sync::Mutex;

use super::event::Event;

pub struct LocalStore {
    conn: Mutex<Connection>,
}

impl LocalStore {
    pub fn open(path: impl AsRef<Path>) -> Result<Self> {
        let conn = Connection::open(path)?;
        conn.execute_batch("PRAGMA journal_mode = WAL; PRAGMA synchronous = NORMAL;")?;
        Self::migrate(&conn)?;
        Ok(Self {
            conn: Mutex::new(conn),
        })
    }

    // migrate — версионируем SQLite через PRAGMA user_version. Каждая миграция
    // выполняется один раз; user_version подтверждает factual состояние схемы.
    // Чтобы добавить новую — append-only: добавьте match-ветку с очередным
    // номером, не редактируя предыдущие.
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
        let conn = self.conn.lock().unwrap();
        let payload = serde_json::to_string(event)?;
        let now = chrono::Utc::now().timestamp();
        conn.execute(
            "INSERT INTO event_buffer (payload, created_at) VALUES (?1, ?2)",
            params![payload, now],
        )?;
        Ok(())
    }

    /// Берёт до `limit` событий из буфера, проставляет lease до `lease_secs` секунд вперёд.
    /// Возвращает (id, Event) — id нужен для commit/release.
    pub fn lease_batch(&self, limit: usize, lease_secs: i64) -> Result<Vec<(i64, Event)>> {
        let conn = self.conn.lock().unwrap();
        let now = chrono::Utc::now().timestamp();
        let lease_until = now + lease_secs;

        let mut stmt = conn.prepare(
            "SELECT id, payload FROM event_buffer
             WHERE lease_until IS NULL OR lease_until < ?1
             ORDER BY id ASC
             LIMIT ?2",
        )?;
        let rows: Vec<(i64, String)> = stmt
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
        for (id, payload) in rows {
            let ev: Event = serde_json::from_str(&payload)?;
            out.push((id, ev));
        }
        Ok(out)
    }

    /// commit — окончательно удаляет события (после успешной отправки).
    pub fn commit(&self, ids: &[i64]) -> Result<()> {
        if ids.is_empty() {
            return Ok(());
        }
        let conn = self.conn.lock().unwrap();
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

    /// release — снимает lease, чтобы события снова стали доступны (после ошибки отправки).
    pub fn release(&self, ids: &[i64]) -> Result<()> {
        if ids.is_empty() {
            return Ok(());
        }
        let conn = self.conn.lock().unwrap();
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
        let conn = self.conn.lock().unwrap();
        let n: i64 = conn.query_row("SELECT COUNT(*) FROM event_buffer", [], |r| r.get(0))?;
        Ok(n)
    }

    /// gc — удаляет события старше `older_than_secs` секунд. Защита от
    /// бесконечного роста SQLite если backend недоступен длительное время
    /// (offline > week → ~360k events × 5s polling = пара GB локально).
    /// Возвращает количество удалённых строк.
    pub fn gc(&self, older_than_secs: i64) -> Result<usize> {
        let conn = self.conn.lock().unwrap();
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

    // tmp_store — открывает SQLite на in-memory path. Для unit-тестов не
    // нужен реальный файл; rusqlite поддерживает ":memory:" but миграции
    // и pragma WAL требуют named file → используем tempfile.
    fn tmp_store() -> (LocalStore, tempfile::NamedTempFile) {
        let f = tempfile::NamedTempFile::new().unwrap();
        let s = LocalStore::open(f.path()).unwrap();
        (s, f)
    }

    fn ev(app: &str) -> Event {
        Event::os_focus(app, Category::Other, 1000)
    }

    #[test]
    fn migrations_set_user_version() {
        let f = tempfile::NamedTempFile::new().unwrap();
        let s = LocalStore::open(f.path()).unwrap();
        let v: i64 = s
            .conn
            .lock()
            .unwrap()
            .query_row("PRAGMA user_version", [], |r| r.get(0))
            .unwrap();
        assert!(v >= 1, "user_version should be bumped after migrate, got {v}");
    }

    #[test]
    fn push_and_pending_count() {
        let (s, _f) = tmp_store();
        assert_eq!(s.pending_count().unwrap(), 0);
        s.push(&ev("a")).unwrap();
        s.push(&ev("b")).unwrap();
        assert_eq!(s.pending_count().unwrap(), 2);
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
        // Повторный lease не должен видеть свежие активные lease'ы.
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
        // Свежий event — должен остаться.
        s.push(&ev("fresh")).unwrap();
        // Имитируем старый event, выставляя created_at вручную.
        {
            let conn = s.conn.lock().unwrap();
            conn.execute(
                "INSERT INTO event_buffer (payload, created_at) VALUES ('{}', ?1)",
                params![chrono::Utc::now().timestamp() - 99_999],
            )
            .unwrap();
        }
        let removed = s.gc(60_000).unwrap();
        assert_eq!(removed, 1);
        assert_eq!(s.pending_count().unwrap(), 1);
    }
}
