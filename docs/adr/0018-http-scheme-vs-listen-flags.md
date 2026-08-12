# Separate flags for URL scheme vs cleartext listen

HTTP-related risk is split into two Operator-controlled allowances:

1. **Public URL scheme** — whether the AS may use `http://` in the Issuer URL and other public URLs (dev/cleartext-as-channel).
2. **Cleartext listen** — whether the AS may bind an unencrypted socket (typical behind a TLS-terminating proxy, or local cleartext).

Secure default: both denied. Enabling proxy-style cleartext listen does not by itself allow `http` public URLs; enabling `http` URLs does not by itself allow a cleartext bind. Combinations are explicit and e2e-tested. Enabling either requires `--i-understand-production-risk` (ADR-0067).
