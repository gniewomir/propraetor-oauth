# Refresh Tokens for all Public Clients (unified v1)

All Public Clients that complete the Authorization Code grant may receive Refresh Tokens. There is no separate browser vs native profile in v1: one policy for every Public Client. Refresh Tokens are opaque, returned in the Token Endpoint JSON response (RFC 6749), and presented on refresh via the `refresh_token` request parameter. Lifecycle rules are in ADR-0007; storage and transport hardening for browser runtimes are deferred to v2 crossroads (docs/crossroads/v2-public-client-token-handling.md).
