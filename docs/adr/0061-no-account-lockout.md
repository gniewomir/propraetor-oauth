# No End-User account lockout in v1

Login abuse control in v1 is rate limiting only (per IP and per username, ADR-0040/0059). There is no soft or hard account lockout and no progressive delay. Lockout was rejected for v1 because it adds Operator unlock burden and enables denial-of-service against End-Users; rate limits at the AS boundary are the conscious compensating control. Revisit if credential stuffing exceeds what ceilings allow.
