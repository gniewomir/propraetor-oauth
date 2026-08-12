# Authorization Session TTL is explicit CLI config

Authorization Session lifetime has no baked-in default. Server start requires an explicit CLI duration; expired Authorization Sessions are not usable. Removal of expired sessions is via Operator `cli purge` only (ADR-0060)—no automatic purge in `cli server`.
