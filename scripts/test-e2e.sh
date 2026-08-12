#!/usr/bin/env bash
# Format, lint, ensure test storage, build bin/oauth, run e2e tests.
# Usage: ./scripts/test-e2e.sh [go test args...]
#   ./scripts/test-e2e.sh
#   ./scripts/test-e2e.sh -run Verify -count=1
# Fail-closed if Podman/test storage cannot be ensured (docs/agents/testing.md).
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib.sh
source "${SCRIPT_DIR}/lib.sh"

cd "${PROJECT_ROOT}"
project_quality
require_go

log "ensure e2e test storage"
"${SCRIPT_DIR}/storage.sh" ensure --env test

env_file="${PROJECT_ROOT}/.local/oauth-storage/test.env"
[[ -f "${env_file}" ]] || die "missing ${env_file} after storage ensure"
set -a
# shellcheck disable=SC1090
source "${env_file}"
set +a
[[ -n "${OAUTH_STORAGE_URL:-}" ]] || die "OAUTH_STORAGE_URL not set after sourcing ${env_file}"
[[ "${OAUTH_STORAGE_ENV:-}" == "test" ]] || die "OAUTH_STORAGE_ENV must be test (got ${OAUTH_STORAGE_ENV:-})"

# ADR-0073 suite bootstrap: ensure → truncate → migrate-to-head.
# migrate is not implemented yet; truncate is a no-op until tables exist.
log "truncate e2e test storage (suite bootstrap)"
"${SCRIPT_DIR}/storage.sh" truncate --env test

mkdir -p "${BIN_DIR}"
out="${BIN_DIR}/${BIN_NAME}"
log "build (${CMD_PACKAGE} -> ${out})"
go build -o "${out}" "${CMD_PACKAGE}"
export OAUTH_BIN="${out}"

if (($# == 0)); then
  log "test-e2e (go test -race ${E2E_PACKAGES})"
  go test -race "${E2E_PACKAGES}"
elif go_test_args_have_package "$@"; then
  log "test-e2e (go test -race $*)"
  go test -race "$@"
else
  log "test-e2e (go test -race $* ${E2E_PACKAGES})"
  go test -race "$@" "${E2E_PACKAGES}"
fi
