-- Reverse 002_teams_auth.up.sql.

DROP INDEX IF EXISTS idx_team_invites_team;
DROP INDEX IF EXISTS idx_team_invites_code;
DROP TABLE IF EXISTS team_invites;

ALTER TABLE users DROP CONSTRAINT IF EXISTS users_team_id_fkey;
ALTER TABLE users DROP COLUMN IF EXISTS display_name;
ALTER TABLE users DROP COLUMN IF EXISTS password_hash;

-- Изначально у role не было default'а; возвращаем это состояние.
ALTER TABLE users ALTER COLUMN role DROP DEFAULT;
