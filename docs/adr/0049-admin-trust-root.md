# Postgres credentials are Operator authority

Possession of the configured Postgres credentials is Operator authority for admin CLI in v1. There is no separate Operator password or admin authn factor. Protecting database credentials and network access to Postgres is the Operator trust boundary (alongside ADR-0022).
