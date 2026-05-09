-- Password reset tokens.
-- Хранится только хеш токена (sha256 hex). Plaintext-токен виден ровно один раз
-- в email-ссылке; даже компрометация БД не даёт прямого восстановления паролей.
--
-- expires_at — короткий TTL (1 час) для защиты от долгоживущих ссылок.
-- used_at    — после использования reset помечается, повторно нельзя.

CREATE TABLE IF NOT EXISTS password_resets (
  token_hash TEXT PRIMARY KEY,
  user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  expires_at TIMESTAMPTZ NOT NULL,
  used_at    TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_password_resets_user ON password_resets(user_id);
