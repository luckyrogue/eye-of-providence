-- Reverse 001_init.up.sql.
-- Drop в порядке, обратном FK-зависимостям. pgcrypto оставляем на месте —
-- extension может использоваться другими БД на том же сервере.

DROP INDEX IF EXISTS idx_reports_user_period;
DROP TABLE IF EXISTS reports;
DROP TABLE IF EXISTS api_tokens;
DROP TABLE IF EXISTS consent;
DROP TABLE IF EXISTS projects;
DROP TABLE IF EXISTS devices;
DROP TABLE IF EXISTS teams;
DROP TABLE IF EXISTS users;
