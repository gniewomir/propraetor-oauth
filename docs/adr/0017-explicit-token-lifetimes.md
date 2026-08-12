# Token lifetimes must be set explicitly

Access Token, Refresh Token, and Authorization Code lifetimes have no baked-in defaults. The process requires explicit CLI configuration at every start. Fail-closed warn/refuse caps apply at start: Access Token (ADR-0047), Authorization Code and Refresh Token plus `code_ttl < access_ttl < refresh_ttl` (ADR-0052). Short Access Token TTL remains the primary compensating control for JWT-without-revocation and key-rotation-via-restart.
