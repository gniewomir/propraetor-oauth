#!/usr/bin/env bash
# Format, lint, then test the module (with race detector).
# Usage: ./scripts/test.sh [go test args...]
# Extra args are appended after the default package pattern, e.g.:
#   ./scripts/test.sh -count=1 ./internal/domain/...
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib.sh
source "${SCRIPT_DIR}/lib.sh"

cd "${PROJECT_ROOT}"
project_quality

if (($# == 0)); then
  log "test (go test -race ${MODULE_PACKAGES})"
  go test -race ${MODULE_PACKAGES}
else
  log "test (go test -race $*)"
  go test -race "$@"
fi
