DROP TABLE IF EXISTS pairing_codes;
DROP INDEX IF EXISTS idx_api_tokens_kind_user;
ALTER TABLE api_tokens DROP COLUMN IF EXISTS kind;
