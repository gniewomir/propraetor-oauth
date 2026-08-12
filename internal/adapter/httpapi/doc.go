// Package httpapi is the driving adapter for the Authorization Server HTTP
// surface.
//
// Named httpapi (not http) so imports do not collide with net/http.
//
// It translates allowlisted HTTP (Authorization Endpoint, Token Endpoint,
// JWKS, Resource Owner login/Consent screens and static assets — ADR-0024,
// ADR-0035) into domain port calls. It does not own protocol or business
// rules (ADR-0033).
//
// Constructed by adapter/cli when the Operator runs `server`. Externally
// observable HTTP behavior is owned by end-to-end tests.
package httpapi
