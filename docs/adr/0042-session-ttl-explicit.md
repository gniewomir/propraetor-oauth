# Authorization Session TTL is explicit Server Policy

Authorization Session lifetime has no baked-in default. Server start requires an explicit duration in Server Policy (ADR-0068); expired Authorization Sessions are not usable. Removal of expired sessions is via Operator `cli purge` only (ADR-0060)—no automatic purge in `cli server`.
