# Authorization Codes and Refresh Tokens stored as salted hashes

Authorization Codes and Refresh Tokens are stored at rest only as salted hashes, including algorithm metadata sufficient to verify later. The v1 algorithm is HMAC-SHA256 with per-token random salt (ADR-0054). Plaintext values are shown only at issuance (to the Client); the database never stores the raw secrets.
