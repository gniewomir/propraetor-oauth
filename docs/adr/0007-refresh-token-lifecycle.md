# Rotating opaque refresh tokens with CLI invalidation

Refresh Tokens are opaque, stored server-side, and rotated on every successful refresh. Reuse of a retired Refresh Token invalidates its Refresh Token Family (project security term). The Operator CLI can invalidate Refresh Tokens for a User (End-User) or Client. There is still no client-facing RFC 7009 revocation endpoint (ADR-0003); CLI invalidation is Operator-side control only, and Access Tokens remain valid until exp.
