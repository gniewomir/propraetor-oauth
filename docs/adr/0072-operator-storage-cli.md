# Operator storage CLI: bootstrap SQL, verify, migrate, env URL

Postgres credentials for the AS and admin CLI are the app role URL in `OAUTH_STORAGE_URL` (ADR-0006, 0022, 0049). Operator maintenance of that database is under `oauth storage`, parallel to `oauth policy` for Server Policy.

`oauth storage bootstrap-sql --prefix <instance>` prints SQL for a superuser to create database and role `{prefix}_oauth` with a generated random password. The same invocation emits `OAUTH_STORAGE_URL` for that role alongside the SQL. Output is secret-bearing (no password placeholder). Day-to-day processes use only that role; superuser is for applying bootstrap SQL once.

`oauth storage verify` checks that `OAUTH_STORAGE_URL` is set and that the process can open storage (connectivity, e.g. `SELECT 1`). It does **not** ensure Storage schema migration version yet—that joins the shared boot path later. Locked Operator-facing messages: stderr `OAUTH_STORAGE_URL is not set` when the URL is missing or empty; stderr `storage: connection failed` when the URL is present but storage is unreachable; stdout `storage: ok` on success. Missing URL or connection failure exits non-zero.

`oauth storage migrate` requires an explicit direction: `--up` or `--down`. Bare `migrate` errors. `--up` advances the Storage schema to head. `--down` is allowed only when `OAUTH_STORAGE_ENV` is `dev` or `test`; otherwise it refuses.

Any command that opens Postgres shares a boot path that checks migration version against the expected Storage schema. If behind and `OAUTH_STORAGE_ENV` is `dev` or `test`, it auto-applies `--up`. If behind otherwise (including unset env), it fails closed until the Operator runs `oauth storage migrate --up`. (`verify` is exempt until schema checks are added to it.)

Local disposable Postgres helpers (e.g. Podman scripts) and test runners are out of scope of this ADR.
