# Audit to Postgres and stdout; explicit purge window

Security-relevant Audit Events are written to Postgres and also emitted as structured logs on stdout/stderr. Coverage is the v1 security set (login, Consent, Access Token / Refresh Token issue outcomes, admin mutations, rate-limit trips)—not every HTTP request. Purge is via unified `cli purge --target=… --older-than` (ADR-0060); there is no standing retention policy configured on the server (ADR-0041).
