#!/usr/bin/env bash
# One-shot setup: создать все Fly apps + volumes + secrets для Eye of Providence.
# Не идемпотентный — запускается один раз. Повторный запуск может зафейлиться
# на existing-resource ошибках, это нормально.
#
# Требует: fly CLI (https://fly.io/docs/hands-on/install-flyctl/), залогинен `fly auth login`.

set -euo pipefail

ORG="${FLY_ORG:-personal}"
REGION="${FLY_REGION:-ams}"
API_APP="eop-api"
CH_APP="eop-clickhouse"
DASH_APP="eop-dashboard"
PG_APP="eop-postgres"

echo "==> 1. Postgres (managed Fly Postgres)"
fly postgres create --name "${PG_APP}" --region "${REGION}" --org "${ORG}" \
  --initial-cluster-size 1 --vm-size shared-cpu-1x --volume-size 1 || true

echo "==> 2. ClickHouse app"
fly apps create "${CH_APP}" --org "${ORG}" || true
fly volumes create clickhouse_data --app "${CH_APP}" --region "${REGION}" --size 5 -y || true
read -rp "ClickHouse password (set CLICKHOUSE_PASSWORD): " -s CH_PASS
echo
fly secrets set --app "${CH_APP}" CLICKHOUSE_PASSWORD="${CH_PASS}"
(cd "$(dirname "${BASH_SOURCE[0]}")/clickhouse" && fly deploy --app "${CH_APP}")

echo "==> 3. API app"
fly apps create "${API_APP}" --org "${ORG}" || true
fly postgres attach "${PG_APP}" --app "${API_APP}" || true   # выставит DATABASE_URL secret

# Конвертируем DATABASE_URL → EOP_POSTGRES_DSN (Fly attach даёт postgres:// URL)
DATABASE_URL=$(fly secrets list --app "${API_APP}" --json | python3 -c "
import sys, json
secrets = json.load(sys.stdin)
for s in secrets:
    if s.get('Name') == 'DATABASE_URL':
        print(s.get('Value', ''))
        break
" 2>/dev/null || true)

read -rp "EOP_JWT_SECRET (random ≥32 chars): " -s JWT
echo
read -rp "EOP_GEMINI_API_KEY (skip with empty for mock-mode): " GEMINI
echo
read -rp "EOP_GITHUB_CLIENT_ID (skip empty): " GH_ID
read -rp "EOP_GITHUB_CLIENT_SECRET (skip empty): " -s GH_SEC
echo

fly secrets set --app "${API_APP}" \
  EOP_POSTGRES_DSN="postgres://eop:eop_dev@${PG_APP}.flycast:5432/eop?sslmode=disable" \
  EOP_CLICKHOUSE_DSN="clickhouse://eop:${CH_PASS}@${CH_APP}.flycast:9000/eop" \
  EOP_JWT_SECRET="${JWT}" \
  ${GEMINI:+EOP_GEMINI_API_KEY="${GEMINI}"} \
  ${GH_ID:+EOP_GITHUB_CLIENT_ID="${GH_ID}"} \
  ${GH_SEC:+EOP_GITHUB_CLIENT_SECRET="${GH_SEC}"}

echo
echo "==> 4. Apply migrations"
echo "Postgres:"
fly postgres connect --app "${PG_APP}" -d eop < ../../backend/migrations/001_init.up.sql || \
  echo "  (skip: применить вручную через 'fly postgres connect')"
echo "ClickHouse: применить через ssh:"
echo "  fly ssh console -a ${CH_APP} -C 'clickhouse-client --user eop --password ${CH_PASS} --database eop --multiquery' < ../../backend/migrations/clickhouse_001_init.sql"

echo
echo "==> 5. Deploy API"
(cd ../../backend && fly deploy --app "${API_APP}")

echo
echo "==> 6. Dashboard"
fly apps create "${DASH_APP}" --org "${ORG}" || true
echo "Dashboard build args VITE_BACKEND_URL=https://${API_APP}.fly.dev"
(cd ../.. && fly deploy -c dashboard/fly.toml -a "${DASH_APP}" \
  --build-arg VITE_BACKEND_URL="https://${API_APP}.fly.dev")

echo
echo "✅ Done!"
echo "  API:       https://${API_APP}.fly.dev"
echo "  Dashboard: https://${DASH_APP}.fly.dev"
echo "  Healthz:   https://${API_APP}.fly.dev/healthz"
echo "  Metrics:   https://${API_APP}.fly.dev/metrics"
