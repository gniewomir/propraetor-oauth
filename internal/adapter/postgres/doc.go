// Package postgres is a driven adapter: Postgres implementations of domain
// persistence ports (ADR-0006, ADR-0033).
//
// It satisfies seams defined in domain (stores for Clients, codes, refresh
// tokens, sessions, audit, etc.). No protocol rules and no HTTP/CLI concerns
// belong here. Add sibling driven adapters under internal/adapter/ as other
// I/O seams become real (one adapter is hypothetical; two make a seam).
package postgres
