# Resource Owner login is password-based

End-Users (Resource Owners) authenticate on AS-hosted login screens with username and password only. Passwords are hashed at rest (ADR-0026, 0053). Minimum password length is ADR-0062. Federation and external identity providers are out of v1 scope. Resource Owner–facing screens use pure CSS/JS with no external dependencies (ADR-0035) and strict CSP (ADR-0063).
