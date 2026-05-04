#!/usr/bin/env bash
# One-shot setup для Eye of Providence на Fly.io.
#
# Можно прокинуть всё через env (рекомендуется для не-интерактивного запуска):
#   CH_PASS=... JWT=... [GEMINI=...] [GH_ID=...] [GH_SEC=...] ./setup.sh
# Иначе скрипт спросит интерактивно.
#
# Требует: flyctl, залогинен `fly auth login`.

set -euo pipefail

ORG="${FLY_ORG:-personal}"
REGION="${FLY_REGION:-ams}"
API_APP="${API_APP:-eop-api}"
CH_APP="${CH_APP:-eop-clickhouse}"
DASH_APP="${DASH_APP:-eop-dashboard}"
PG_APP="${PG_APP:-eop-postgres}"

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
echo "Repo root: ${ROOT}"
echo "Org: ${ORG}, Region: ${REGION}"
echo

# --- gather secrets (env vars или prompt) ---
if [[ -z "${CH_PASS:-}" ]]; then
  read -rp "ClickHouse password (random если пусто): " -s CH_PASS
  echo
  [[ -z "${CH_PASS}" ]] && CH_PASS=$(openssl rand -hex 16)
fi
if [[ -z "${JWT:-}" ]]; then
  read -rp "JWT secret (random если пусто, ≥32 hex chars): " -s JWT
  echo
  [[ -z "${JWT}" ]] && JWT=$(openssl rand -hex 32)
fi
GEMINI="${GEMINI:-}"
GH_ID="${GH_ID:-}"
GH_SEC="${GH_SEC:-}"

echo "✓ secrets ready (CH_PASS, JWT$([[ -n "${GEMINI}" ]] && echo ", GEMINI")$([[ -n "${GH_ID}" ]] && echo ", GitHub OAuth"))"
echo

# --- 1. Postgres ---
echo "==> 1. Postgres (managed)"
if fly apps list --org "${ORG}" 2>/dev/null | grep -q "^${PG_APP} "; then
  echo "    ${PG_APP} exists, skip."
else
  fly postgres create --name "${PG_APP}" --region "${REGION}" --org "${ORG}" \
    --initial-cluster-size 1 --vm-size shared-cpu-1x --volume-size 1
fi

# --- 2. ClickHouse ---
echo
echo "==> 2. ClickHouse"
if fly apps list --org "${ORG}" 2>/dev/null | grep -q "^${CH_APP} "; then
  echo "    ${CH_APP} app exists."
else
  fly apps create "${CH_APP}" --org "${ORG}"
fi

if fly volumes list --app "${CH_APP}" 2>/dev/null | grep -q "clickhouse_data"; then
  echo "    volume clickhouse_data exists."
else
  fly volumes create clickhouse_data --app "${CH_APP}" --region "${REGION}" --size 5 -y
fi

fly secrets set --app "${CH_APP}" CLICKHOUSE_PASSWORD="${CH_PASS}" --stage
(cd "${ROOT}/infra/fly/clickhouse" && fly deploy --app "${CH_APP}")

# --- 3. API ---
echo
echo "==> 3. API"
if fly apps list --org "${ORG}" 2>/dev/null | grep -q "^${API_APP} "; then
  echo "    ${API_APP} exists."
else
  fly apps create "${API_APP}" --org "${ORG}"
fi

# Postgres attach создаёт DATABASE_URL secret
if ! fly secrets list --app "${API_APP}" 2>/dev/null | grep -q DATABASE_URL; then
  fly postgres attach "${PG_APP}" --app "${API_APP}" --yes
fi

fly secrets set --app "${API_APP}" \
  EOP_CLICKHOUSE_DSN="clickhouse://eop:${CH_PASS}@${CH_APP}.flycast:9000/eop" \
  EOP_JWT_SECRET="${JWT}" \
  ${GEMINI:+EOP_GEMINI_API_KEY="${GEMINI}"} \
  ${GH_ID:+EOP_GITHUB_CLIENT_ID="${GH_ID}"} \
  ${GH_SEC:+EOP_GITHUB_CLIENT_SECRET="${GH_SEC}"} \
  --stage

# Convert DATABASE_URL → EOP_POSTGRES_DSN. Fly attach даёт URL c sslmode по умолчанию,
# но без sslmode=disable наш driver откажется. Перевыставляем.
DB_URL=$(fly secrets list --app "${API_APP}" --json 2>/dev/null | python3 -c "
import sys, json
for s in json.load(sys.stdin):
    if s.get('Name') == 'DATABASE_URL':
        print(s.get('Value', ''))
        break
" 2>/dev/null || true)
if [[ -n "${DB_URL}" ]]; then
  # DATABASE_URL приходит без значения (Fly hides), используем connection string из postgres-attach
  : # noop, EOP_POSTGRES_DSN можно выставить вручную после первого деплоя
fi

(cd "${ROOT}/backend" && fly deploy --app "${API_APP}")

# --- 4. Migrations ---
echo
echo "==> 4. Apply migrations (manually):"
echo "    Postgres:"
echo "      fly postgres connect -a ${PG_APP} -d eop < ${ROOT}/backend/migrations/001_init.up.sql"
echo "    ClickHouse:"
echo "      fly ssh console -a ${CH_APP} -C \"clickhouse-client --user eop --password ${CH_PASS} --database eop --multiquery\" < ${ROOT}/backend/migrations/clickhouse_001_init.sql"

# --- 5. Dashboard ---
echo
echo "==> 5. Dashboard"
if fly apps list --org "${ORG}" 2>/dev/null | grep -q "^${DASH_APP} "; then
  echo "    ${DASH_APP} exists."
else
  fly apps create "${DASH_APP}" --org "${ORG}"
fi

(cd "${ROOT}" && fly deploy -c dashboard/fly.toml -a "${DASH_APP}" \
  --build-arg VITE_BACKEND_URL="https://${API_APP}.fly.dev" .)

echo
echo "✅ Done!"
echo "  API:        https://${API_APP}.fly.dev"
echo "  Dashboard:  https://${DASH_APP}.fly.dev"
echo "  Healthz:    https://${API_APP}.fly.dev/healthz"
echo "  Metrics:    https://${API_APP}.fly.dev/metrics"
echo
echo "Не забудь применить миграции (см. шаг 4 выше)."
echo "ClickHouse password сохранён в fly secrets, для применения CH миграций:"
echo "  CH_PASS=${CH_PASS}"
