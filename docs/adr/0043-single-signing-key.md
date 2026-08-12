# Single ES256 key pair per process start

v1 loads one active ES256 key pair from PEM paths at server start and publishes it in JWKS. There is no multi-kid graceful rotation in-process; Operators rotate by replacing key material and restarting. A short Access Token TTL is the recommended compensating control during rotation (TTL values remain explicitly configured—ADR-0017).
