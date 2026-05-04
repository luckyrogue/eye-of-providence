#!/usr/bin/env bash
# Деплой backend (eop-api) на Fly.io.
# Pull последнего кода + fly deploy с логом.

set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
APP="${API_APP:-eop-api}"
LOG="/tmp/eop-deploy-api.log"

cd "${ROOT}"
echo "==> git pull"
git pull --ff-only

cd "${ROOT}/backend"
echo "==> fly deploy --app ${APP}"
fly deploy --app "${APP}" 2>&1 | tee "${LOG}"

echo
echo "==> verify"
sleep 2
curl -fsS "https://${APP}.fly.dev/healthz" && echo " ✓"

echo
echo "log: ${LOG}"
