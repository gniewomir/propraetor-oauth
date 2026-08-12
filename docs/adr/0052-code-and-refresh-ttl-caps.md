# Authorization Code and Refresh Token TTL caps

Authorization Code and Refresh Token lifetimes remain explicit CLI config (ADR-0017). At server start: Authorization Code warn above 2 minutes and refuse above 10 minutes; Refresh Token warn above 30 days and refuse above 90 days. Refuse start unless `code_ttl < access_ttl < refresh_ttl` (strict). Access Token caps stay ADR-0047. Fail-closed bounds stop nonsensical or overly long grants alongside JWT-without-revocation.
