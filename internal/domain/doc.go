// Package domain is the hexagonal core: pure Authorization Server process
// logic and the ports (interfaces) that core needs.
//
// Ports live here — not in a separate port package — so the interface sits
// beside the code that depends on it (Go idiom: accept interfaces at the
// seam). Driven ports are things like persistence, clock, or entropy;
// driving ports are the use-case entrypoints adapters call into.
//
// Dependency rule: domain must not import internal/adapter/... . Adapters
// implement ports and call into domain. Domain is the sole enforcer of
// protocol and business rules and stays unit-testable without I/O (ADR-0033).
package domain
