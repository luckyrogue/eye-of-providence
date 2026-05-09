-- Reverse 003_multi_team_projects.up.sql.
-- ВНИМАНИЕ: down НЕ восстанавливает users.team_id (он удалён в M2M-схеме).
-- Если откатываемся к 002 — надо вручную перезаписать users.team_id из team_members
-- ДО запуска этого down (или потерять связь юзер ↔ команда).

DROP INDEX IF EXISTS idx_commits_project;
DROP INDEX IF EXISTS idx_commits_user;
DROP INDEX IF EXISTS idx_commits_team;
DROP TABLE IF EXISTS commits;

DROP INDEX IF EXISTS idx_projects_team;
ALTER TABLE projects DROP COLUMN IF EXISTS name;
ALTER TABLE projects DROP COLUMN IF EXISTS team_id;

ALTER TABLE teams DROP COLUMN IF EXISTS created_by;

DROP INDEX IF EXISTS idx_team_members_user;
DROP TABLE IF EXISTS team_members;

ALTER TABLE users DROP COLUMN IF EXISTS global_role;
