#!/usr/bin/env bash
# Деплой dashboard (eop-dashboard) на Fly.io.
# Pull последнего кода + fly deploy + verify что новый дизайн в проде.

set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
APP="${DASH_APP:-eop-dashboard}"
API_APP="${API_APP:-eop-api}"
BACKEND_URL="${BACKEND_URL:-https://${API_APP}.fly.dev}"
LOG="/tmp/eop-deploy-dashboard.log"

cd "${ROOT}"
echo "==> git pull"
git pull --ff-only

echo
echo "==> fly deploy --app ${APP}"
echo "    VITE_BACKEND_URL=${BACKEND_URL}"
fly deploy -c dashboard/fly.toml -a "${APP}" \
  --build-arg "VITE_BACKEND_URL=${BACKEND_URL}" . 2>&1 | tee "${LOG}"

echo
echo "==> verify"
sleep 3
JS_BUNDLE=$(curl -fsS "https://${APP}.fly.dev/" | grep -oE 'assets/index-[A-Za-z0-9_-]+\.js' | head -1)
if [[ -n "${JS_BUNDLE}" ]]; then
  echo "    bundle: ${JS_BUNDLE}"
  RU_STRINGS=$(curl -fsS "https://${APP}.fly.dev/${JS_BUNDLE}" | grep -oE 'Доля AI|Тепловая карта|Часовой пояс' | sort -u || true)
  if [[ -n "${RU_STRINGS}" ]]; then
    echo "    ✓ новый RU-design в проде:"
    echo "${RU_STRINGS}" | sed 's/^/      /'
  else
    echo "    ⚠️  RU-строки не найдены в bundle — возможно старая версия"
  fi
fi

echo
echo "Открой: https://${APP}.fly.dev"
echo "Если в браузере старый дизайн — Cmd+Shift+R (hard refresh) или открой в incognito."
echo "log: ${LOG}"
