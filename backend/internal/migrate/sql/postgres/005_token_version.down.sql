-- Reverse 005_token_version.up.sql.
-- ВНИМАНИЕ: после down все выпущенные JWT не смогут проверить tv-claim,
-- middleware начнёт отказывать. Перед откатом — спустить sessions всем юзерам.

ALTER TABLE users DROP COLUMN IF EXISTS token_version;
