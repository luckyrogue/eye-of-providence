#!/usr/bin/env bash
# Postgres daily backup для self-hosted PG на VM.
#
# Установка (на production VM):
#   sudo cp infra/backup-postgres.sh /usr/local/bin/eop-backup-postgres
#   sudo chmod +x /usr/local/bin/eop-backup-postgres
#   sudo crontab -e
#     # Каждый день в 03:17 UTC (нечётное время чтоб не совпадать с другими cron'ами)
#     17 3 * * *  /usr/local/bin/eop-backup-postgres >> /var/log/eop-backup.log 2>&1
#
# Требует переменные окружения (например, в /etc/eop-backup.env):
#   EOP_PG_CONTAINER       — имя docker container'а с PG (e.g. eop-postgres)
#   EOP_PG_USER            — postgres user
#   EOP_PG_DB              — postgres db
#   EOP_BACKUP_DIR         — куда складывать (e.g. /var/backups/eop)
#   EOP_BACKUP_RETENTION   — keep last N days (default: 14)
#   EOP_BACKUP_REMOTE      — опционально, rclone destination (e.g. b2:eop-backups)
#                            если задано, дамп закидывается туда после локального сохранения.

set -euo pipefail

# Подгружаем env-файл если есть
if [ -f /etc/eop-backup.env ]; then
  # shellcheck disable=SC1091
  source /etc/eop-backup.env
fi

CONTAINER="${EOP_PG_CONTAINER:-eop-postgres}"
PG_USER="${EOP_PG_USER:-eop}"
PG_DB="${EOP_PG_DB:-eop}"
BACKUP_DIR="${EOP_BACKUP_DIR:-/var/backups/eop}"
RETENTION="${EOP_BACKUP_RETENTION:-14}"
REMOTE="${EOP_BACKUP_REMOTE:-}"

mkdir -p "$BACKUP_DIR"

ts="$(date -u +%Y%m%dT%H%M%SZ)"
out="$BACKUP_DIR/pg-${PG_DB}-${ts}.sql.gz"

echo "[$(date -uIs)] dumping $PG_DB → $out"
docker exec -i "$CONTAINER" pg_dump -U "$PG_USER" --clean --if-exists "$PG_DB" \
  | gzip -9 > "$out"

# Sanity-check — не нулевой байтсайз и валидный gzip.
test -s "$out"
gzip -t "$out"

size=$(du -h "$out" | cut -f1)
echo "[$(date -uIs)] dump ok ($size)"

# Прорежаем старые
find "$BACKUP_DIR" -name "pg-${PG_DB}-*.sql.gz" -mtime +"$RETENTION" -print -delete

# Offsite (опционально)
if [ -n "$REMOTE" ]; then
  echo "[$(date -uIs)] rclone copy → $REMOTE"
  rclone copy "$out" "$REMOTE/" --progress=false --no-traverse
fi

echo "[$(date -uIs)] done"
