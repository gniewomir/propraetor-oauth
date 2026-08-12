# Access Token TTL caps at start

Access Token lifetime is still explicitly configured (ADR-0017). At server start: if Access Token TTL is greater than 5 minutes, emit a warning; if greater than 15 minutes, refuse to start. This fail-closed bound backs JWT-without-revocation and key-rotation-via-restart.
