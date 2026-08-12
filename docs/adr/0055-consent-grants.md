# Persistent Consent Grants with Operator revocation

Resource Owner approval of Scopes for a Public Client is stored as Consent Grants in Postgres (one row per End-User × Client × Scope). Incremental Consent (ADR-0011) skips already-granted Scopes. If the Operator removes a Scope from a Client’s allowlist, matching Grants become invalid and the next Authorization Request requires Consent again. The Operator CLI can revoke Grants per End-User, per End-User × Client, or per End-User × Client × Scope; every revocation is audited.
