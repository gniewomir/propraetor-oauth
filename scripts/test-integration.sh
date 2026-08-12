#!/usr/bin/env bash
# Format, lint, ensure integration storage, run integration tests.
# Usage: ./scripts/test-integration.sh [go test args...]
#   ./scripts/test-integration.sh
#   ./scripts/test-integration.sh -run Ping -count=1
# Fail-closed if Podman/integration storage cannot be ensured (docs/agents/testing.md).
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib.sh
source "${SCRIPT_DIR}/lib.sh"

cd "${PROJECT_ROOT}"
project_quality
require_go

log "ensure integration test storage"
"${SCRIPT_DIR}/storage.sh" ensure --env integration

env_file="${PROJECT_ROOT}/.local/oauth-storage/integration.env"
[[ -f "${env_file}" ]] || die "missing ${env_file} after storage ensure"
set -a
# shellcheck disable=SC1090
source "${env_file}"
set +a
[[ -n "${OAUTH_STORAGE_URL:-}" ]] || die "OAUTH_STORAGE_URL not set after sourcing ${env_file}"
[[ "${OAUTH_STORAGE_ENV:-}" == "integration" ]] || die "OAUTH_STORAGE_ENV must be integration (got ${OAUTH_STORAGE_ENV:-})"

# ADR-0073 suite bootstrap: ensure → truncate → migrate-to-head.
# migrate is not implemented yet; truncate is a no-op until tables exist.
log "truncate integration test storage (suite bootstrap)"
"${SCRIPT_DIR}/storage.sh" truncate --env integration

if (($# == 0)); then
  log "test-integration (go test -race ${INTEGRATION_PACKAGES})"
  go test -race "${INTEGRATION_PACKAGES}"
elif go_test_args_have_package "$@"; then
  log "test-integration (go test -race $*)"
  go test -race "$@"
else
  log "test-integration (go test -race $* ${INTEGRATION_PACKAGES})"
  go test -race "$@" "${INTEGRATION_PACKAGES}"
fi
