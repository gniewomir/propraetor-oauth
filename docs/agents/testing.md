# Testing

How agents should run tests in this repo.

## Runners

| Script | Role |
| --- | --- |
| `./scripts/test-unit.sh` | Format + lint (`project_quality`), then `go test -race` for the module **excluding** `./e2e/...` |
| `./scripts/test-e2e.sh` | Format + lint, ensure disposable test storage, build `bin/oauth`, then `go test -race ./e2e/...` |

Both accept extra **`go test` args** (same pattern as the former `scripts/test.sh`), e.g. `-run`, `-count`, package paths.

```bash
# Full unit suite
./scripts/test-unit.sh

# Narrow unit tests by name / package
./scripts/test-unit.sh -run Bootstrap -count=1 ./internal/adapter/postgres/

# Full e2e suite (requires Podman; fail-closed if storage cannot be ensured)
./scripts/test-e2e.sh

# Narrow e2e
./scripts/test-e2e.sh -run Verify -count=1
```

Do **not** use a generic `./scripts/test.sh` — use `test-unit` or `test-e2e` explicitly.

## E2E requirements

- E2E lives in `e2e/` at the repo root and drives the **`bin/oauth` subprocess** (exit codes and Operator messages), not in-process `cli.Run` alone.
- `test-e2e.sh` runs `./scripts/storage.sh ensure --env test`, loads `.local/oauth-storage/test.env` (`OAUTH_STORAGE_URL`, `OAUTH_STORAGE_ENV=test`), then runs the suite.
- If Podman / test storage is unavailable, **e2e fails closed** (non-zero). Do not treat a skip as green.
- Prefer `--env test` only; do not point e2e at the `dev` database.

## `oauth storage verify` (e2e target)

Behavior and locked strings are in [ADR-0072](../adr/0072-operator-storage-cli.md): missing URL, connection failure, and success (`storage: ok`). E2E should assert those exit codes and substrings. Schema/migration checks are **not** part of `verify` yet.
