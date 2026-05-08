# Audit findings — 2026-05-08

Сведённые отчёты от code-reviewer + security audit + QA. Файлы и строки на момент аудита.

---

## P0 — must fix перед onboarding'ом первой компании

| # | Источник | Файл | Проблема | Fix |
|---|---|---|---|---|
| 1 | Sec | `auth/handler.go:14`, `teams/handler.go:78` | JWT TTL = 90 дней, нет revocation, при `handleAdminUpdateUser` демоут не инвалидирует токен | TTL → 7-14 дней + `users.token_version` чек в Middleware; bump на демоут/удаление/wipe |
| 2 | Sec | `teams/handler.go:671/686` | `findInvite` + `consumeInvite` non-atomic — два concurrent accept'а с `use_count = max-1` оба проходят | Single `UPDATE ... WHERE use_count < max_uses RETURNING team_id` в одной tx |
| 3 | Code+Sec | `teams/handler.go:329` | TOCTOU на `handleCreateTeam` — concurrent creators при beta=3 могут создать 4 команды | `pg_advisory_xact_lock` или partial unique index + retry на конфликт |
| 4 | Code | `handler.go:936, 1121` | Promote-to-owner не проверяет, что target уже не owner где-то ещё → ломает 1-owner invariant | Перед `UPDATE role='owner'`: `count(*) FROM team_members WHERE user_id=$1 AND role='owner' AND team_id<>$2` |
| 5 | Code | `handler.go:1121 + addMember` | `handleAdminAddMember` с ролью owner для существующего member молча no-op'ит (`ON CONFLICT DO NOTHING`) | `ON CONFLICT (team_id, user_id) DO UPDATE SET role=EXCLUDED.role` |
| 6 | Code | `handler.go:923, 1054` | Cascade-delete не трогает ClickHouse — orphaned events после удаления юзера/команды | На `handleAdminDeleteUser` + `handleDeleteTeam`: вызвать `EventStore.DeleteUserData` для каждого затронутого юзера |
| 7 | Sec | `cmd/api/main.go:124` | Rate limit только на /auth/* — `POST /v1/teams`, `POST /v1/commits`, `POST /v1/ingest`, `/v1/admin/*` без лимита | Глобальный authed-limiter (120/min/user), strict на `/v1/admin/*` |
| 8 | Sec | `auth/github.go`, `auth/handler.go:67` | OAuth не проверяет verified email; на ошибке Upsert (collision) JWT всё равно выдаётся → orphan token | Использовать `/user/emails` с `verified:true`; на `Upsert` error возвращать 500, не выдавать токен |
| 9 | Sec | `config.go` | `EnableDevToken = (env != "production")`, `JWTSecret = default` — self-hoster без `EOP_ENV=production` хостит публично с открытым `/dev-token` | Hard-fail запуск если default JWT secret И не explicit `EOP_ENV=development` |

## P1 — high impact, easy

| # | Источник | Файл | Проблема | Fix |
|---|---|---|---|---|
| 10 | Code | `App.tsx:66 logout()` | Не чистит `eop_team`, `eop_display_name` — утечка между юзерами при смене аккаунта | Чистить все ключи `eop_*` в logout |
| 11 | Code | `Admin.tsx:148` | React fragment без `key` внутри `teams.map` — реконсиляция attach'ит state input'ов не к той строке | `<React.Fragment key={t.id}>` |
| 12 | Sec | `cmd/api/main.go:94`, `config.go:69` | `EOP_ALLOWED_ORIGINS=https://*.foo.com` молча сломан (Fiber только exact match); `AllowMethods` без PATCH | Reject wildcard в Validate; добавить PATCH в AllowMethods |
| 13 | Code | `handler.go:1192-1234 handleSetSubscription` | Validation после `tx.Begin` — idle txn slots на каждом 400 (DoS amp) | Валидация до `Begin` |
| 14 | Code | `handler.go:558 handleIngestCommit` | Бранч `if teamID != nil` skip'ает membership check, когда `projects.team_id IS NULL` (после удалённой команды) | Reject если `teamID == nil` (или upgrade FK на CASCADE) |
| 15 | Sec | `handler.go:609 handleProjectCommits` | `WHERE project_id AND team_id` для cross-team возвращает 0 строк, но не 404 — отличает "пусто" от "не существует" | Сначала `SELECT 1 FROM projects WHERE id=$pid AND team_id=$tid`, иначе 404 |

## P2 — medium

| # | Источник | Файл | Проблема | Fix |
|---|---|---|---|---|
| 16 | Code | `Admin.tsx:293, 303` | `until + "T23:59:59Z"` — UTC end-of-day, не локальный (off-by-one в +tz) | Вычислять конец дня в tz пользователя |
| 17 | Code | `Admin.tsx:296` + `handler.go:1232` | `amount=0` (или пустая строка → NaN) silently пропускается, plan меняется без записи платежа | Reject 400 на server при `payment != nil && amount <= 0`; client disable Save до валидного amount |
| 18 | Code | `handler.go:1054 handleAdminDeleteUser` | super_admin может удалить других super_admin'ов (только self заблокирован) | Refuse если хотя бы 1 super_admin останется |
| 19 | Sec | `handler.go:990 handleRemoveMember` (и `handleUpdateMemberRole`) | "Last owner" check в отдельном SELECT — два owner'а одновременно демоутящие друг друга могут оставить team без owner | Single statement: `WHERE (SELECT count(*) FROM team_members WHERE team_id=$1 AND role='owner') > 1` |
| 20 | Code | `auth/me.go DELETE /v1/me/data` | Если юзер был sole owner team — после wipe команда висит без owner | Перед wipe: или auto-delete команды где ты sole owner, или 409 "transfer ownership first" |
| 21 | Code | `api.ts contracts` | `TeamMember.created_at` → должно быть `joined_at`; `adminSetSubscription.until: null` semantics — `null` ≠ clear (нужна `""`) | Поправить типы; на бэке трактовать `null` как clear (или явно документировать) |
| 22 | Code | `api.ts:222` | `AdminUser.teams_count?: number` — backend всегда возвращает int, optional маскирует баги | Сделать обязательным |
| 23 | Sec | `auth/middleware.go:19` | Cas-sensitive `Bearer ` префикс (`bearer ` фейлит) | `EqualFold("bearer ", ...)` |
| 24 | Sec | `requireSuperAdmin` | Делает `SELECT global_role` per request | OK как есть (нет JWT-кэша → нет stale role); просто оставить |

## Не баги, для информации

- `handleAdminListAllTeams:746` использует `isSuperAdmin` напрямую вместо `requireSuperAdmin` — same effect, inconsistency.
- 90-дневный JWT + super_admin token → длинный blast radius при компрометации.
- Multiple owners в одной команде технически возможны схемой и не запрещены (`PRIMARY KEY (team_id, user_id)`). Это feature, но в комбинации с invariant "1-owner-per-user" даёт unexpected состояния.
- `subscription_until` в прошлом не trigger'ит downgrade — нет крона. Сейчас не критично (нет gating'а по плану), но flag когда добавим.

---

# QA strategy (Q1, target 40% coverage)

## Test stack
- **Backend integration tests** — HTTP-level против реального Postgres из dev-compose (отдельная БД `eop_test`), `TRUNCATE` в `t.Cleanup`
- **Backend unit tests** — `auth/jwt`, `auth/password`, `teams/validators`, `config` (cheap LOC)
- **Frontend** — Playwright happy-paths (8 сценариев), без Vitest пока
- **No testify, no testcontainers, no pgxmock** — стиль уже есть в `config_test.go`

## Helpers (день 1, ~150 LOC)
```
setupTestDB(t) -> *pgxpool.Pool       // migrate + truncate + cleanup
newTestApp(t, pool) -> *fiber.App
truncateAll(t, pool)
createUser(t, pool, email, password) -> uuid.UUID
makeSuperAdmin(t, pool, userID)
loginAs(t, app, email, password) -> string  // returns JWT
createTeam(t, pool, ownerID, name) -> uuid.UUID
addMember(t, pool, teamID, userID, role)
do(t, app, method, path, token, body) -> (status, body)
mustJSON(t, raw, dst)
```

## Test cases (54 штуки расписаны в детальном QA-отчёте)

10 priority-H блоков:
- handleRegister (8 тестов): bootstrap, invite-only, валидации
- handleLogin (4): happy path, wrong password, no email leak
- handleCreateTeam (7): beta limit, 1-owner, super_admin bypass
- handleUpdateMemberRole (5): cannot demote last owner
- handleRemoveMember (5): admin can't remove owner, last owner protection
- handleSetSubscription (7): atomic update, rollback on bad payment ⭐ — самый ценный тест
- handleAdminDeleteUser (4): cannot delete self
- handleAdminUpdateUser (5): cannot demote self, mass-assignment refused
- auth.Middleware (5): missing/malformed/expired/valid JWT
- DELETE /v1/me/data (4): wipe + isolation between users

## Двухнедельный sprint (week 1: ~25%, week 2: ~40%)

**Week 1:** helpers + unit'ы (jwt/password/validators) → middleware → register/login → createTeam → updateMemberRole → removeMember.

**Week 2:** subscription → admin delete/update user → me/data wipe → Playwright (3 happy-paths) → CI workflow с coverage gate (35% → 40%) → fill gaps в read paths.

## Coverage math
~1300 LOC в teams/handler.go × 60% + auth/* × 70% + config × 80% ÷ 3300 LOC backend ≈ **39-42%**.
