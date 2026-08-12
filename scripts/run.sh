#!/usr/bin/env bash
# Format, lint, then run the oauth CLI (args forwarded).
# Usage: ./scripts/run.sh [oauth args...]
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib.sh
source "${SCRIPT_DIR}/lib.sh"

cd "${PROJECT_ROOT}"
project_quality

log "run (${CMD_PACKAGE})"
exec go run "${CMD_PACKAGE}" "$@"
