# HMAC-SHA256 for Authorization Codes and Refresh Tokens at rest

Authorization Codes and Refresh Tokens are stored as HMAC-SHA256 digests with a per-token random salt and verified with constant-time comparison. Algorithm and salt metadata are stored with the hash (ADR-0030). Argon2id is reserved for passwords and client secrets (ADR-0026, 0053): high-entropy opaque tokens do not need a slow KDF, and Token Endpoint load must stay predictable.
