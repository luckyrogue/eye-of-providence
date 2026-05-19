# Disaster Recovery Runbook

**Status:** v0.1.x alpha · last updated 2026-05-19
**Audience:** Operators / on-call engineer
**Scope:** Single-region Dokploy deployment of Eye of Providence

---

## RTO / RPO targets (alpha)

| Tier | RTO | RPO |
| --- | --- | --- |
| API + dashboard | 30 min | n/a (stateless) |
| Postgres (auth, teams, payments, audit) | 1 h | 24 h (daily dump) |
| ClickHouse (events) | 4 h | 24 h (snapshot) |
| Redis (cache + WebAuthn) | 5 min | n/a (cache rebuildable; required at runtime) |

These are **alpha** targets. For GA contractual SLAs we tighten RPO Postgres to 1h via WAL streaming.

---

## What gets backed up

1. **Postgres** — `pg_dump --format=custom` daily at 03:00 UTC. Includes:
   `users`, `teams`, `team_members`, `team_payments`, `api_tokens`,
   `webhooks`, `sso_configs`, `audit_log`, `pairing_codes`, etc.
   Retention: 30 days local, 90 days S3.
2. **ClickHouse** — daily volume snapshot via Dokploy (or `clickhouse-backup`
   to S3). Single table `events` + materialized views.
3. **Container images** — GHCR `ghcr.io/luckyrogue/eop:*`
   (immutable per-deploy tag + `:main` floating).
4. **Configuration** — env vars in Dokploy UI; export via Dokploy CLI weekly
   to git-encrypted repo (NOT to public eye-of-providence repo).
5. **Migrations** — already in image (`backend/internal/migrate/sql/postgres/*.up.sql`).

What is **NOT** backed up:
- Redis data — cache rebuildable on cold start; WebAuthn pending challenges are ephemeral
- Pairing codes — ephemeral (10 min TTL)
- Push subscriptions — users can re-subscribe (UA-bound)

---

## Backup procedure (operator runs this)

### Postgres dump (daily cron)

```bash
docker exec eop-postgres pg_dump -U eop -d eop \
  --format=custom --no-owner --no-privileges \
  --file=/backups/eop-pg-$(date +%Y%m%d).dump

# Upload to S3 (Dokploy backup feature handles this if enabled).
aws s3 cp /backups/eop-pg-$(date +%Y%m%d).dump \
  s3://eop-backups/postgres/ --storage-class STANDARD_IA
```

Verify dump validity:

```bash
docker run --rm -v $(pwd)/backups:/data postgres:16-alpine \
  pg_restore --list /data/eop-pg-YYYYMMDD.dump | head
```

### ClickHouse snapshot

```bash
# Option A: clickhouse-backup (preferred — handles parts/metadata correctly)
docker exec eop-clickhouse clickhouse-backup create eop-$(date +%Y%m%d)
docker exec eop-clickhouse clickhouse-backup upload eop-$(date +%Y%m%d)

# Option B: volume snapshot (Dokploy schedule → daily)
# Use only when CH is idle (≤1 INSERT/sec). Otherwise data files inconsistent.
```

---

## Restore procedure

### Postgres — full restore from dump

```bash
# 1. Stop API to prevent concurrent writes during restore.
# Service name in compose is `eop`; container_name is `eop-app` (for docker exec).
docker compose -f infra/docker-compose.full.yml stop eop

# 2. Drop target DB (DESTRUCTIVE — only on confirmed loss).
docker exec eop-postgres psql -U eop -d postgres -c "DROP DATABASE IF EXISTS eop;"
docker exec eop-postgres psql -U eop -d postgres -c "CREATE DATABASE eop OWNER eop;"

# 3. Restore.
docker exec -i eop-postgres pg_restore -U eop -d eop --no-owner \
  < /backups/eop-pg-YYYYMMDD.dump

# 4. Verify row counts:
docker exec eop-postgres psql -U eop -d eop -c \
  "SELECT 'users' AS table, count(*) FROM users
   UNION ALL SELECT 'teams', count(*) FROM teams
   UNION ALL SELECT 'team_payments', count(*) FROM team_payments
   UNION ALL SELECT 'audit_log', count(*) FROM audit_log;"

# 5. Start API; миграции автоприменятся если backup был старше последней miграции.
docker compose -f infra/docker-compose.full.yml start eop

# 6. Check /healthz returns 200.
curl -fsS http://localhost:8080/healthz
```

### Postgres — point-in-time restore (PITR)

