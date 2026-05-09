DROP INDEX IF EXISTS idx_team_invites_email;
ALTER TABLE team_invites DROP COLUMN IF EXISTS sent_at;
ALTER TABLE team_invites DROP COLUMN IF EXISTS email;
