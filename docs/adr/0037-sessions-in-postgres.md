# Authorization Sessions stored in Postgres

Authorization Sessions that span login and Consent are stored in Postgres so any `cli server` instance can continue the Authorization Request. Process-local in-memory sessions are rejected because they break multi-instance deployments (ADR-0036).
