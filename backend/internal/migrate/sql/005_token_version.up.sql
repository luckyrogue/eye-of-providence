-- token_version — счётчик для invalidation выпущенных JWT.
-- Bump'ается при смене global_role (демоут super_admin'а), удалении юзера,
-- wipe данных. Middleware проверяет совпадение `tv` claim'а с этим полем
-- и отказывает 401, если не совпадает — это позволяет ребутнуть все сессии
-- юзера без ожидания истечения JWT.

ALTER TABLE users
  ADD COLUMN IF NOT EXISTS token_version INTEGER NOT NULL DEFAULT 0;
