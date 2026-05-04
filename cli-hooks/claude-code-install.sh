#!/usr/bin/env bash
# Claude Code hooks installer для Eye of Providence.
# Регистрирует Stop / PostToolUse / UserPromptSubmit хуки в ~/.claude/settings.json,
# которые шлют только метаданные срабатывания (без content).
#
# Требования: jq, curl.
# Запускать: ./claude-code-install.sh [--dry-run]

set -euo pipefail

SETTINGS_FILE="${HOME}/.claude/settings.json"
HOOK_SCRIPT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/eop-claude-hook.sh"
DRY_RUN=0
[[ "${1:-}" == "--dry-run" ]] && DRY_RUN=1

if ! command -v jq >/dev/null 2>&1; then
  echo "error: jq не установлен (brew install jq)." >&2
  exit 1
fi

if [[ ! -f "${HOOK_SCRIPT}" ]]; then
  echo "error: hook script не найден: ${HOOK_SCRIPT}" >&2
  exit 1
fi
chmod +x "${HOOK_SCRIPT}"

if [[ ! -f "${SETTINGS_FILE}" ]]; then
  mkdir -p "$(dirname "${SETTINGS_FILE}")"
  echo "{}" > "${SETTINGS_FILE}"
fi

# Резервная копия
cp "${SETTINGS_FILE}" "${SETTINGS_FILE}.eop-backup-$(date +%s)"

build_hook() {
  local hook_name="$1"
  jq -n --arg cmd "EOP_HOOK=${hook_name} ${HOOK_SCRIPT}" '{
    matcher: "",
    hooks: [{type: "command", command: $cmd}]
  }'
}

STOP=$(build_hook Stop)
POST=$(build_hook PostToolUse)
SUBMIT=$(build_hook UserPromptSubmit)

NEW=$(jq \
  --argjson stop "${STOP}" \
  --argjson post "${POST}" \
  --argjson submit "${SUBMIT}" '
  .hooks = (.hooks // {}) |
  .hooks.Stop = ((.hooks.Stop // []) + [$stop]) |
  .hooks.PostToolUse = ((.hooks.PostToolUse // []) + [$post]) |
  .hooks.UserPromptSubmit = ((.hooks.UserPromptSubmit // []) + [$submit])
' "${SETTINGS_FILE}")

if [[ ${DRY_RUN} -eq 1 ]]; then
  echo "${NEW}"
  exit 0
fi

echo "${NEW}" > "${SETTINGS_FILE}"
echo "✓ hooks установлены в ${SETTINGS_FILE}"
echo
echo "Не забудь экспортировать в shell:"
echo "  export EOP_TOKEN=<твой dev token>"
echo "  export EOP_URL=http://localhost:8080"
