-- Public API tokens — metadata для UI listing + revocation tracking.
--
-- name        — пользовательская подпись ("CI exporter", "personal BI")
-- prefix      — первые 8 chars plain token, чтобы UI мог показать
--               "eop_a4f3..." без хранения секрета (как GitHub PATs)
-- last_used_at — обновляется middleware'ом для UI "когда последний раз использовался"
-- revoked_at   — soft-delete, чтобы аудит/incident response мог посмотреть историю
ALTER TABLE api_tokens
  ADD COLUMN IF NOT EXISTS name         TEXT NOT NULL DEFAULT 'token',
  ADD COLUMN IF NOT EXISTS prefix       TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS last_used_at TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS revoked_at   TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_api_tokens_user_active
  ON api_tokens(user_id)
  WHERE revoked_at IS NULL;
