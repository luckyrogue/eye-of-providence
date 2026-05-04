#!/usr/bin/env bash
# One-shot setup для Eye of Providence на Fly.io. Cheap-mode (single-node Postgres).
#
# CH_PASS=... PG_PASS=... JWT=... [GEMINI=...] [GH_ID=...] [GH_SEC=...] ./setup.sh
#
# Стоимость после $5 Hobby credit:
#   eop-postgres single-node 256MB + 1GB volume      ≈ $2/мес
#   eop-clickhouse 1GB + 5GB volume                  ≈ $5/мес
#   eop-api 256MB auto-stop                          ≈ $1/мес
#   eop-dashboard 256MB auto-stop                    ≈ $0.50/мес
#   ----------------------------------
#   Итого    ≈ $8.50/мес минус $5 credit = ≈ $3.50/мес

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

[[ -z "${PG_PASS:-}" ]] && PG_PASS=$(openssl rand -hex 16)
[[ -z "${CH_PASS:-}" ]] && CH_PASS=$(openssl rand -hex 16)
[[ -z "${JWT:-}" ]] && JWT=$(openssl rand -hex 32)
GEMINI="${GEMINI:-}"
GH_ID="${GH_ID:-}"
GH_SEC="${GH_SEC:-}"

echo "✓ secrets ready"
echo "  PG_PASS=${PG_PASS}"
echo "  CH_PASS=${CH_PASS}"
echo "  JWT=${JWT:0:8}…"
echo

app_exists() {
  fly apps list --org "${ORG}" 2>/dev/null | awk '{print $1}' | grep -qx "$1"
}
volume_exists() {
  fly volumes list --app "$1" 2>/dev/null | grep -q "$2"
}

# --- 1. Postgres (single-node) ---
echo "==> 1. Postgres (single-node, не managed — cheap mode)"
if app_exists "${PG_APP}"; then
  echo "    ${PG_APP} exists."
else
  fly apps create "${PG_APP}" --org "${ORG}"
fi
if volume_exists "${PG_APP}" "postgres_data"; then
  echo "    volume postgres_data exists."
else
  fly volumes create postgres_data --app "${PG_APP}" --region "${REGION}" --size 1 -y
fi
fly secrets set --app "${PG_APP}" POSTGRES_PASSWORD="${PG_PASS}" --stage
(cd "${ROOT}/infra/fly/postgres" && fly deploy --app "${PG_APP}")

# --- 2. ClickHouse ---
echo
echo "==> 2. ClickHouse"
if app_exists "${CH_APP}"; then
  echo "    ${CH_APP} exists."
else
  fly apps create "${CH_APP}" --org "${ORG}"
fi
if volume_exists "${CH_APP}" "clickhouse_data"; then
  echo "    volume clickhouse_data exists."
else
  fly volumes create clickhouse_data --app "${CH_APP}" --region "${REGION}" --size 5 -y
fi
fly secrets set --app "${CH_APP}" CLICKHOUSE_PASSWORD="${CH_PASS}" --stage
(cd "${ROOT}/infra/fly/clickhouse" && fly deploy --app "${CH_APP}")

# --- 3. API ---
echo
echo "==> 3. API"
if app_exists "${API_APP}"; then
  echo "    ${API_APP} exists."
else
  fly apps create "${API_APP}" --org "${ORG}"
fi
fly secrets set --app "${API_APP}" \
  EOP_POSTGRES_DSN="postgres://eop:${PG_PASS}@${PG_APP}.flycast:5432/eop?sslmode=disable" \
  EOP_CLICKHOUSE_DSN="clickhouse://eop:${CH_PASS}@${CH_APP}.flycast:9000/eop" \
  EOP_JWT_SECRET="${JWT}" \
  ${GEMINI:+EOP_GEMINI_API_KEY="${GEMINI}"} \
  ${GH_ID:+EOP_GITHUB_CLIENT_ID="${GH_ID}"} \
  ${GH_SEC:+EOP_GITHUB_CLIENT_SECRET="${GH_SEC}"} \
  --stage
(cd "${ROOT}/backend" && fly deploy --app "${API_APP}")

# --- 4. Migrations ---
echo
echo "==> 4. Apply migrations"
echo "    Postgres:"
fly ssh console -a "${PG_APP}" -C "psql -U eop -d eop" \
  < "${ROOT}/backend/migrations/001_init.up.sql" || \
  echo "    (try manually: fly ssh console -a ${PG_APP} -C 'psql -U eop -d eop' < ${ROOT}/backend/migrations/001_init.up.sql)"
echo "    ClickHouse:"
fly ssh console -a "${CH_APP}" -C "clickhouse-client --user eop --password ${CH_PASS} --database eop --multiquery" \
  < "${ROOT}/backend/migrations/clickhouse_001_init.sql" || \
  echo "    (run manually)"

# --- 5. Dashboard ---
echo
echo "==> 5. Dashboard"
if app_exists "${DASH_APP}"; then
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
echo
echo "Secrets (СОХРАНИ):"
echo "  PG_PASS=${PG_PASS}"
echo "  CH_PASS=${CH_PASS}"
echo "  JWT=${JWT}"
