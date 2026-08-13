# Scope catalog lives in the database

Scopes are rows in Postgres, maintained by the Operator CLI (create names with ADR-0077 string rules, soft-delete/deactivate, assign to Clients — ADR-0075). This is an Operator-managed catalog, not compile-time config (ADR-0015).
