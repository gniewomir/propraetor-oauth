# Audit PII rules and purge UX

Audit Events may include username, Client Identifier, IP, outcome, and reason codes, and may reference database entities involved (e.g. ids for Authorization Codes, Refresh Tokens, Clients, Users/End-Users) without storing raw secrets (passwords, token/code plaintext, client secrets). Purge requires an explicit `--older-than` window each run via `cli purge` (ADR-0060); there is no standing retention policy enforced at server start.
