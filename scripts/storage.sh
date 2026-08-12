#!/usr/bin/env bash
# Local disposable Postgres for Operator DX (not an ADR; ADR-0072).
# Usage:
#   ./scripts/storage.sh create|ensure|remove --env dev|test
# Then:
#   set -a && source .local/oauth-storage/<env>.env && set +a
# shellcheck shell=bash

set -euo pipefail

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
readonly PROJECT_ROOT

STATE_DIR="${PROJECT_ROOT}/.local/oauth-storage"
readonly STATE_DIR

IMAGE="${OAUTH_STORAGE_IMAGE:-docker.io/library/postgres:16-alpine}"
ROOT_USER="oauth_storage_root"
ROOT_PASSWORD="oauth-storage-local-root"

log() { printf '==> %s\n' "$*" >&2; }
die() { printf 'error: %s\n' "$*" >&2; exit 1; }

usage() {
  cat >&2 <<'EOF'
Usage: storage.sh create|ensure|remove --env dev|test

  create   Create a fresh container+volume, bootstrap app role, write env file (fails if present)
  ensure   Create if missing; start if stopped; require env file
  remove   Remove container, volume, and env file

After create/ensure, load credentials with:
  set -a && source .local/oauth-storage/<env>.env && set +a

Ports: 55432 (dev), 55433 (test). Container names: oauth-storage-<env>.
EOF
}

require_podman() {
  command -v podman >/dev/null 2>&1 || die "podman not found on PATH"
  if ! podman info >/dev/null 2>&1; then
    log "podman not ready; trying 'podman machine start'"
    podman machine start >/dev/null 2>&1 || die "podman is not usable (start/init podman machine first)"
  fi
}

env_file() {
  printf '%s/%s.env' "${STATE_DIR}" "$1"
}

container_name() {
  printf 'oauth-storage-%s' "$1"
}

volume_name() {
  printf 'oauth-storage-%s-data' "$1"
}

host_port() {
  case "$1" in
    dev) printf '55432' ;;
    test) printf '55433' ;;
    *) die "unknown env: $1 (want dev|test)" ;;
  esac
}

parse_args() {
  ACTION="${1:-}"
  shift || true
  ENV_NAME=""
  while (($#)); do
    case "$1" in
      --env)
        ENV_NAME="${2:-}"
        [[ -n "${ENV_NAME}" ]] || die "--env requires dev|test"
        shift 2
        ;;
      -h | --help)
        usage
        exit 0
        ;;
      *)
        die "unknown argument: $1"
        ;;
    esac
  done
  case "${ACTION}" in
    create | ensure | remove) ;;
    *)
      usage
      die "action required: create|ensure|remove"
      ;;
  esac
  case "${ENV_NAME}" in
    dev | test) ;;
    *)
      usage
      die "--env required: dev|test"
      ;;
  esac
}

container_exists() {
  podman container exists "$(container_name "$1")"
}

container_running() {
  [[ "$(podman inspect -f '{{.State.Running}}' "$(container_name "$1")" 2>/dev/null || true)" == "true" ]]
}

wait_ready() {
  local name="$1"
  local _
  for _ in $(seq 1 60); do
    if podman exec "${name}" pg_isready -U "${ROOT_USER}" -d postgres >/dev/null 2>&1; then
      return 0
    fi
    sleep 0.5
  done
  die "postgres in ${name} did not become ready"
}

