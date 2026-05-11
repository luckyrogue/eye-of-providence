-- Audit log — единая cobweb всех "consequential" действий super_admin'ов и
-- owner'ов. Используется для compliance (SOC2/GDPR ready), forensics при
-- incident'ах, и UI таба в /admin.
--
-- Schema:
--   actor_id/email — кто сделал. email денормализован (если user удалён,
--     audit-row сохраняет историю).
--   action — машинно-читаемый код (team.deleted, user.role_changed, ...).
--   target_type/id — что затронуло (team, user, sso_config, payment, ...).
--   metadata — произвольный JSON со специфичными деталями.
--
-- Retention: 1 год по умолчанию (DELETE FROM audit_log WHERE created_at <
--   now() - interval '1 year'). Можно настроить per-plan позже.

CREATE TABLE IF NOT EXISTS audit_log (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    actor_id     UUID,
    actor_email  TEXT,
    action       TEXT NOT NULL,
    target_type  TEXT,
    target_id    TEXT,
    metadata     JSONB,
    ip           TEXT,
    user_agent   TEXT,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_audit_log_created ON audit_log(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_log_actor ON audit_log(actor_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_log_action ON audit_log(action, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_log_target ON audit_log(target_type, target_id, created_at DESC);
