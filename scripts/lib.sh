#!/usr/bin/env bash
# Shared helpers for project scripts (format + lint).
# shellcheck shell=bash

set -euo pipefail

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
readonly PROJECT_ROOT

MODULE_PACKAGES="./..."
CMD_PACKAGE="./cmd/oauth"
E2E_PACKAGES="./e2e/..."
BIN_DIR="${PROJECT_ROOT}/bin"
BIN_NAME="oauth"

# STATICCHECK_VERSION pins `go run` when staticcheck is not on PATH.
# Bump deliberately; matches research baseline (staticcheck alongside go vet).
STATICCHECK_VERSION="${STATICCHECK_VERSION:-2025.1.1}"

log() {
  printf '==> %s\n' "$*"
}

die() {
  printf 'error: %s\n' "$*" >&2
  exit 1
}

require_go() {
  command -v go >/dev/null 2>&1 || die "go not found on PATH"
  command -v gofmt >/dev/null 2>&1 || die "gofmt not found on PATH"
}

# project_format writes canonical gofmt style for all .go files in the module.
project_format() {
  require_go
  log "format (gofmt -w)"
  local files=()
  while IFS= read -r -d '' f; do
    files+=("$f")
  done < <(find "${PROJECT_ROOT}" -type f -name '*.go' \
    ! -path '*/vendor/*' \
    ! -path '*/.git/*' \
    -print0)

  if ((${#files[@]} == 0)); then
    log "format: no .go files"
    return 0
  fi
  gofmt -w "${files[@]}"
}

# project_lint runs go vet and staticcheck (and golangci-lint when installed).
project_lint() {
  require_go
  log "lint (go vet)"
  (cd "${PROJECT_ROOT}" && go vet ${MODULE_PACKAGES})

  log "lint (staticcheck)"
  if command -v staticcheck >/dev/null 2>&1; then
    (cd "${PROJECT_ROOT}" && staticcheck ${MODULE_PACKAGES})
  else
    (cd "${PROJECT_ROOT}" && go run "honnef.co/go/tools/cmd/staticcheck@${STATICCHECK_VERSION}" ${MODULE_PACKAGES})
  fi

  if command -v golangci-lint >/dev/null 2>&1; then
    log "lint (golangci-lint)"
    (cd "${PROJECT_ROOT}" && golangci-lint run)
  else
    log "lint: golangci-lint not on PATH (optional; skipped)"
  fi
}

project_quality() {
  project_format
  project_lint
}

# unit_packages lists module packages excluding e2e.
unit_packages() {
  (cd "${PROJECT_ROOT}" && go list ./... | grep -v '/e2e$')
}

# go_test_args_have_package returns 0 if any arg looks like a package pattern.
go_test_args_have_package() {
  local a
  for a in "$@"; do
    case "$a" in
      ./* | ../* | github.com/*) return 0 ;;
    esac
  done
  return 1
}
