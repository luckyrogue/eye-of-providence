// Local SQLite-буфер событий: WAL mode, retry-friendly, encrypted-at-rest.
// Pop возвращает события + lease_id, чтобы при ошибке отправки можно было
// поставить их обратно в очередь без потерь.
//
// Encryption:
//   Каждая запись `event_buffer.payload` — AES-256-GCM ciphertext поверх
//   serde_json(Event). Ключ хранится в Keychain/DPAPI через `core/crypto`.
//   См. crypto.rs для формата и threat model.

use anyhow::{Context, Result};
use rusqlite::{params, Connection};
use std::path::Path;
use std::sync::{Arc, Mutex};

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

    // migrate — версионируем SQLite через PRAGMA user_version. Каждая миграция
    // выполняется один раз; user_version подтверждает factual состояние схемы.
    // Чтобы добавить новую — append-only: добавьте match-ветку с очередным
    // номером, не редактируя предыдущие.
    //
    // v1 → v2: payload TEXT (plaintext JSON) → payload BLOB (AES-GCM ciphertext).
    // Старые plaintext-записи дропаем — это локальный send-buffer, потеря
    // unsent данных приемлема (события обычно отправляются за 15 секунд).
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
                // Шифрование at-rest. Колонка теперь BLOB; старые plaintext-
                // записи дропаем (одноразовая «потеря» unsent buffer'а
                // при апгрейде агента — acceptable).
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
        let conn = self.conn.lock().unwrap();
        let json = serde_json::to_vec(event).context("serialize event")?;
        let payload = self.crypto.seal(&json).context("seal event payload")?;
        let now = chrono::Utc::now().timestamp();
        conn.execute(
            "INSERT INTO event_buffer (payload, created_at) VALUES (?1, ?2)",
            params![payload, now],
        )?;
        Ok(())
    }

    /// Берёт до `limit` событий из буфера, проставляет lease до `lease_secs` секунд вперёд.
    /// Возвращает (id, Event) — id нужен для commit/release.
    ///
    /// Если запись не расшифровывается (например, ключ был ротирован вручную,
    /// или диск битый) — событие тихо дропается и логируется. Альтернатива
    /// (вернуть ошибку и остановить весь pump) хуже: один corrupted row
    /// заблокирует всю очередь.
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
        // Чистим corrupted строки чтобы pump не тыкался в них на каждом цикле.
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
    //
    // Crypto в тестах — детерминированный test-key через `LocalCrypto::with_key`,
    // не трогает системный keyring.
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
            .unwrap()
            .query_row("PRAGMA user_version", [], |r| r.get(0))
            .unwrap();
        assert!(
            v >= 2,
            "user_version should be bumped after migrate, got {v}"
        );
    }

    #[test]
    fn payload_column_is_blob_not_text() {
        // Гарантирует что v2-миграция реально применилась — payload должен
        // быть BLOB, чтобы plaintext-чтение через `sqlite3 .dump` не выдало
        // содержимое событий.
        let (s, _f) = tmp_store();
        let conn = s.conn.lock().unwrap();
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

    #[test]
    fn stored_payload_is_ciphertext_not_plaintext() {
        // Critical: дамп SQLite не должен содержать `app_bundle` или другие
        // поля Event в plaintext.
        let (s, _f) = tmp_store();
        s.push(&ev("com.apple.dt.Xcode")).unwrap();
        let conn = s.conn.lock().unwrap();
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
        // Имитируем старый event, выставляя created_at вручную. Payload =
        // valid encrypted blob с другим event'ом, чтобы migrate-схема BLOB
        // приняла INSERT (раньше тут писали `'{}'` plaintext, теперь не пройдёт).
        {
            let conn = s.conn.lock().unwrap();
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
        // Вставляем мусор как payload — decrypt должен упасть, и lease_batch
        // должен дропнуть строку, продолжив с «good».
        {
            let conn = s.conn.lock().unwrap();
            conn.execute(
                "INSERT INTO event_buffer (payload, created_at) VALUES (?1, ?2)",
                params![vec![0u8; 50], chrono::Utc::now().timestamp()],
            )
            .unwrap();
        }
        assert_eq!(s.pending_count().unwrap(), 2);
        let batch = s.lease_batch(10, 60).unwrap();
        // Один валидный + один corrupted дропнут → возвращается 1.
        assert_eq!(batch.len(), 1);
        // Corrupted строка должна быть удалена, остаётся только leased good.
        assert_eq!(s.pending_count().unwrap(), 1);
    }
}
