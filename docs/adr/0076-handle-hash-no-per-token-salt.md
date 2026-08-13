# High-entropy handles: SHA-256 at rest, no per-token salt

Authorization Codes, Refresh Tokens, and Authorization Session cookie secrets are high-entropy handles (≥128 bits CSPRNG), not password-class secrets. At rest the Authorization Server stores only a **SHA-256** digest of the full opaque value as the row primary key / lookup key (`bytea`) and verifies with constant-time compare of digests; plaintext is shown only at issuance (cookie or token response). This follows RFC 6819 §5.1.4.1.3 and §5.1.4.2.2: hash credentials (no cleartext); salt when entropy is low (passwords); rely on entropy for machine-generated handles. Per-token salt on these handles is out — it adds lookup complexity without material benefit for ≥128-bit secrets. Identifier and PK layout for these rows is ADR-0077.

Argon2id with salt remains mandatory for End-User passwords and Confidential Client secrets (ADR-0026, 0053).
