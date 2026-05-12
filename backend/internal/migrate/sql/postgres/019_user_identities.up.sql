-- Migration 019: таблица user_identities — связь между внутренним user.id и
-- внешними провайдерами (github / google / apple / passkey).
--
-- Один user может иметь несколько identities (один email = github + google
-- + apple linked). Уникальность (provider, subject) гарантирует, что один
-- внешний аккаунт не привязан сразу к двум локальным юзерам.
--
-- Backfill (idempotent): для каждого users.github_login IS NOT NULL —
-- создаём identity-строку provider='github', subject=github_login.
-- При повторном запуске миграции ON CONFLICT DO NOTHING сохраняет идемпотентность.
--
-- Note: после Phase 2 (когда новые github-логины пойдут через unified flow),
-- subject будет числовой github ID, не login. Backfilled значения с login'ом
-- останутся как историческая запись — поиск по новому subject не сматчится,
-- но email-based linking в callback подберёт правильный user_id.

CREATE TABLE user_identities (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider   TEXT NOT NULL CHECK (provider IN ('github', 'google', 'apple', 'passkey')),
    subject    TEXT NOT NULL,
    email      TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (provider, subject)
);

CREATE INDEX idx_user_identities_user ON user_identities(user_id);

-- Backfill существующих GitHub-юзеров. Subject = github_login (legacy formato);
-- новые регистрации Phase 2 будут писать numeric ID. Оба формата сосуществуют.
INSERT INTO user_identities (user_id, provider, subject, email)
SELECT id, 'github', github_login, email
FROM users
WHERE github_login IS NOT NULL AND github_login != ''
ON CONFLICT (provider, subject) DO NOTHING;
