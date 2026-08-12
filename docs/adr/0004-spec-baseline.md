# Spec baseline without RFC 7009

v1 protocol baseline is RFC 6749, RFC 6750, RFC 6819, RFC 7636, and RFC 8252. State and PKCE are required on the Authorization Code Authorization Grant. RFC 7009 is out of baseline: there is no client-facing revocation endpoint (Operator CLI may still invalidate Refresh Tokens—ADR-0007).
