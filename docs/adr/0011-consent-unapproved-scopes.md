# Consent only for unapproved scopes

After End-User login, the AS shows Consent (obtaining Resource Owner authorization) only when the Public Client’s Authorization Request includes Scopes that Resource Owner has not already approved for that Client. Approvals are Consent Grants in Postgres (ADR-0055). First-party deployments still use incremental Consent rather than skipping Resource Owner authorization entirely.
