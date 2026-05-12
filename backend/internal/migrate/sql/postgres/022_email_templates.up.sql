-- Migration 022: таблица email_templates — overrides для transactional
-- email templates per (key, locale). Phase 3 admin workstream.
--
-- Design:
--   * key — машинно-читаемый идентификатор шаблона ('password_reset',
--     'team_invite', 'subscription_activated'). Не FK; список валидируется в
--     application layer (backend/internal/mailer.TemplateStore).
--   * locale — один из 'ru' | 'en' | 'kk' | 'es' (sync с mailer.Locale).
--   * subject + body_html + body_text — full content для render.
--   * updated_by — кто изменил (для audit + UI). FK с ON DELETE SET NULL —
--     если super_admin удалён, override остаётся (с потерянной авторской
--     меткой).
--
-- PRIMARY KEY (key, locale) — один override на пару. Отсутствие row =>
-- mailer fallback на embedded baseline в backend/internal/mailer/templates.go.
--
-- 023 пропущена по продуктовому решению (custom roles cancelled,
-- см. .team/product-decisions-confirmed.md §4).

CREATE TABLE IF NOT EXISTS email_templates (
    key         TEXT        NOT NULL,
    locale      TEXT        NOT NULL CHECK (locale IN ('ru', 'en', 'kk', 'es')),
    subject     TEXT        NOT NULL,
    body_html   TEXT        NOT NULL,
    body_text   TEXT        NOT NULL,
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_by  UUID        REFERENCES users(id) ON DELETE SET NULL,
    PRIMARY KEY (key, locale)
);
