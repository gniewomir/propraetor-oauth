# JWKS at /.well-known/jwks.json

The JWKS document is served at `/.well-known/jwks.json`, i.e. `{Issuer URL}/.well-known/jwks.json` under the project Issuer URL convention (ADR-0008). This is not RFC 8414 authorization-server metadata (still a non-goal per ADR-0025).
