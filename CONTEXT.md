# Propraetor OAuth

A single-tenant, first-party Authorization Server. It issues OAuth 2.0 tokens for provisioned Clients and publishes JWKS so Resource Servers can verify Access Tokens; it does not host Resource Servers. v1 has no RFC 7009 revocation endpoint.

## Language

### OAuth roles (RFC 6749)

**Resource Owner**:
An entity capable of granting access to a Protected Resource. When the Resource Owner is a person, it is an End-User. In v1 interactive flows, that is the human who authorizes a Public Client after password login.
_Avoid_: account, customer; do not imply a Resource Owner participates in Client Credentials

**End-User**:
RFC term for a Resource Owner who is a person.
_Avoid_: using End-User for Operator or Client

**User**:
CLI/provisioning synonym for an End-User (Resource Owner). Use **Resource Owner** or **End-User** in protocol and HTTP-facing docs.
_Avoid_: using User for Operator or Client

**Client**:
An application that makes Protected Resource requests. Provisioned by the Operator; first-party only in v1.
_Avoid_: app, consumer, customer, user

**Public Client**:
A Client unable to maintain the confidentiality of credentials. Uses the Authorization Code grant with PKCE. All Public Clients may receive Refresh Tokens in v1 (ADR-0050); v2 browser hardening crossroads A, B, D only (ADR-0051, docs/crossroads/v2-public-client-token-handling.md).
_Avoid_: mobile app, SPA (unless a concrete instance); browser profile (until v2)

**Confidential Client**:
A Client able to maintain the confidentiality of credentials. Uses the Client Credentials grant in v1.
_Avoid_: server client, M2M client (informal gloss only)

**Authorization Server (AS)**:
This product — the server that issues Access Tokens to a Client after authenticating the Client and, when the grant requires it, authenticating the Resource Owner and obtaining authorization.
_Avoid_: IdP, identity platform, auth service (too broad)

**Resource Server (RS)**:
The server hosting Protected Resources, capable of accepting Protected Resource requests using Access Tokens. External to this product.
_Avoid_: guarded resource (as something implemented here), API gateway

**Protected Resource**:
An access-restricted resource hosted by a Resource Server.
_Avoid_: guarded resource, API (as the OAuth term)

**Sample Resource Server**:
A minimal Resource Server used only in end-to-end tests to prove issued Access Tokens are usable. Not a product feature.
_Avoid_: reference API, demo API (as product scope)

### Grants and credentials

**Authorization Grant**:
A credential representing the Resource Owner’s authorization (or, for Client Credentials, the Client’s own authority) used to obtain an Access Token — in v1: Authorization Code, Client Credentials, or Refresh Token.
_Avoid_: permission, license

**Authorization Code**:
A one-time, short-lived Authorization Grant issued after Resource Owner authorization, exchanged at the Token Endpoint with PKCE. Only a salted hash plus algorithm metadata is stored at rest.
_Avoid_: auth code (informal ok), grant code

**Access Token**:
A credential used to access Protected Resources. In this project: a JWT Bearer Token presented to a Resource Server.
_Avoid_: API key, session token

**Refresh Token**:
A credential used to obtain new Access Tokens from the Token Endpoint. Issued to all Public Clients on the Authorization Code grant in v1. Opaque; rotated on use; returned in the Token Endpoint JSON response and sent on refresh via the `refresh_token` parameter (RFC 6749). Only a salted hash plus algorithm metadata is stored at rest.
_Avoid_: calling this an Access Token; JWT refresh; HttpOnly cookie transport (v2 crossroad only)

**Refresh Token Family**:
Project security term: the lineage of Refresh Tokens from successive rotations after one Authorization Code redemption. The same User×Client may have multiple concurrent families. Reuse of a retired member invalidates that entire family only. Not an RFC 6749 term.
_Avoid_: session (overloaded), token chain; one family per User×Client

**Bearer Token**:
An Access Token presented using the HTTP Authorization scheme defined by RFC 6750. This project accepts `Authorization: Bearer` only.
_Avoid_: token in query string, form-body bearer

**Client Identifier**:
The unique string assigned to each Client (`client_id`).
_Avoid_: client name (as the identifier), app id

**Client Authentication**:
How a Confidential Client proves its identity at the Token Endpoint. In v1: `client_secret_basic` only. Public Clients use none (PKCE instead).
_Avoid_: login (that is Resource Owner authentication)

### Endpoints and requests

**Authorization Endpoint**:
The AS HTTP endpoint used by the Client to obtain Resource Owner authorization (Authorization Code flow).
_Avoid_: login URL (login is a step behind this endpoint)

**Token Endpoint**:
The AS HTTP endpoint used by the Client to exchange an Authorization Grant for an Access Token (and optionally a Refresh Token).
_Avoid_: token API, auth API

**Authorization Request**:
The Client’s request to the Authorization Endpoint to start Authorization Code + PKCE (includes `client_id`, `redirect_uri`, `scope`, `state`, PKCE parameters, etc.).
_Avoid_: login request, auth request (ambiguous)

