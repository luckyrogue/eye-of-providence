-- Migration 021: добавить роль 'observer' в team_members.role CHECK.
--
-- observer = read-only участник команды. Mutating routes (create project,
-- ingest, invites, members management) гейтятся через requireRole('owner', 'admin')
-- или requireRole('owner', 'admin', 'member') — observer туда не попадает.
-- Read endpoints (team detail, members list, projects list, commits, summary)
-- допускают любого участника команды.
--
-- Никакой новой таблицы (team_custom_roles) — Phase 1 ограничивается одной
-- встроенной ролью. См. .team/product-decisions-confirmed.md §4.

ALTER TABLE team_members DROP CONSTRAINT IF EXISTS team_members_role_check;
ALTER TABLE team_members ADD CONSTRAINT team_members_role_check
    CHECK (role IN ('owner', 'admin', 'member', 'observer'));
