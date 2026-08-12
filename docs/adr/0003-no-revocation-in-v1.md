# No client-facing token revocation endpoint in v1

v1 does not implement RFC 7009. Access Tokens are JWTs and remain valid until exp for Resource Servers that do not consult the Authorization Server. Compensating controls are fail-closed Access Token TTL caps at start (ADR-0047), Authorization Code and Refresh Token TTL caps and ordering (ADR-0052), and Refresh Token lifecycle policy including Operator CLI invalidation (ADR-0007). Clients have no public revocation API until requirements name it.
