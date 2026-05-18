#!/usr/bin/env bash
# Fails if any internal/*/domain package imports forbidden dependencies.
set -euo pipefail
root="$(cd "$(dirname "$0")/.." && pwd)"
forbidden='(github.com/gofiber/fiber|github.com/jackc/pgx|go.uber.org/zap)'
failed=0
while IFS= read -r -d '' f; do
  if grep -E "^[[:space:]]*\"${forbidden}" "$f" >/dev/null 2>&1; then
    echo "FORBIDDEN IMPORT in $f"
    grep -E "^[[:space:]]*\"${forbidden}" "$f" || true
    failed=1
  fi
done < <(find "$root/internal" -path '*/domain/*.go' -print0 2>/dev/null)
if [[ $failed -ne 0 ]]; then
  echo "domain import check failed"
  exit 1
fi
echo "domain import check OK"
