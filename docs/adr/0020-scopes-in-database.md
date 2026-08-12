# Scope catalog lives in the database

Scopes are rows in Postgres, maintained by the Operator CLI (create/delete names, assign to Clients). This is an Operator-managed catalog, not compile-time config (ADR-0015).
