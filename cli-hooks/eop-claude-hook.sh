#!/usr/bin/env bash
# Универсальная hook-обёртка для Claude Code.
#
# Вызывается из ~/.claude/settings.json hooks. Получает JSON event payload
# на stdin (Claude Code feature), но мы используем только тип хука + время —
# никогда не пересылаем prompt/response контент в backend.
#
# Env:
#   EOP_TOKEN   — Bearer токен (получи через `eop dev-token` или dashboard)
#   EOP_URL     — backend URL (default http://localhost:8080)
#   EOP_HOOK    — тип хука (Stop | PostToolUse | UserPromptSubmit)

set -euo pipefail

EOP_URL="${EOP_URL:-http://localhost:8080}"
EOP_HOOK="${EOP_HOOK:-unknown}"

if [[ -z "${EOP_TOKEN:-}" ]]; then
  exit 0
fi

# Малый payload без контента: фиксируем только факт срабатывания + длительность 0.
# Реальная длительность накопится из watcher loop в desktop agent.
PAYLOAD=$(cat <<JSON
{"events":[{
  "app_bundle":"claude-code",
  "category":"ai",
  "source":"cli",
  "ai_provider":"anthropic",
  "ai_channel":"agent",
  "duration_ms":0,
  "lines_added":0,
  "lines_removed":0,
  "chars_in":0
}]}
JSON
)

curl -sS -X POST "${EOP_URL}/v1/ingest" \
  -H "Authorization: Bearer ${EOP_TOKEN}" \
  -H "Content-Type: application/json" \
  -d "${PAYLOAD}" >/dev/null 2>&1 || true

# Не блокируем Claude Code — всегда exit 0.
# stdin (event JSON) намеренно НЕ читаем, чтобы не было соблазна логировать контент.
exit 0
