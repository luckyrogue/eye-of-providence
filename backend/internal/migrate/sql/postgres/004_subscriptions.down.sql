-- Reverse 004_subscriptions.up.sql.

DROP INDEX IF EXISTS idx_team_payments_team;
DROP TABLE IF EXISTS team_payments;

ALTER TABLE teams DROP COLUMN IF EXISTS subscription_note;
ALTER TABLE teams DROP COLUMN IF EXISTS subscription_until;
ALTER TABLE teams DROP COLUMN IF EXISTS subscription_plan;
