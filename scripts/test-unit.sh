#!/usr/bin/env bash
# Format, lint, then run unit tests (excludes ./e2e/...).
# Usage: ./scripts/test-unit.sh [go test args...]
#   ./scripts/test-unit.sh
#   ./scripts/test-unit.sh -run Bootstrap -count=1 ./internal/adapter/postgres/
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib.sh
source "${SCRIPT_DIR}/lib.sh"

cd "${PROJECT_ROOT}"
project_quality
require_go

pkgs=()
while IFS= read -r pkg; do
  pkgs+=("${pkg}")
done < <(unit_packages)

if (($# == 0)); then
  log "test-unit (go test -race <unit packages>)"
  go test -race "${pkgs[@]}"
elif go_test_args_have_package "$@"; then
  log "test-unit (go test -race $*)"
  go test -race "$@"
else
  log "test-unit (go test -race $* <unit packages>)"
  go test -race "$@" "${pkgs[@]}"
fi
