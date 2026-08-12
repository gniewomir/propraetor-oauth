# Go test database mutations: isolation and cleanup

Research notes from **primary sources** (Go `testing` / `database/sql` / go.dev docs and blogs, PostgreSQL documentation, Testcontainers for Go, Django / Rails official testing docs as general-convention contrast, and first-party patterns visible in well-known Go projects’ docs or test code).  
Retrieved / verified: 2026-08-13.

Scope: how tests that **mutate a shared Postgres** should handle isolation and cleanup — established strategies, tradeoffs, and **Go-specific** conventions vs what ORMs/frameworks usually do elsewhere. Tailored at the end to propraetor-oauth’s disposable Podman Postgres + subprocess e2e (`bin/oauth`).

---

## Executive summary / verdict

Go’s `testing` package gives **lifecycle hooks** (`TestMain`, `t.Cleanup`, `t.Run`, `t.Parallel`) but **no database fixture or cleanup primitives**. Isolation is something you build (or import) on top of SQL and process boundaries.

| Layer | What usually works | What usually fails |
| --- | --- | --- |
| **In-process** integration (repo opens `*sql.DB` / `pgx` and runs queries) | Outer `BEGIN` + `ROLLBACK` (or `t.Cleanup` truncate), shared disposable DB | Assuming `testing` will reset DB for you |
| **Subprocess e2e** (`exec` of CLI / server against real URL) | Truncate / delete, DB- or schema-per-test, container recreate / snapshot restore, namespaced rows | Wrapping the subprocess in the test’s SQL transaction |

**General testing convention** (Django `TestCase`, Rails `use_transactional_tests`) defaults to **transaction rollback per test**. That is excellent for in-process ORMs that share one connection/tx. It is **not** the default Go stdlib pattern, and it **cannot** wrap a separate `oauth` process that opens its own connections and commits.

