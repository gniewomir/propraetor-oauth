# Security and design gap grilling

Follow-on grilling closed underspecified security/design items. Locked in first pass: proxy shared secret (0046), Access Token TTL caps (0047), Authorization Session hardening (0048), admin trust root (0049), Refresh Tokens for all Public Clients (0050). Browser token transport alternatives: docs/crossroads/v2-public-client-token-handling.md.

Second pass (recommended + undocumented security gaps): code/refresh TTL caps (0052), Argon2id parameters (0053), HMAC-SHA256 for codes/refresh (0054), Consent Grants (0055), state bar (0056), PKCE verifier RFC 7636 (0057), authorization error redirects (0058), rate-limit ceilings (0059), unified purge CLI (0060), no account lockout (0061), password min length (0062), strict CSP (0063), session cookie `__Host-` (0064), RS verification contract (0065), exact native/loopback redirects (0066), production-risk ack flag (0067).
