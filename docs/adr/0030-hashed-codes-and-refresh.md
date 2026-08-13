# Authorization Codes and Refresh Tokens stored as hashes

Authorization Codes and Refresh Tokens are stored at rest only as SHA-256 digests of the opaque handle (`bytea` primary key); the database never stores the raw secrets. Plaintext values are shown only at issuance (to the Client). Policy and rationale for high-entropy handles vs Argon2id secrets are ADR-0076; key layout is ADR-0077.