For propraetor-oauth e2e mutations (migrate, admin CLI, etc.), plan around **committed** writes and explicit cleanup or isolation — not harness-level rollback. See [Decision factors for propraetor-oauth](#decision-factors-for-propraetor-oauth).

---



## Project constraints (this repo)



| Constraint | Source |
| --- | --- |
| E2E in `e2e/` drives **`bin/oauth` subprocess**; separate unit vs e2e runners; fail-closed if test storage unavailable | [docs/agents/testing.md](../agents/testing.md), `scripts/test-e2e.sh` |
| Disposable Podman Postgres via `scripts/storage.sh ensure --env test`; `OAUTH_STORAGE_ENV=test` | same; ADR-0072 notes scripts out of ADR scope |
| `migrate --down` only in `dev`/`test`; auto-migrate `--up` when behind in `dev`/`test` | [ADR-0072](../adr/0072-operator-storage-cli.md) |
| Today e2e is mostly read-only `storage verify`; mutating commands are upcoming | `e2e/storage_verify_test.go`, ADR-0072 |



---



## What Go’s `testing` package actually provides (and does not)



Go testing is package-oriented functions + optional hooks. Relevant primitives:

| Primitive | Role for DB tests |
| --- | --- |
| `func TestXxx(t *testing.T)` | Unit of work; no implicit DB scope |
| `t.Run` / subtests | Shared setup/teardown around a group; fine-grained `-run` | 
| `t.Cleanup(fn)` | LIFO cleanup when the test (and subtests) finish — good place for truncate / drop schema / terminate container |
| `TestMain(m *testing.M)` | Package-level setup/teardown around `m.Run()` (start shared container, migrate once, exit) |
| `t.Parallel()` | Concurrent tests **within** a binary (see parallelism section) |
| `-count`, `-parallel`, test result cache | Affect how often / how concurrently mutable DB state is exercised |

Sources:

- [https://pkg.go.dev/testing](https://pkg.go.dev/testing) (`TestMain`, `(*T).Cleanup`, `(*T).Parallel`, `(*T).Run`)
- [https://go.dev/blog/subtests](https://go.dev/blog/subtests) (setup/teardown around `Run`; parallel groups)
- [https://pkg.go.dev/cmd/go](https://pkg.go.dev/cmd/go) / `go help testflag` (`-parallel`, `-count`, cacheable flags; idiomatic cache bypass: `-count=1`)

**Not provided:** fixtures, factories, transactional test cases, truncate helpers, or “test database” naming. Contrast with Django’s `TestCase` / `TransactionTestCase` and Rails’ transactional tests (below).

**`database/sql` note:** a transaction is an `sql.Tx` from `DB.Begin` / `DB.BeginTx`. Official guidance: manage transactions via those APIs (not raw `BEGIN`/`COMMIT` SQL); do **not** mix non-`Tx` `DB` methods into the same logical transaction — they use other pool connections and run **outside** the tx.

Sources:

- [https://go.dev/doc/database/execute-transactions](https://go.dev/doc/database/execute-transactions)
- [https://go.dev/doc/database/manage-connections](https://go.dev/doc/database/manage-connections) (`sql.DB` is concurrent-safe via a pool; dedicated `DB.Conn` when session affinity is required)

pgx documents the same `defer tx.Rollback(ctx)` idiom as safe after a successful `Commit` (`Rollback` returns `ErrTxClosed`).

Source: [https://github.com/jackc/pgx/blob/master/tx.go](https://github.com/jackc/pgx/blob/master/tx.go) (`Tx.Rollback` comment).

---



## Strategy matrix (overview)



| # | Strategy | Best fit | Go caveats | Subprocess e2e? |
| --- | --- | --- | --- | --- |
| 1 | Truncate / delete between tests | Shared DB, committed writes, sequential or carefully locked parallel | List tables + FKs; `TRUNCATE` takes strong locks | **Yes** (harness SQL against same URL) |
| 2 | Transaction rollback per test | In-process code that can use (or be injected with) one `Tx` / conn | Must not hit `DB` outside the `Tx`; nested app txs need savepoints or redesign | **No** (separate sessions commit) |
| 3 | Schema- or database-per-test | Strong isolation + parallel | `CREATE DATABASE` not in a tx; privileges; migrate cost | **Yes** (pass per-test URL / `search_path`) |
| 4 | Unique prefixes / tenant IDs | Soft isolation when hard reset is costly | Collisions, leftover data, assertions must filter | **Yes** |
| 5 | Ephemeral container / recreate / snapshot | Suite or per-test clean slate | Startup cost; Podman/Docker; snapshot restore not free under parallel | **Yes** (suite already close to this) |
| 6 | Migrate down/up between tests | Testing migrator itself; rare full reset | Slow; down may be lossy; only allowed in test env here | **Yes**, but heavy |
| 7 | Parallelism discipline | Any shared mutable store | `t.Parallel` + shared DB ⇒ races/flakes; `-race` does not see SQL races | Prefer serial e2e or per-test DB |
| 8 | Fixture seeding | Known baseline after reset | Seed once vs per test; order dependence if no reset | **Yes** |



---



## 1. Truncate / delete all tables between tests (or `TestMain`)



### Behavior (Postgres)

`TRUNCATE` empties tables quickly without scanning; can `RESTART IDENTITY`; with `CASCADE` includes FK dependents. It acquires **`ACCESS EXCLUSIVE`** on each table (blocks concurrent use). It is **transaction-safe for table data** in PostgreSQL (truncation rolls back if the surrounding transaction does not commit). It is **not MVCC-safe** for concurrent readers with older snapshots (table appears empty). It does not fire `ON DELETE` triggers (only `ON TRUNCATE`).

Source: [https://www.postgresql.org/docs/current/sql-truncate.html](https://www.postgresql.org/docs/current/sql-truncate.html)

`DELETE` is slower on large data but allows concurrent access patterns that `TRUNCATE`’s exclusive locks would serialize. Prefer truncate when the test DB is dedicated and small; prefer delete (or per-test DB) when you need concurrency.

### When it fits

- Shared disposable test database (exactly propraetor-oauth’s Podman `test` env).
- Tests (or CLI under test) **commit** normally.
- Cleanup in `t.Cleanup` or between subtests; optionally once in `TestMain` only if every test leaves a known seed (usually **per-test** truncate is safer).

### Tradeoffs

| Pros | Cons |
| --- | --- |
| Works across process boundaries | Must maintain table list / discover tables; FK order or `CASCADE` |
| Exercises real commit/visibility paths | Exclusive locks hurt `t.Parallel` on one DB |
| Fast on small schemas | Leaves schema/migration version as-is (good for DML tests; bad if a test mutates schema) |

### Go convention

Common pattern: `TestMain` migrates once; each test `t.Cleanup(truncateAll)`. Libraries such as Testcontainers document **snapshot restore** as an alternative to “heavy cleanup scripts” (see §5) — same problem space, different mechanism.

---



## 2. Transaction rollback per test (`BEGIN` in setup, `ROLLBACK` in teardown)



### Behavior

PostgreSQL transactions are atomic and isolated from other sessions until commit; `ROLLBACK` cancels all updates in the block. Savepoints allow partial undo inside one block.

Source: [https://www.postgresql.org/docs/current/tutorial-transactions.html](https://www.postgresql.org/docs/current/tutorial-transactions.html)

### General testing convention (non-Go)

This is the **default** in major web frameworks:

- **Django `TestCase`:** wraps each test in a transaction and rolls it back; does **not** truncate. Use `TransactionTestCase` when you must observe real commit/rollback (then Django truncates).  
  Sources: [https://docs.djangoproject.com/en/6.1/topics/testing/tools/](https://docs.djangoproject.com/en/6.1/topics/testing/tools/), [https://docs.djangoproject.com/en/6.0/topics/db/transactions/](https://docs.djangoproject.com/en/6.0/topics/db/transactions/)
- **Rails:** wraps tests in a DB transaction rolled back when finished; opt out with `self.use_transactional_tests = false` (then you must clean up yourself).  
  Source: [https://guides.rubyonrails.org/testing.html](https://guides.rubyonrails.org/testing.html) (§ Transactions)

Frameworks can do this because the app and the test harness share the **same** connection/tx abstraction.

### Go / in-process fit

Works when:

1. Code under test accepts a `Tx`-like handle **or** you can force the repository to use one dedicated connection that already began a transaction, **and**
2. The code does not open a second connection from the pool for the same logical work (go.dev warns that `DB` methods outside `Tx` run outside the transaction).

Typical harness:

```go
tx, err := db.BeginTx(ctx, nil)
// ...
t.Cleanup(func() { _ = tx.Rollback() })
// pass tx into repositories
```

### Why this fails for subprocess e2e

A CLI started with `exec.Command` opens **new** server sessions. Uncommitted rows in the test process’s transaction are **invisible** to the subprocess (and vice versa). When the CLI commits, those writes persist after the test’s `ROLLBACK`. There is no supported way for process A’s `ROLLBACK` to undo process B’s committed transaction.

Session-scoped objects (e.g. PostgreSQL `TEMPORARY` tables) are also **per session** — they do not isolate a subprocess.

Source (temp tables session-scoped): [https://www.postgresql.org/docs/current/sql-createtable.html](https://www.postgresql.org/docs/current/sql-createtable.html) (`TEMPORARY` / `ON COMMIT`).

**Implication:** treat rollback-as-fixture as an **in-process integration** tool only. For `e2e/` + `bin/oauth`, choose truncate, namespacing, per-test DB/schema, or container snapshot.

### Other limits (even in-process)

- Cannot honestly test code that must **commit** (advisory locks, `LISTEN`/`NOTIFY`, some concurrency, migration runners that commit DDL).
- Nested transactions: Postgres uses savepoints; app code that always `Begin` on `*sql.DB` will start a **sibling** connection, not nest in the test tx.
- `CREATE DATABASE` cannot run inside a transaction block.  
  Source: [https://www.postgresql.org/docs/current/sql-createdatabase.html](https://www.postgresql.org/docs/current/sql-createdatabase.html)

---



## 3. Schema-per-test or database-per-test



### Database-per-test

`CREATE DATABASE ... TEMPLATE ...` clones a template (often a pre-migrated template DB). Requires `CREATEDB` (or superuser). Template DB must have **no other connections** while cloning. Name length / uniqueness matter (Postgres identifier limits).

Source: [https://www.postgresql.org/docs/current/sql-createdatabase.html](https://www.postgresql.org/docs/current/sql-createdatabase.html)

**Fits:** parallel tests; migration tests that need a virgin catalog; avoiding truncate lock fights.

**Cost:** create/drop (or pool of DBs) per test; connection-string plumbing into subprocess env (`OAUTH_STORAGE_URL`).

### Schema-per-test

`CREATE SCHEMA ...`; set `search_path` (or qualify names). Schemas organize objects inside one database; same unqualified names can exist in different schemas.

Source: [https://www.postgresql.org/docs/current/ddl-schemas.html](https://www.postgresql.org/docs/current/ddl-schemas.html)

**Fits:** cheaper than full DB clone if migrations install into `search_path`; parallel-friendly if each test owns a schema.

**Caveats for this product:** the AS/admin CLI must honor `search_path` or schema-qualified DDL. If migrations and app SQL assume `public` only, schema-per-test needs product support or is a non-starter. `search_path` trust/security notes in the same Postgres chapter matter less on a disposable test DB but still affect whether unqualified SQL hits the right objects.

---



## 4. Unique prefixes / tenant IDs / namespaced data



### Behavior

Each test generates unique client IDs, usernames, redirect URIs, etc., and asserts only on its own rows. Optional delayed cleanup (or truncate at suite end).

### When it fits

- Soft multi-tenant data model (or natural unique keys).
- Parallelism desired without per-test databases.
- Mutations are inserts/updates of distinct keys, not “wipe world” or global singleton rows.

### Tradeoffs

| Pros | Cons |
| --- | --- |
| No truncate locks; easy with subprocess | Leftover data; disk growth; flaky if any global unique constraint / singleton config |
| Simple mentally for CRUD admin tests | Hard for “list all” / purge-everything / migrate-down semantics |
| Composes with a final suite truncate | Requires disciplined factories |

OAuth admin surfaces (clients, users, scopes) often have enough unique keys for this; **migration** and **purge-all** tests usually do not.

---



## 5. Ephemeral container / recreate DB per suite or per test



### Suite-level disposable Postgres

Start (or `ensure`) one container for the package/`TestMain` or for the e2e script; migrate once; run all tests; tear down optionally. This is what `scripts/test-e2e.sh` + `storage.sh ensure --env test` already approximate (Podman, fail-closed).

### Per-test container

Maximum isolation; highest latency. Rarely needed for every test if cleanup is solid.

### Snapshot / restore (Testcontainers Postgres)

Official Testcontainers for Go Postgres module: run migrations once, `Snapshot`, then each test `Restore` in `t.Cleanup` so each case gets a clean DB **without** recreating the container or hand-written truncate scripts. Warning: do not use database name `postgres` if you rely on snapshot (restore drops the connected DB and uses the system DB). Snapshot/restore is illustrated with **sequential** subtests restoring in cleanup — parallel restores of one DB will conflict unless each test has its own DB/snapshot.

Source: [https://golang.testcontainers.org/modules/postgres/](https://golang.testcontainers.org/modules/postgres/)

Ory dockertest similarly documents pool + resource lifecycle with `TestMain` or `t.Cleanup`.

Source: [https://pkg.go.dev/github.com/ory/dockertest/v3](https://pkg.go.dev/github.com/ory/dockertest/v3)

**For this repo:** you already own Podman lifecycle outside Go. Options are: keep script-managed suite DB + truncate/namespace; or add in-test snapshot/clone later. No mandate to switch to Testcontainers if Podman scripts remain the source of truth.

---



## 6. Migrating down/up between tests



### When it fits

- Tests whose **subject** is the migrator (`oauth storage migrate --up` / `--down`).
- Occasional “hard reset” when schema drift is the risk (not every DML test).

### Tradeoffs

| Pros | Cons |
| --- | --- |
| Validates down scripts in `test` env (ADR-0072 allows `--down` only in `dev`/`test`) | Slow; ordering fragility; partial failure leaves odd versions |
| Full schema rebuild without new container | Down migrations are often lossy or under-tested in the industry |
| Aligns with auto-migrate-on-behind in test | Parallelism almost impossible on one DB |

**Practice:** run a **small dedicated** migrator suite (serial) that goes up/down; keep ordinary admin/DML e2e on a migrated baseline + truncate/namespace. Do not down/up around every client-create test.

---



## 7. Parallelism implications (`t.Parallel()`, shared DB, `-count`, race detector)



### What `t.Parallel` means

Calling `t.Parallel()` allows that test to run concurrently with **other parallel tests** in the same binary (cap: `-parallel`, default `GOMAXPROCS`). A parallel test does not run concurrently with sequential tests; a parent waits for its parallel subtests (useful for shared setup/teardown barriers).

Sources:

- [https://pkg.go.dev/testing#T.Parallel](https://pkg.go.dev/testing#T.Parallel) (also: multiple instances from `-count` / `-cpu` of the **same** test never run in parallel with each other)
- [https://go.dev/blog/subtests](https://go.dev/blog/subtests)

`go test -parallel` only applies **within** one test binary; packages may still run in parallel via `go test -p`.

Source: `go help testflag` / [https://pkg.go.dev/cmd/go](https://pkg.go.dev/cmd/go)

### Shared mutable DB + parallel = flakes

Concurrent truncates (`ACCESS EXCLUSIVE`), concurrent migrates, or concurrent asserts on shared rows will flake. Choices:

1. **Do not** call `t.Parallel()` in e2e packages that share one URL; or  
2. Give each parallel test its own database/schema; or  
3. Namespace rows and never truncate mid-suite under parallel.

### `-count` and caching

`-count n` runs tests n times (except examples). Successful package results may be **cached**; the idiomatic way to disable caching is `-count=1`. Cached green runs do **not** re-hit Postgres — dangerous if you are debugging storage flakes and forget `-count=1`.

Source: `go help testflag` (cacheable flags; `-count=1`).

### Race detector

`go test -race` finds **Go memory** races in instrumented code. It does **not** detect lost updates or isolation bugs **inside Postgres**. This repo’s e2e already uses `-race` on the test binary; the CLI subprocess is a separate program (not automatically race-instrumented unless built with `-race` too). Shared DB contention remains a logical/flaky-test problem, not something `-race` will reliably report.

Source: [https://go.dev/blog/race-detector](https://go.dev/blog/race-detector)

### First-party Go pattern note (pgx)

pgx’s own transaction tests often use `t.Parallel()`, a shared `PGX_TEST_DATABASE`, and **`CREATE TEMPORARY TABLE`** so each connection’s fixture is session-private — isolation without truncating shared permanent tables. That pattern does **not** transfer to CLI subprocess e2e (different sessions), but it is a useful **in-process** idiom when you control the connection.

Source: e.g. [https://github.com/jackc/pgx/blob/master/tx_test.go](https://github.com/jackc/pgx/blob/master/tx_test.go) (`t.Parallel`, temporary tables).

---



## 8. Fixture seeding approaches



| Approach | Description | Notes |
| --- | --- | --- |
| **Migrate-only baseline** | Schema at head; empty tables; each test inserts what it needs | Good default for admin CLI e2e |
| **Seed SQL / init scripts** | Load reference rows once (container init or `TestMain`) | Must truncate carefully (preserve seed vs wipe all) or re-seed after truncate |
| **Per-test factories** | Helpers create Client/User/… with unique names | Composes with truncate or namespacing |
| **Snapshot after seed** | Migrate + seed + snapshot; restore per test | Testcontainers pattern; strong isolation |
| **ORM fixtures / fixtures files** | Common in Rails/Django | **No stdlib equivalent** in Go; roll your own or skip |

Django also documents that migration-loaded initial data interacts differently with `TestCase` vs `TransactionTestCase` (transactional tests may not see committed migration data the same way). Go has no such framework split — you explicitly choose commit vs rollback.

Source: [https://docs.djangoproject.com/en/6.1/topics/testing/overview/](https://docs.djangoproject.com/en/6.1/topics/testing/overview/)

---



## In-process integration vs subprocess e2e (critical distinction)



```text
In-process integration                     Subprocess e2e (this repo’s e2e/)
─────────────────────                      ────────────────────────────────
test goroutine                             test goroutine
   │                                          │
   ├─ Begin Tx ─────────────────┐             ├─ exec bin/oauth
   │                            │             │      │
   ├─ repo.Create(tx, …)        │             │      ├─ opens own pool/conn
   ├─ assert via tx             │             │      ├─ BEGIN/COMMIT itself
   └─ Rollback  ← undoes ───────┘             │      └─ process exits
                                              └─ harness must Truncate /
                                                 DROP SCHEMA / Restore /
                                                 use unique IDs
```

| Concern | In-process | Subprocess e2e |
| --- | --- | --- |
| Transactional fixtures | Viable if API takes `Tx`/conn | **Not viable** for cleanup of CLI writes |
| Seeing uncommitted harness rows | Yes (same tx) | No |
| Testing Operator messages / exit codes | Indirect | Direct (propraetor-oauth intent) |
| Auto-migrate on behind (ADR-0072) | Can unit-test boot path | Will **mutate schema** on shared test DB if behind — suite must start migrated or accept auto-up |
| Parallelism | Easier with tx or temp tables | Prefer serial shared DB or DB-per-test |



---



## Go conventions vs general testing conventions



| Topic | General (Django/Rails-style) | Go ecosystem / stdlib |
| --- | --- | --- |
| Default isolation | Transaction rollback per test | **No default**; you choose |
| Flushing when txs don’t apply | Truncate / rebuild test DB | Truncate, recreate DB, or container snapshot — same ideas, hand-rolled |
| Test DB lifecycle | Framework creates `test_*` DB | `TestMain`, scripts (as here), or Testcontainers/dockertest |
| Fixtures | First-class | Helpers you write; table-driven tests for cases, not for rows |
| Parallel DB tests | Often discouraged on one DB | Explicit `t.Parallel`; easy to shoot yourself in the foot with shared URL |
| Honesty about commits | Separate `TransactionTestCase` / opt-out of transactional tests | Natural if you never wrapped a tx — subprocess e2e is always “honest commits” |

**Takeaway:** do not expect Go to feel like Django `TestCase`. Prefer **explicit** cleanup and a clear split: fast in-process tests (optional rollback) vs slower e2e (committed + truncate/namespace/snapshot).

---



## Decision factors for propraetor-oauth



Context: disposable Podman Postgres (`storage.sh ensure --env test`), e2e via `exec` of `bin/oauth`, `-race` in `test-e2e.sh`, upcoming **mutating** Operator commands (`migrate`, admin CRUD, purge, …), ADR-0072 **`--down` only in test/dev** and **auto-migrate up when behind** in test.

### Recommended options (not a single mandate)

**A. Suite DB + per-test truncate (default candidate for DML e2e)**  
Keep one test database from `storage.sh`. In `TestMain` or first test, ensure schema at head (`migrate --up` or rely on auto-migrate once). Each mutating test registers `t.Cleanup` that `TRUNCATE … CASCADE` (or deletes in FK-safe order) for app tables.  
- *Pros:* Matches subprocess reality; simple mental model; fits fail-closed shared URL.  
- *Cons:* Maintain truncate list; avoid `t.Parallel` on this package (or accept locking/flakes).  
- *Fits:* admin create/update/revoke tests.

**B. Unique IDs without truncate (optional for narrow CRUD)**  
Factories mint unique client_id / username; assertions filter by those IDs; periodic or end-of-suite truncate.  
- *Pros:* Less lock churn; some parallelism possible.  
- *Cons:* Weak for list/purge/migrate; debris if cleanup skipped.  
- *Fits:* parallel-friendly insert/get tests only if you never assert global emptiness.

**C. Serial migrator e2e (dedicated `-run` group)**  
Separate tests for `migrate --up` / `--down` that **own** schema version, run **without** `t.Parallel`, and restore to head in `Cleanup` (up again). Do not interleave with CRUD tests on the same DB without barriers.  
- *Pros:* Actually tests ADR-0072 down/up policy.  
- *Cons:* Slow; ordering hazards; auto-migrate in other commands can surprise you if version left dirty.

**D. Database-per-test or snapshot restore (escalation)**  
If truncate lists become painful or you need parallel e2e: clone from a migrated template DB per test, or adopt snapshot/restore (Testcontainers-style) while still launching `bin/oauth` with a per-test URL.  
- *Pros:* Strong isolation.  
- *Cons:* More infra than current scripts; template connection rules; ADR-0035/ops complexity.

**E. In-process integration (optional companion, not a substitute for e2e)**  
For storage adapters and domain+Postgres wiring, use `database/sql`/`pgx` tests with rollback or temp tables **inside** `internal/...` unit/integration packages run by `test-unit.sh` (or a future integration tag). Keep Operator **message/exit-code** contracts in `e2e/`.  
- *Pros:* Fast feedback; rollback works.  
- *Cons:* Does not replace subprocess contract tests.

### Practical defaults to prefer unless contradicted

1. **E2E mutating tests: committed writes + explicit cleanup** (A or B). Do not design e2e around harness `BEGIN`/`ROLLBACK`.  
2. **Keep e2e package serial** w.r.t. the shared test DB unless/until per-test DBs exist.  
3. **Migrate tests: small serial slice** (C); leave ordinary tests on “schema at head.”  
4. **Always use `-count=1` when diagnosing storage flakes** so results are not cached.  
5. **Ensure migrated baseline** before DML suites so ADR-0072 auto-migrate does not surprise mid-run.  
6. **Use `t.Cleanup`** (not only deferred functions in helpers) so cleanup runs on failure paths.  
7. Revisit **D** only if A/B become a maintenance or parallelism bottleneck.

### Anti-patterns for this repo

- Wrapping `runOAuth(...)` in a test-held transaction and expecting rollback to undo CLI commits.  
- Calling `t.Parallel()` freely in `e2e/` against one `OAUTH_STORAGE_URL` while truncating.  
- Down/up migrating around every admin test.  
- Pointing e2e at the `dev` database (already forbidden in testing docs).

---



## Source index



| Topic | URL |
| --- | --- |
| Go `testing` | https://pkg.go.dev/testing |
| Subtests / parallel groups / teardown | https://go.dev/blog/subtests |
| `go test` flags / cache / `-count` | https://pkg.go.dev/cmd/go (`go help testflag`) |
| Race detector | https://go.dev/blog/race-detector |
| `database/sql` transactions | https://go.dev/doc/database/execute-transactions |
| Connection pool / dedicated conn | https://go.dev/doc/database/manage-connections |
| pgx `Tx.Rollback` | https://github.com/jackc/pgx/blob/master/tx.go |
| PostgreSQL `TRUNCATE` | https://www.postgresql.org/docs/current/sql-truncate.html |
| PostgreSQL transactions | https://www.postgresql.org/docs/current/tutorial-transactions.html |
| PostgreSQL `CREATE DATABASE` | https://www.postgresql.org/docs/current/sql-createdatabase.html |
| PostgreSQL schemas / `search_path` | https://www.postgresql.org/docs/current/ddl-schemas.html |
| PostgreSQL `TEMPORARY` tables | https://www.postgresql.org/docs/current/sql-createtable.html |
| Testcontainers Postgres (+ snapshot) | https://golang.testcontainers.org/modules/postgres/ |
| Django test DB tools | https://docs.djangoproject.com/en/6.1/topics/testing/tools/ |
| Rails transactional tests | https://guides.rubyonrails.org/testing.html |
| Repo testing / ADR-0072 | [docs/agents/testing.md](../agents/testing.md), [docs/adr/0072-operator-storage-cli.md](../adr/0072-operator-storage-cli.md) |
