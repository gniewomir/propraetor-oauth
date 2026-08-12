// Package cli is the driving adapter for the Operator CLI and the process
// composition root.
//
// It implements the inbound edge: parse commands, wire concrete driven
// adapters (e.g. postgres) to domain ports, invoke domain use cases for
// admin, and for `server` construct the httpapi driving adapter
// (ADR-0019, ADR-0022, ADR-0033). Admin does not require a running server;
// it still goes CLI → domain → persistence.
//
// Not Resource Owner UI — login/Consent screens live in adapter/httpapi.
package cli
