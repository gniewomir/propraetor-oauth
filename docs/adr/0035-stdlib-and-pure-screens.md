# Stdlib-first; screens have no external deps

Prefer the Go standard library. External modules are allowed only when necessary (e.g. Argon2id, Postgres driver, JWT/ES256 support). Resource Owner–facing screens use pure CSS/JS with no external dependencies or CDNs. Strict CSP on those screens is ADR-0063.
