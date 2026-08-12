# Test suites: separate DBs; integration uses txs; e2e uses unique IDs

Unit tests do not use Postgres. Integration and e2e (and migrator e2e) each get their **own disposable test database**—no shared URL, no cross-suite lock file. Suite runners ensure storage, truncate, and migrate to the current Storage schema once before tests; they never point at `dev`.

**Integration** is in-process against real Postgres: parallel by default; each test runs in a transaction and rolls back so the DB stays empty; fixtures use real internal APIs (direct SQL only with a strong reason); irrelevant fields are random, relevant ones explicit.

**E2E** drives `bin/oauth` as a subprocess: parallel by default; no harness transactions (the CLI commits on its own connections); no mid-suite truncate; isolation is unique IDs so debris from other tests is ignored. Setup goes through the Operator CLI (or helpers that invoke it), not DB surgery.

**Migrator e2e** is a separate suite (own DB): serial schema `--up`/`--down` tests; not interleaved with CRUD e2e.

Rejected: wrapping subprocess e2e in harness `BEGIN`/`ROLLBACK`; one shared test DB with a lock between suites; truncate-between-tests as the default e2e isolation; down/up around every admin test.
