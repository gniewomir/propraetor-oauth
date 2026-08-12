# PKCE S256 only

Authorization Requests using Authorization Code + PKCE accept `code_challenge_method=S256` only. The `plain` method is rejected. `code_verifier` length and charset follow RFC 7636 (ADR-0057).
