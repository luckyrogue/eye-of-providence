-- many-to-many users ↔ teams через team_members.
-- Глобальные роли через users.role (super_admin | user). Per-team роли в team_members.role.
-- + projects per team + commits.

-- 1. Глобальная роль на пользователя
ALTER TABLE users
  ADD COLUMN IF NOT EXISTS global_role TEXT NOT NULL DEFAULT 'user'
    CHECK (global_role IN ('super_admin', 'user'));

-- 2. Many-to-many users ↔ teams
CREATE TABLE IF NOT EXISTS team_members (
  team_id    UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
  user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  role       TEXT NOT NULL DEFAULT 'member'
             CHECK (role IN ('owner', 'admin', 'member')),
  joined_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (team_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_team_members_user ON team_members(user_id);

-- Бэкфилл из старой users.team_id если есть данные
INSERT INTO team_members (team_id, user_id, role, joined_at)
SELECT team_id, id, COALESCE(role, 'member'), created_at
FROM users
WHERE team_id IS NOT NULL
ON CONFLICT (team_id, user_id) DO NOTHING;

-- Команды теперь имеют owner_id (нужен для display) — вычисляется из team_members с role=owner.
-- Добавим denormalized created_by для удобства.
ALTER TABLE teams
  ADD COLUMN IF NOT EXISTS created_by UUID REFERENCES users(id) ON DELETE SET NULL;

-- 3. Projects per team
ALTER TABLE projects
  ADD COLUMN IF NOT EXISTS team_id UUID REFERENCES teams(id) ON DELETE SET NULL,
  ADD COLUMN IF NOT EXISTS name TEXT;

CREATE INDEX IF NOT EXISTS idx_projects_team ON projects(team_id);

-- 4. Commits — отдельная таблица, фиксируется git post-commit hook'ом или CLI.
CREATE TABLE IF NOT EXISTS commits (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  project_id      UUID REFERENCES projects(id) ON DELETE SET NULL,
  team_id         UUID REFERENCES teams(id) ON DELETE SET NULL,
  user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  sha             TEXT NOT NULL,
  message         TEXT,
  branch          TEXT,
  files_changed   INTEGER NOT NULL DEFAULT 0,
  lines_added     INTEGER NOT NULL DEFAULT 0,
  lines_removed   INTEGER NOT NULL DEFAULT 0,
  ai_lines_pct    INTEGER,  -- доля AI-кода в коммите (если attribution worker отметил)
  authored_at     TIMESTAMPTZ NOT NULL,
  ingested_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (project_id, sha)
);

CREATE INDEX IF NOT EXISTS idx_commits_team ON commits(team_id, authored_at DESC);
CREATE INDEX IF NOT EXISTS idx_commits_user ON commits(user_id, authored_at DESC);
CREATE INDEX IF NOT EXISTS idx_commits_project ON commits(project_id, authored_at DESC);
