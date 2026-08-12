# Rate limiting: auth-focused with stricter login; pluggable store

v1 rate limiting covers a coarse global HTTP cap plus stricter limits on the Authorization Endpoint, Token Endpoint, and login, with login the strictest. Counter persistence is selected explicitly at server start via required CLI store (`postgres` or `memory`; no default—ADR-0068). Multi-instance production must use the shared Postgres-backed store; in-memory is for tests and single-process use only. Redis and other external stores are not required for v1.
