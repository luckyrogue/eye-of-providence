-- Down: возвращаем CHECK без observer. Внимание: если в БД уже есть
-- observer-роли, ALTER упадёт. Down-migration предназначен для свежеприменённых
-- миграций и тестовых сценариев — production rollback должен сначала
-- удалить observer-роли (UPDATE team_members SET role='member' WHERE role='observer').

ALTER TABLE team_members DROP CONSTRAINT IF EXISTS team_members_role_check;
ALTER TABLE team_members ADD CONSTRAINT team_members_role_check
    CHECK (role IN ('owner', 'admin', 'member'));
