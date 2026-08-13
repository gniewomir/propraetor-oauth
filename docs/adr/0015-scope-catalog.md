# Scopes are an Operator-managed catalog

Scopes are an Operator-managed catalog stored in the database (not a compile-time fixed set). Scope names are natural primary keys (ADR-0077): RFC 6749 §3.3 `scope-token` charset, case-sensitive exact match on the wire, and create rejects ASCII-case-only duplicates. The Operator CLI defines Scope names, Deactivates them rather than hard-removing rows (ADR-0075), and Allows / Disallows Scopes on a Client allowlist. Access Tokens carry the granted Scopes. Free-form scope strings and scopeless Access Tokens are out of v1.
