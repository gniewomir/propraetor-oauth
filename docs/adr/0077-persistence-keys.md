# Persistence keys and identifiers

Primary keys are always chosen by the application before insert (never database serials or identity columns). Integer keys are out of v1.

**Natural string PKs** where the protocol or Operator name is the identity and is not renamed: Operator-supplied `client_id` and Scope name. Both use the RFC 6749 §3.3 `scope-token` character class (`1*( %x21 / %x23-5B / %x5D-7E )`). Wire and storage compare with exact equality (scope values are case-sensitive per §3.3). At create, reject a new `client_id`, Scope name, or username that differs from an existing one only by ASCII case. Audience is an exact string on the Client row (not a separate catalog table). Username is unique and exact-matched for login but is not the User PK.

**Surrogate UUIDs:** UUIDv7 for internal non-bearer rows (Refresh Token Family, Audit Event, and similar). UUIDv4 for values exported as capability-adjacent identifiers — User PK / Access Token `sub`, and Access Token `jti`. Deactivate uses `deactivated_at timestamptz NULL` (NULL = active).

**Composite natural PKs:** Redirect URI registrations are `(client_id, redirect_uri)`. Consent Grants are `(user_id, client_id, scope)`; Deactivate sets `deactivated_at`; Reactivate clears it rather than inserting a new row; both mutations are Audit Events.

**High-entropy handles** (Authorization Codes, Refresh Tokens, Authorization Session cookie secrets): ≥128-bit CSPRNG opaque values; table PK / lookup key is the SHA-256 digest as `bytea` (ADR-0076). No per-token salt and no split `{id}.{secret}` for lookup.