**Redirect URI**:
A pre-registered absolute redirection URI where the AS may send the Authorization Code. Matched by exact string equality only. Custom URI schemes and explicit loopback URIs may be registered; wildcards (including localhost port wildcards) are not allowed.
_Avoid_: callback URL (informal ok), wildcard redirect

### Scope and Resource Owner authorization

**Scope**:
A space-delimited access-right value from an Operator-managed catalog stored in the database. The Operator CLI defines and deletes Scopes and assigns them to Clients; Access Tokens carry granted Scopes.
_Avoid_: free-form scope strings; “fixed” catalog (implies compile-time); role (unless later aliased deliberately)

**Consent**:
The AS UI step that obtains Resource Owner authorization for Scopes a Public Client has not already been granted by that Resource Owner. Maps to “obtaining authorization” in RFC 6749; not a separate RFC term.
_Avoid_: permission dialog, OAuth prompt (as the domain term); do not treat Consent as required for Client Credentials

**Consent Grant**:
A persisted approval that an End-User has authorized a specific Scope for a specific Public Client. Stored in Postgres; drives incremental Consent (ADR-0011, 0055). Invalidated when the Operator removes that Scope from the Client allowlist or revokes via CLI.
_Avoid_: permission, entitlement, OAuth grant (ambiguous with Authorization Grant)

### JWT token profile (not RFC 6749 roles)

**Issuer**:
The single logical authority this deployment represents in issued JWTs (`iss`) and related configuration. One tenant, one Issuer.
_Avoid_: tenant, realm, organization; do not treat Issuer as an RFC 6749 role

**Issuer URL**:
The absolute base URL identifying the Issuer in Access Token `iss` claims and JWKS-related links. Scheme (`http` vs `https`) is controlled separately from whether the process listens without TLS.
_Avoid_: base URL, public URL (as the term for Issuer URL)

**Audience**:
The intended Resource Server identifier in an Access Token’s `aud` claim. Each Client has exactly one registered Audience. JWT-profile term, not an RFC 6749 role.
_Avoid_: using Issuer as Audience; per-request resource indicators (out of v1)

**JWKS**:
The AS HTTP document of public keys (RFC 7517) Resource Servers use to verify Access Token signatures, served at `/.well-known/jwks.json`. Project convention: JWKS URL = `{Issuer URL}/.well-known/jwks.json` (layout rule, not RFC 8414 discovery).
_Avoid_: public key file distribution (as the v1 mechanism); calling this authorization-server metadata

### Project operations (not OAuth RFC terms)

**Operator**:
The human who runs the AS and provisions Clients, Users, and Scopes via the CLI.
_Avoid_: admin user (ambiguous with User), tenant admin

**CLI**:
The single Operator-facing binary. Subcommands select mode: **server** runs the AS process; other commands perform administration. Admin and server entrypoints are adapters over the domain (CLI → domain → persistence), not a second rules engine and not ad-hoc SQL.
_Avoid_: admin API, control plane (as separate products)

**Server Policy**:
The Operator-authored file of Access Token / Refresh Token / Authorization Code / Authorization Session lifetimes and rate-limit thresholds and windows. Required to start **server** mode; not used by admin subcommands. Its shape is the Policy schema.
_Avoid_: config file (too generic); policy (alone — overloaded with Consent and Authorization Grant)

**Policy schema**:
The closed set of keys a Server Policy file must contain; unknown keys are rejected. Distinct from the Storage schema.
_Avoid_: bare schema; closed schema (as the name — prefer Policy schema)

**Storage schema**:
The versioned Postgres structure the Authorization Server expects for persisted state. Advanced by Operator CLI storage migrate commands. Distinct from the Policy schema.
_Avoid_: bare schema; calling this Server Policy

**Authorization Session**:
Server-side state (stored in Postgres) that ties an in-progress Authorization Request across login and Consent, referenced by an opaque cookie. Hardened per ADR-0048 and ADR-0064 (HttpOnly/Secure/`__Host-` when https, SameSite, CSRF, fixation rotation, bound request fields).
_Avoid_: browser session (vague), JWT session; not an RFC 6749 term

**Not-Before**:
An Operator-set instant on a User or on a Client. At the Authorization Server, all tokens and Authorization Sessions created before that instant are rejected on use; already-issued Access Tokens at Resource Servers remain valid until exp. Framing is “all tokens” so future ID tokens share the same rule when OIDC is introduced.
_Avoid_: revoke-tokens, token revocation (as the Operator action; RFC 7009 is separate), per-token invalidation (as the primary Operator model)

**Audit Event**:
A recorded security-relevant fact (e.g. login failure, token issue, admin mutation, rate-limit trip). Persisted in Postgres and emitted as structured stdout/stderr. May reference related entity ids; never stores raw secrets.
_Avoid_: access log, debug log (as the audit trail)
