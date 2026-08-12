# PKCE code_verifier must meet RFC 7636

Beyond `code_challenge_method=S256` only (ADR-0023), the AS enforces RFC 7636 `code_verifier` length and charset: 43–128 characters from the unreserved set. Checked at the Authorization Request (with the challenge) and again at the Token Endpoint when verifying the verifier. Weak or non-interoperable verifiers are rejected.
