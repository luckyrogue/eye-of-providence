-- sso_configs — per-team SSO/OIDC configuration. SAML добавится в отдельной
-- column block'е в будущей миграции (issuer/sso_url/x509_cert).
--
-- Per-team scoping: enterprise team может настроить свой IdP, free/pro
-- юзают password / GitHub OAuth. Один team — один IdP (uniqueness через
-- team_id PK).
--
-- jit_provision_role: что назначать новым юзерам которые впервые logini'ются
-- через SSO ('member' default). Owner/admin promotion идёт через team UI.

CREATE TABLE IF NOT EXISTS sso_configs (
    team_id            UUID PRIMARY KEY REFERENCES teams(id) ON DELETE CASCADE,
    provider           TEXT NOT NULL CHECK (provider IN ('oidc', 'saml')),
    enabled            BOOLEAN NOT NULL DEFAULT false,

    -- OIDC fields. Discovery URL = issuer + /.well-known/openid-configuration.
    oidc_issuer        TEXT,
    oidc_client_id     TEXT,
    oidc_client_secret TEXT,        -- зашифрован на rest? Phase 2: вынести в KMS.
    oidc_scopes        TEXT[],      -- default ['openid','email','profile']

    -- SAML fields (Phase 2 — заглушки). idp_metadata = XML.
    saml_idp_metadata  TEXT,
    saml_sp_entity_id  TEXT,
    saml_acs_url       TEXT,

    -- Allowed email domains: если пусто — любой кто прошёл IdP. Иначе
    -- email после auth ДОЛЖЕН быть в одном из доменов (защита от
    -- misconfigured IdP отдающего external emails).
    allowed_domains    TEXT[],

    -- JIT провижининг. Если false — юзер должен быть pre-invited.
    jit_provision      BOOLEAN NOT NULL DEFAULT true,
    jit_role           TEXT NOT NULL DEFAULT 'member'
                         CHECK (jit_role IN ('owner', 'admin', 'member')),

    created_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- sso_states — short-lived (10min) CSRF tokens для OIDC state параметра.
-- В отличие от GitHub OAuth (cookie state) у нас нет cookie сессии до
-- authentication — state хранится в БД с TTL.
CREATE TABLE IF NOT EXISTS sso_states (
    state       TEXT PRIMARY KEY,
    team_id     UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    nonce       TEXT NOT NULL,                -- OIDC nonce claim для ID token
    return_to   TEXT,                          -- куда редирект после auth ("" = /dashboard)
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at  TIMESTAMPTZ NOT NULL
);

-- TTL cleanup index (background job или admin endpoint может truncate
-- expired rows; для prod использовать pg_cron).
CREATE INDEX IF NOT EXISTS sso_states_expires_idx ON sso_states (expires_at);

-- users.sso_subject — стабильный subject identifier от IdP'а. Позволяет
-- distinguish юзера если email меняется на стороне IdP'а. Composite key
-- (team_id, sso_subject) уникален.
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS sso_team_id UUID REFERENCES teams(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS sso_subject TEXT;

CREATE UNIQUE INDEX IF NOT EXISTS users_sso_subject_idx
    ON users (sso_team_id, sso_subject)
    WHERE sso_subject IS NOT NULL;
