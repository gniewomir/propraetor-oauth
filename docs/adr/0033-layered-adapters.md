# Layered adapters around the domain

The codebase separates HTTP adapters, CLI adapters, and persistence adapters from the domain. The domain is the sole enforcer of protocol and business rules and is fully unit-testable. Externally observable HTTP behavior is owned by end-to-end tests. CLI admin and server entrypoints are adapters, not alternate rule engines.
