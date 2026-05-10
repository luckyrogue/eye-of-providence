DROP INDEX IF EXISTS idx_api_tokens_user_active;
ALTER TABLE api_tokens
  DROP COLUMN IF EXISTS revoked_at,
  DROP COLUMN IF EXISTS last_used_at,
  DROP COLUMN IF EXISTS prefix,
  DROP COLUMN IF EXISTS name;
