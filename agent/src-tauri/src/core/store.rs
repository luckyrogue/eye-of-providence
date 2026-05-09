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
        conn.execute_batch(
            r#"
            PRAGMA journal_mode = WAL;
            PRAGMA synchronous = NORMAL;
            CREATE TABLE IF NOT EXISTS event_buffer (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                payload TEXT NOT NULL,
                created_at INTEGER NOT NULL,
                lease_until INTEGER
            );
            CREATE INDEX IF NOT EXISTS idx_event_buffer_lease
                ON event_buffer(lease_until);
            "#,
        )?;
        Ok(Self {
            conn: Mutex::new(conn),
        })
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
}
