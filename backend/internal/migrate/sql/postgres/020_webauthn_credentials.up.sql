-- Migration 020: webauthn_credentials — passkey credentials per user.
--
-- Passkeys лежат в ОТДЕЛЬНОЙ таблице (не в user_identities), потому что у них
-- свой набор полей (sign_count, AAGUID, attestation type, transports). Для
-- "last factor lockout protection" см. auth.CountAuthFactors — он суммирует
-- identities + passkeys + password.
--
-- credential_id хранится как BYTEA (raw bytes), но через PRIMARY KEY гарантируем
-- глобальную уникальность. Frontend получает id в base64url, мы декодируем.
--
-- sign_count — anti-cloning guard. WebAuthn-спека: каждое assertion должно
-- иметь sign_count > previously stored value. Хранится как BIGINT (uint32
-- максимум, BIGINT с запасом для будущего расширения spec'и).
--
-- public_key — COSE-encoded public key (CBOR). Хранится как BYTEA, декодируется
-- библиотекой go-webauthn при verify.
--
-- aaguid — AAGUID authenticator'а (Apple/Google/YubiKey и т.д.). Помогает UI
-- показать иконку девайса. UUID, не FK (Apple Keychain = 00000000-0000-0000-0000-000000000000).
--
-- transports — comma-separated list "internal,hybrid,usb,nfc,ble".

CREATE TABLE webauthn_credentials (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    credential_id   BYTEA NOT NULL UNIQUE,
    public_key      BYTEA NOT NULL,
    aaguid          UUID,
    sign_count      BIGINT NOT NULL DEFAULT 0,
    transports      TEXT NOT NULL DEFAULT '',
    nickname        TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_used_at    TIMESTAMPTZ
);

CREATE INDEX idx_webauthn_credentials_user ON webauthn_credentials(user_id);
