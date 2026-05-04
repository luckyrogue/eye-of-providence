-- Postgres: метаданные пользователей, устройств, проектов, согласий.

CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE users (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email        TEXT UNIQUE NOT NULL,
    name         TEXT,
    github_login TEXT,
    role         TEXT,
    team_id      UUID,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at   TIMESTAMPTZ
);

CREATE TABLE teams (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name          TEXT NOT NULL,
    plan          TEXT NOT NULL DEFAULT 'free',
    settings_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE devices (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    os          TEXT NOT NULL,        -- macos | windows
    hostname    TEXT,
    fingerprint TEXT NOT NULL,
    last_seen   TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, fingerprint)
);

CREATE TABLE projects (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id        UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    repo_url       TEXT,
    root_path_hash TEXT NOT NULL,
    lang_primary   TEXT,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, root_path_hash)
);

CREATE TABLE consent (
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    scope      TEXT NOT NULL,
    granted_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    revoked_at TIMESTAMPTZ,
    PRIMARY KEY (user_id, scope)
);

CREATE TABLE api_tokens (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    scope        TEXT NOT NULL,
    hashed_token TEXT NOT NULL UNIQUE,
    expires_at   TIMESTAMPTZ,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE reports (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id        UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    period         TEXT NOT NULL,    -- weekly_2026_W18 | monthly_2026_05
    model          TEXT NOT NULL,    -- gemini-2.5-flash | ...
    body_md        TEXT NOT NULL,
    prompt_version TEXT NOT NULL,
    generated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_reports_user_period ON reports(user_id, period);
