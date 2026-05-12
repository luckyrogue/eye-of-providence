-- Migration 025: content_blocks — CMS-lite Phase 4 (Workstream 4).
--
-- Один row на пару (slug, locale). Slug — это namespace.section.field
-- (например `landing.hero.headline`), допустимый список захардкожен в
-- backend/internal/content/registry.go (Phase 1 allowlist = 8 slugs).
-- Locale — один из 'ru' | 'en' | 'kk' | 'es' (совпадает с mailer.Locale).
--
-- Колонки:
--   * content        — published JSONB; читается публичным GET /v1/content/:slug
--   * draft_content  — ad-hoc draft, видим только super_admin'у через preview
--   * published_at   — NULL пока draft не promote'нут; non-null = публичный read
--                      возвращает content (Scenario 2 cms-content-render.md)
--   * schema_version — для forward-compat: при чтении row backend сравнивает
--                      с registry; mismatch = X-Eop-Content-Source:
--                      schema_version_unsupported, frontend fallback на i18n
--   * updated_by     — кто последний раз сохранил (audit + UI)
--
-- PRIMARY KEY (slug, locale) — один row на пару. DELETE row = revert к
-- bundled i18n default (Scenario 6 admin-cms-publish.md).
--
-- idx_content_blocks_published — partial index для hot-path GET /v1/content/:slug
-- (only published rows), уменьшает size vs full PK scan.

CREATE TABLE IF NOT EXISTS content_blocks (
    slug            TEXT        NOT NULL,
    locale          TEXT        NOT NULL CHECK (locale IN ('ru', 'en', 'kk', 'es')),
    content         JSONB       NOT NULL DEFAULT '{}'::jsonb,
    schema_version  INT         NOT NULL DEFAULT 1,
    published_at    TIMESTAMPTZ,
    draft_content   JSONB,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_by      UUID        REFERENCES users(id) ON DELETE SET NULL,
    PRIMARY KEY (slug, locale)
);

CREATE INDEX IF NOT EXISTS idx_content_blocks_published
    ON content_blocks (slug, locale)
    WHERE published_at IS NOT NULL;
