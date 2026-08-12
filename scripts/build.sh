#!/usr/bin/env bash
# Format, lint, then build the oauth binary into bin/oauth.
# Usage: ./scripts/build.sh [go build args...]
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib.sh
source "${SCRIPT_DIR}/lib.sh"

cd "${PROJECT_ROOT}"
project_quality

mkdir -p "${BIN_DIR}"
out="${BIN_DIR}/${BIN_NAME}"

log "build (${CMD_PACKAGE} -> ${out})"
go build -o "${out}" "$@" "${CMD_PACKAGE}"
log "built ${out}"
