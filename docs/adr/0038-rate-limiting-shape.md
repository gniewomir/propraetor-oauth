# Rate limiting: auth-focused with stricter login; pluggable store

v1 rate limiting covers a coarse global HTTP cap plus stricter limits on the Authorization Endpoint, Token Endpoint, and login, with login the strictest. Counter persistence is Postgres by default for multi-instance correctness. The persistence port is switchable by design: an in-memory store is allowed for tests and single-process use only; multi-instance production must use the shared Postgres-backed store. Redis and other external stores are not required for v1.
