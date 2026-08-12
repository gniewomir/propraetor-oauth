# Rate limiting: auth-focused with stricter login; Postgres counters

v1 rate limiting covers a coarse global HTTP cap plus stricter limits on the Authorization Endpoint, Token Endpoint, and login, with login the strictest. Counter persistence is Postgres with the rest of system state (ADR-0006)—no Operator-selectable store and no product in-memory backend (ADR-0068). A separate rate-limit backend (e.g. Redis) is out of v1 until requirements name it. Multi-instance deployments share counters via Postgres.
