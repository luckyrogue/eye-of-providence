ALTER TABLE teams
    DROP COLUMN IF EXISTS plan_limits_override,
    DROP COLUMN IF EXISTS flags;
