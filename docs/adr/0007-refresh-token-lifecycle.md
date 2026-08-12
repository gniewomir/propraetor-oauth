# Rotating opaque refresh tokens; concurrent families; Not-Before

Refresh Tokens are opaque, stored server-side, and rotated on every successful refresh. A **Refresh Token Family** is the lineage from **one Authorization Code redemption** (CONTEXT.md): the same User×Client may have **multiple concurrent families** (e.g. the same SPA in two browsers). Rotation retires only the previous token in that family; it does not invalidate other families. Reuse of a retired member invalidates that entire family (theft signal).

Operator-side invalidation of outstanding tokens/sessions for a User or Client is via **Not-Before** (ADR-0069), not per-token revoke. Not-Before does not skip reuse detection: if a presented token authenticates as a retired family member, family-wide invalidation still runs; unknown/garbage tokens stay generic invalid. There is still no client-facing RFC 7009 revocation endpoint (ADR-0003); Access Tokens remain valid at Resource Servers until exp.
