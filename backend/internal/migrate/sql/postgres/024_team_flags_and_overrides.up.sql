-- Migration 024: per-team feature flags + plan-limit overrides. Phase 3
-- admin workstream.
--
-- Дизайн:
--   * teams.flags JSONB — boolean / numeric overrides (см.
--     backend/internal/plans/flags.go allowlist). Shallow-merge через
--     PATCH /v1/admin/teams/:id/flags. DEFAULT '{}' гарантирует backward
--     compat — существующие команды видят плановые defaults.
--   * teams.plan_limits_override JSONB (nullable) — numeric overrides
--     mirror'ят plans.Limits ключи (см.
--     .team/product-acceptance/admin-plan-overrides.md). NULL = use plan
--     defaults; явный {} тоже фактически означает defaults но сохраняет
--     "override row exists" для audit trail.
--
-- 023 пропущена: custom roles cancelled. observer role была введена в 021
-- через ALTER CHECK constraint, без новой таблицы.
--
-- Не создаём отдельную таблицу для overrides ради простоты read-path: у
-- нас уже есть teams row, который читаем на каждом request'е (для
-- subscription_plan/until). Добавить два JSONB-колонки дешевле, чем JOIN.

ALTER TABLE teams
    ADD COLUMN IF NOT EXISTS flags JSONB NOT NULL DEFAULT '{}'::jsonb,
    ADD COLUMN IF NOT EXISTS plan_limits_override JSONB;
