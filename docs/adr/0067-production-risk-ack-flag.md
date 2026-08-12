# Production-risk acknowledgment for dangerous listen flags

Enabling `--allow-http-issuer` or `--allow-cleartext-listen` (ADR-0018) requires a single companion flag `--i-understand-production-risk`. CLI help for each dangerous flag explains the risk (cleartext tokens/cookies; trusting the channel behind a proxy). One acknowledgment covers either or both allows in that start. There is no separate production profile and no magic environment detection—dev flexibility stays, accidental prod misconfig needs an explicit double opt-in.
