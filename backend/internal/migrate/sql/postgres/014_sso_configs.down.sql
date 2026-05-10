DROP INDEX IF EXISTS users_sso_subject_idx;
ALTER TABLE users DROP COLUMN IF EXISTS sso_subject;
ALTER TABLE users DROP COLUMN IF EXISTS sso_team_id;
DROP INDEX IF EXISTS sso_states_expires_idx;
DROP TABLE IF EXISTS sso_states;
DROP TABLE IF EXISTS sso_configs;
