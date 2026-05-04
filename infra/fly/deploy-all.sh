#!/usr/bin/env bash
# Деплой и backend и dashboard одной командой.

set -euo pipefail
HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "============== API =============="
"${HERE}/deploy-api.sh"

echo
echo "============== DASHBOARD =============="
"${HERE}/deploy-dashboard.sh"
