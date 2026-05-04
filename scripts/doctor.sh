#!/usr/bin/env bash
# scripts/doctor.sh — проверяет всё что нужно для разработки и деплоя.
# Print status table, exit 0 если ok, 1 если что-то критичное отсутствует.

set -uo pipefail

GREEN=$'\e[32m'
RED=$'\e[31m'
YELLOW=$'\e[33m'
BOLD=$'\e[1m'
RST=$'\e[0m'

errors=0
warnings=0

check() {
  local name="$1" expect="$2" install_hint="$3"
  if eval "$expect" >/dev/null 2>&1; then
    printf "  ${GREEN}✓${RST} %-30s %s\n" "$name" "$($expect 2>/dev/null | head -1 | cut -c1-50)"
  else
    if [[ "${4:-required}" == "required" ]]; then
      printf "  ${RED}✗${RST} %-30s ${RED}missing${RST}\n      → %s\n" "$name" "$install_hint"
      errors=$((errors + 1))
    else
      printf "  ${YELLOW}!${RST} %-30s ${YELLOW}optional${RST}\n      → %s\n" "$name" "$install_hint"
      warnings=$((warnings + 1))
    fi
  fi
}

echo "${BOLD}Eye of Providence — environment doctor${RST}"
echo

echo "${BOLD}Required for backend:${RST}"
check "go ≥ 1.23"              "go version | grep -E 'go1.(2[3-9]|[3-9][0-9])'"  "brew install go (или https://go.dev/dl/)"
check "docker"                 "docker version --format '{{.Client.Version}}'"   "Docker Desktop: https://www.docker.com/products/docker-desktop"
check "docker daemon running"  "docker info"                                     "Запусти Docker Desktop"

echo
echo "${BOLD}Required for frontend (dashboard, agent UI, browser-extension):${RST}"
check "node ≥ 20"              "node --version | grep -E 'v(2[0-9]|[3-9][0-9])'"  "brew install node, или https://nodejs.org"
check "pnpm 9+"                "pnpm --version | grep -E '^(9|[1-9][0-9])'"       "corepack enable && corepack prepare pnpm@9 --activate"

echo
echo "${BOLD}Required for desktop agent build:${RST}"
check "rustc ≥ 1.78"           "rustc --version | grep -E '1\.(7[8-9]|[8-9][0-9])'" "brew install rustup-init && rustup-init -y" "optional"
check "cargo"                  "cargo --version"                                  "идёт вместе с rustc через rustup" "optional"

echo
echo "${BOLD}Optional — for deploy:${RST}"
check "fly (flyctl)"           "fly version"                                     "brew install flyctl, или https://fly.io/docs/hands-on/install-flyctl/" "optional"
check "git"                    "git --version"                                   "brew install git"

echo
echo "${BOLD}Optional — for Claude Code hooks installer:${RST}"
check "jq"                     "jq --version"                                    "brew install jq" "optional"

echo
if [[ $errors -gt 0 ]]; then
  echo "${RED}${BOLD}${errors} required dependency(ies) missing.${RST}"
  exit 1
elif [[ $warnings -gt 0 ]]; then
  echo "${YELLOW}OK with ${warnings} optional warning(s).${RST}"
  exit 0
else
  echo "${GREEN}${BOLD}✓ All checks passed.${RST}"
  exit 0
fi
