# JWKS endpoint for Access Token verification

The AS exposes JWKS at `/.well-known/jwks.json` (ADR-0029). Resource Servers use the project convention **JWKS URL = `{Issuer URL}/.well-known/jwks.json`**. That is our URL layout rule, not RFC 8414 authorization-server metadata. Symmetric (HS256) was rejected as it forces every Resource Server to hold the signing secret. JWKS publishes the process’s current public key; graceful multi-kid rotation is out of v1 (ADR-0043). JWKS responses use `Cache-Control: public, max-age=300`; RS verification rules are ADR-0065.
