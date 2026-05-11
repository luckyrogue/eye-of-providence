-- Pairing-code flow для нативных клиентов (browser ext, Tauri agent, VS Code).
--
-- Существующая таблица api_tokens хранит долгоживущие токены. Здесь:
--   1) добавляем api_tokens.kind — отличает paired device от обычного API-token
--      (NULL/empty = классический API-token, "ext"/"agent"/"ide" = paired).
--   2) создаём pairing_codes — эфемерное состояние во время handshake'а
--      (живёт <=10 минут, удаляется по claim/expiry).
--
-- Flow:
--   POST /v1/devices/pair  (unauthed) → выдаёт {pair_id, secret, code}
--   POST /v1/devices/poll  (unauthed) → status pending/expired или token при claim
--   POST /v1/me/devices/claim (authed) → создаёт api_token kind=<kind>, scope=write:ingest
--   GET  /v1/me/devices    (authed) → list api_tokens WHERE kind IS NOT NULL
--   DEL  /v1/me/devices/:id (authed) → soft-delete (RevokeAPIToken)

ALTER TABLE api_tokens
  ADD COLUMN IF NOT EXISTS kind TEXT;

CREATE INDEX IF NOT EXISTS idx_api_tokens_kind_user
  ON api_tokens(user_id, kind)
  WHERE kind IS NOT NULL AND revoked_at IS NULL;

CREATE TABLE IF NOT EXISTS pairing_codes (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code              CHAR(6) NOT NULL UNIQUE,
    secret_hash       TEXT NOT NULL,
    kind              TEXT NOT NULL,
    name_hint         TEXT,
    code_expires_at   TIMESTAMPTZ NOT NULL,
    claimed_token_id  UUID REFERENCES api_tokens(id) ON DELETE SET NULL,
    claimed_plaintext TEXT,  -- plaintext API token; зачищается после первого успешного poll
    claimed_at        TIMESTAMPTZ,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_pairing_codes_expires
  ON pairing_codes(code_expires_at);
