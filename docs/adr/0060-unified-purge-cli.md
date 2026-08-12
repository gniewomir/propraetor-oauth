# Unified purge CLI; no automatic purge

Expired Authorization Codes, Refresh Tokens, and Authorization Sessions, and Audit Events, are removed only via Operator CLI. A single command `cli purge --target=entities|audit|all --older-than D` requires an explicit window every run (no standing retention). `entities` covers expired codes, refresh tokens, and Authorization Sessions; `audit` covers Audit Events (ADR-0039/0041); `all` runs both. `cli server` does not background-purge or lazily delete on access—DB growth until Operator runs purge is an accepted v1 tradeoff.
