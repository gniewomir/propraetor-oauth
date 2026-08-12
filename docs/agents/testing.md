# Testing

How agents should run tests and how suites treat Postgres.

Decision record: [ADR-0073](../adr/0073-test-suite-storage-isolation.md). Research background: [go-test-database-mutations](../research/go-test-database-mutations.md).

## Runners

| Script | Role |
| --- | --- |
| `./scripts/test-unit.sh` | Format + lint (`project_quality`), then `go test -race` excluding `./e2e/...` and `./integration/...` (no Postgres) |
| `./scripts/test-integration.sh` | Format + lint, ensure integration storage, then `go test -race ./integration/...` |
| `./scripts/test-e2e.sh` | Format + lint, ensure e2e test storage, build `bin/oauth`, then `go test -race ./e2e/...` |

All accept extra **`go test` args** (e.g. `-run`, `-count`, package paths).

```bash
./scripts/test-unit.sh
./scripts/test-unit.sh -run Bootstrap -count=1 ./internal/adapter/postgres/

./scripts/test-integration.sh
./scripts/test-integration.sh -run Ping -count=1

./scripts/test-e2e.sh
./scripts/test-e2e.sh -run Verify -count=1
```

Do **not** use a generic `./scripts/test.sh` — pick `test-unit`, `test-integration`, or `test-e2e` explicitly.

Migrator-e2e runner is not wired yet; when added it must use its **own** `storage.sh` env (never share integration/e2e DBs, never use `dev`).

## Storage helpers

| Command | When |
| --- | --- |
| `./scripts/storage.sh ensure --env <suite>` | Suite runner only (or local DX). Fail closed if Podman/storage unavailable — no skip-as-green. |
| `./scripts/storage.sh truncate --env <suite>` | Suite bootstrap only: empty all `public` tables (no-op if none). Not between tests. |
| `./scripts/storage.sh create\|remove --env <suite>` | Recreate/teardown disposable DB for that env |

After ensure, load `.local/oauth-storage/<env>.env` (`OAUTH_STORAGE_URL`, `OAUTH_STORAGE_ENV`). Suite envs: `test` (e2e), `integration`. Do not point any suite at `dev`.

Suite bootstrap (once per run): **ensure → truncate → migrate to Storage schema head**. Runners do ensure + truncate; migrate-to-head lands when `oauth storage migrate` exists. Per-test cleanup differs by suite (below).

## Suite constraints

| Suite | Process | DB | Per-test isolation | Parallel |
| --- | --- | --- | --- | --- |
| Unit | in-process | none | — | yes |
| Integration | in-process (`integration/`) | `--env integration` | transaction + rollback; always empty | yes (default) |
| E2E | `bin/oauth` subprocess (`e2e/`) | `--env test` | unique IDs; no harness tx; no mid-suite truncate | yes (default) |
| Migrator e2e | subprocess | own disposable DB | owns schema version; serial | no |

Shared rules for Integration and E2E:

- Prefer real APIs for fixtures (ports/adapters in-process; Operator CLI in e2e). Direct DB surgery only with a strong reason.
- Irrelevant fields: random. Relevant fields: explicit in the test.
- Never wrap `exec` of `bin/oauth` in a harness transaction expecting rollback to undo CLI commits.

E2E asserts Operator **exit codes and messages**, not in-process `cli.Run` alone.

## `oauth storage verify` (current e2e target)

Locked strings and exits: [ADR-0072](../adr/0072-operator-storage-cli.md). Schema/migration checks are **not** part of `verify` yet.