run_bootstrap_sql() {
  local env_name="$1"
  local port="$2"
  local sql_file url_file
  sql_file="$(mktemp)"
  url_file="$(mktemp)"
  # shellcheck disable=SC2064
  trap "rm -f '${sql_file}' '${url_file}'" RETURN

  (
    cd "${PROJECT_ROOT}"
    if [[ -x "${PROJECT_ROOT}/bin/oauth" ]]; then
      "${PROJECT_ROOT}/bin/oauth" storage bootstrap-sql \
        --prefix "${env_name}" --host 127.0.0.1 --port "${port}" \
        >"${sql_file}" 2>"${url_file}"
    else
      command -v go >/dev/null 2>&1 || die "go not found and bin/oauth missing"
      go run ./cmd/oauth storage bootstrap-sql \
        --prefix "${env_name}" --host 127.0.0.1 --port "${port}" \
        >"${sql_file}" 2>"${url_file}"
    fi
  )

  podman exec -i "$(container_name "${env_name}")" \
    psql -v ON_ERROR_STOP=1 -U "${ROOT_USER}" -d postgres <"${sql_file}" >/dev/null

  local url_line
  url_line="$(grep '^OAUTH_STORAGE_URL=' "${url_file}" || true)"
  [[ -n "${url_line}" ]] || die "bootstrap-sql did not emit OAUTH_STORAGE_URL on stderr"

  mkdir -p "${STATE_DIR}"
  umask 077
  cat >"$(env_file "${env_name}")" <<EOF
${url_line}
OAUTH_STORAGE_ENV=${env_name}
EOF
}

create_cluster() {
  local env_name="$1"
  local cname vname port
  cname="$(container_name "${env_name}")"
  vname="$(volume_name "${env_name}")"
  port="$(host_port "${env_name}")"

  if container_exists "${env_name}"; then
    die "container ${cname} already exists (use ensure or remove)"
  fi
  if podman volume exists "${vname}"; then
    die "volume ${vname} already exists (use remove)"
  fi

  require_podman
  log "creating ${cname} on host port ${port}"
  podman volume create "${vname}" >/dev/null
  podman run -d \
    --name "${cname}" \
    -p "${port}:5432" \
    -e POSTGRES_USER="${ROOT_USER}" \
    -e POSTGRES_PASSWORD="${ROOT_PASSWORD}" \
    -e POSTGRES_DB=postgres \
    -v "${vname}:/var/lib/postgresql/data:Z" \
    "${IMAGE}" >/dev/null

  wait_ready "${cname}"
  log "bootstrapping app role (${env_name}_oauth)"
  run_bootstrap_sql "${env_name}" "${port}"
  log "wrote $(env_file "${env_name}")"
}

ensure_cluster() {
  local env_name="$1"
  local cname
  cname="$(container_name "${env_name}")"

  require_podman
  if ! container_exists "${env_name}"; then
    if podman volume exists "$(volume_name "${env_name}")"; then
      die "volume $(volume_name "${env_name}") exists without container; run remove then ensure"
    fi
    create_cluster "${env_name}"
    return
  fi

  if ! container_running "${env_name}"; then
    log "starting ${cname}"
    podman start "${cname}" >/dev/null
    wait_ready "${cname}"
  fi

  if [[ ! -f "$(env_file "${env_name}")" ]]; then
    die "container ${cname} exists but $(env_file "${env_name}") is missing; run remove then create/ensure"
  fi
}

remove_cluster() {
  local env_name="$1"
  local cname vname
  cname="$(container_name "${env_name}")"
  vname="$(volume_name "${env_name}")"

  require_podman
  if container_exists "${env_name}"; then
    log "removing container ${cname}"
    podman rm -f "${cname}" >/dev/null
  fi
  if podman volume exists "${vname}"; then
    log "removing volume ${vname}"
    podman volume rm "${vname}" >/dev/null
  fi
  if [[ -f "$(env_file "${env_name}")" ]]; then
    log "removing $(env_file "${env_name}")"
    rm -f "$(env_file "${env_name}")"
  fi
}

print_env_hint() {
  local env_name="$1"
  local file
  file="$(env_file "${env_name}")"
  [[ -f "${file}" ]] || die "missing ${file}"
  log "env file: ${file}"
  log "load with: set -a && source ${file} && set +a"
}

parse_args "$@"
case "${ACTION}" in
  create)
    create_cluster "${ENV_NAME}"
    print_env_hint "${ENV_NAME}"
    ;;
  ensure)
    ensure_cluster "${ENV_NAME}"
    print_env_hint "${ENV_NAME}"
    ;;
  remove)
    remove_cluster "${ENV_NAME}"
    ;;
esac