PITR requires WAL archiving (not configured by default). For alpha, fall back to
"daily dump + accept 24h loss". For GA: configure `archive_command` to S3 +
restore WAL replay.

### ClickHouse — restore from backup

```bash
docker exec eop-clickhouse clickhouse-backup download eop-YYYYMMDD
docker exec eop-clickhouse clickhouse-backup restore eop-YYYYMMDD
```

### Full DR (region failure)

1. **Provision new host** in alternative region (Dokploy on new VM).
2. **Pull latest image:** `docker pull ghcr.io/luckyrogue/eop:main`.
3. **Restore env vars** from git-encrypted backup.
4. **Restore Postgres + ClickHouse** dumps from S3.
5. **Update DNS** (`eop.rysdavletov.org` A record → new host IP).
6. **Wait for cert** (Let's Encrypt provisioning ~2 min).
7. **Smoke test:** `/healthz`, login flow, dashboard renders.

Expected total RTO: 45–60 min from declaration.

---

## Failure scenarios + playbook

### 1. Postgres data corruption (pg_isready healthy but queries return errors)

```bash
# Quick check:
docker exec eop-postgres psql -U eop -d eop -c "SELECT 1;"
docker logs eop-postgres --tail 100

# If corruption suspected → restore from last clean dump.
# Last-known-good: check `audit_log` last entry's created_at;
# if it's older than backup, you have a recent dump.
```

### 2. ClickHouse OOM kill (container exits with 137)

```bash
docker stats --no-stream | grep clickhouse
# If memory > 4G limit — bump deploy.resources.limits.memory and restart.
# Otherwise check parts merge backlog:
docker exec eop-clickhouse clickhouse-client \
  --query "SELECT count() FROM system.merges WHERE is_done=0"
# >20 = stuck; force OPTIMIZE TABLE events.
```

### 3. JWT secret leaked

1. **Generate new secret:** `openssl rand -hex 32`.
2. **Update `EOP_JWT_SECRET`** in Dokploy UI.
3. **Restart API** — все existing JWT инвалидируются (signature не сходится).
4. **Bump token_version** для всех super_admin'ов опционально:
   `UPDATE users SET token_version = token_version + 1 WHERE global_role='super_admin';`
5. **Audit:** проверить `audit_log` на подозрительные действия за последние 24h.

### 4. Single super_admin lost access

```sql
-- В Postgres:
UPDATE users SET global_role = 'super_admin' WHERE email = 'rescue@yourcompany.com';
UPDATE users SET token_version = token_version + 1 WHERE email = 'rescue@yourcompany.com';
```

Защита от accidentially deleting last super_admin уже в коде
(`teams/admin.go:166` — `last_super_admin` conflict).

### 5. Migration deployed broken schema (forward-only)

```bash
# Список применённых migrations:
docker exec eop-postgres psql -U eop -d eop -c \
  "SELECT * FROM schema_migrations ORDER BY version DESC LIMIT 5;"

# Откатить на N версий:
docker exec eop-api /usr/local/bin/migrate \
  -path /usr/local/share/migrate/postgres \
  -database "$EOP_POSTGRES_DSN" \
  down 1

# Если migration сломала dirty state:
docker exec eop-api /usr/local/bin/migrate \
  -path /usr/local/share/migrate/postgres \
  -database "$EOP_POSTGRES_DSN" \
  force <version>
```

---

## Quarterly DR test (mandatory)

Раз в квартал operator выполняет:

1. **Postgres dump → новая VM → restore** — измерить RTO end-to-end.
2. **Verify counts** = production counts (±1 для race-during-dump).
3. **Smoke test** restored instance с production frontend pointed at new backend.
4. **Document anomalies** в `docs/dr-test-log.md` (отдельный файл).

Если RTO > 1h или counts mismatch — escalate as P1.

---

## Contacts

- **On-call rotation:** see `#oncall` Slack channel
- **Cloud account access:** Vault `secret/eop/prod-aws` + `secret/eop/dokploy`
- **Domain control:** Cloudflare account, `eop.rysdavletov.org` zone
- **Image registry:** GitHub → Settings → Packages → `eop`

---

## Open follow-ups (track in project board)

- [ ] Set up WAL streaming for Postgres (PITR, RPO < 5min)
- [ ] Automate ClickHouse incremental backups (currently full snapshot only)
- [ ] Cross-region replica (currently single point of failure if Dokploy host dies)
- [ ] Quarterly DR test schedule + log
- [ ] Document Redis flush procedure (cache poisoning recovery)
