#!/usr/bin/env bash
# Claude Code hooks installer для Eye of Providence.
# Phase 3: добавляет Stop / PostToolUse / UserPromptSubmit хуки в ~/.claude/settings.json.

set -euo pipefail

SETTINGS_FILE="${HOME}/.claude/settings.json"

if [[ ! -f "${SETTINGS_FILE}" ]]; then
  echo "error: ${SETTINGS_FILE} не найден. Запусти Claude Code хотя бы раз и попробуй снова." >&2
  exit 1
fi

# TODO Phase 3: jq merge с существующими hooks, шлём в http://127.0.0.1:7373/v1/event/cli.
echo "skeleton: install не реализован — будет в Phase 3"
