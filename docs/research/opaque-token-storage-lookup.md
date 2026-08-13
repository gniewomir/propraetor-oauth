# Opaque OAuth token / code / session storage and O(1) lookup

Research notes from **primary sources only** (IETF RFCs and BCPs, NIST / OWASP cheat sheets, official product docs, and published source from major Authorization Server projects).  
Retrieved / verified: 2026-08-13.

Scope: compare how a Postgres-backed Authorization Server can store and look up high-entropy opaque secrets — Authorization Codes, Refresh Tokens, and Authorization Session cookie ids — when:

- Plaintext must not be stored at rest
- Verification uses SHA-256 of the handle as PK/lookup ([ADR-0030](../adr/0030-hashed-codes-and-refresh.md), [ADR-0076](../adr/0076-handle-hash-no-per-token-salt.md), [ADR-0077](../adr/0077-persistence-keys.md))
- Token Endpoint / session load must be **O(1)**
- Refresh-token reuse detection may require retaining **retired** token rows

Focus: patterns that satisfy those constraints; comparative alternatives (including per-token salt / split tokens) remain below for design context.

---

## Executive summary / verdict

RFC 6819 requires **no cleartext storage** of credentials and **≥128 bits of entropy** for handle-based secrets; it recommends storing **hashes** (or encryption). Salt is called out mainly when entropy is low (passwords). The RFC does **not** prescribe a lookup schema. RFC 9700 requires public-client refresh tokens to be **sender-constrained or rotated**, and rotation explicitly **retains relationship information** so presenting a previously invalidated refresh token detects breach — retired rows must remain **addressable** by whatever lookup key the presented value implies.

**This repo (ADR-0076 / 0077):** approach **E** — SHA-256 of the full high-entropy handle as `bytea` PK; O(1) is `SHA256(presented)` → `SELECT`. Argon2id + salt remains for passwords and client secrets (ADR-0026/0053). Per-token salt on handles is out (RFC 6819: salt for low entropy). Comparative approaches below:

| Approach | Lookup key | Notes |
| --- | --- | --- |
| **A. Split** `{public_id}.{secret}` | stored `public_id` | Needed if verifier uses per-token salt |
| **B. Server-pepper deterministic index** | `HMAC(pepper, full_token)` | Extra server secret for lookup |
| **C. Salted hash only** | none | Not O(1) |
| **D. Plaintext handle as PK** | the secret itself | Cleartext — rejected |
| **E. SHA-256 of full token** | `SHA256(token)` | **Chosen** |
| **F. Fosite-style MAC** | signature half | Global MAC key model |

---

## Project constraints

| Constraint | Source |
| --- | --- |
| Codes & refresh: SHA-256 of handle at rest; plaintext only at issuance; no per-token salt | [ADR-0030](../adr/0030-hashed-codes-and-refresh.md), [ADR-0076](../adr/0076-handle-hash-no-per-token-salt.md) |
| Argon2id (+ salt) for passwords / client secrets only | [ADR-0026](../adr/0026-argon2id-secrets.md), [ADR-0053](../adr/0053-argon2id-parameters.md) |
| Deactivate non-TTL rows; hard-delete TTL + audit via purge only | [ADR-0075](../adr/0075-deactivate-and-purge.md) |
| Refresh opaque; rotated on use; family reuse detection | CONTEXT.md; [ADR-0007](../adr/0007-refresh-token-lifecycle.md) |
| Authorization Session: server-side Postgres state, opaque cookie | CONTEXT.md |
| No cleartext credential storage; hash or encrypt | [RFC 6819 §5.1.4.1.3](https://www.rfc-editor.org/rfc/rfc6819.html#section-5.1.4.1.3) |
| Handle entropy ≥128 bits | [RFC 6819 §5.1.4.2.2](https://www.rfc-editor.org/rfc/rfc6819.html#section-5.1.4.2.2) |
| Public-client refresh: sender-constrain **or** rotate + retain relationship | [RFC 9700 §4.14.2](https://www.rfc-editor.org/rfc/rfc9700.html) |

---

## Threat model the storage choice must answer

From RFC 6819 (AS database disclosure of codes / refresh / access-token handles; online guessing):

- **DB read compromise** should not yield usable bearer secrets. Countermeasure: store hashes only ([§5.1.4.1.3](https://www.rfc-editor.org/rfc/rfc6819.html#section-5.1.4.1.3); authorization-code DB threat errata: store **authorization code** hashes — [Errata 5965](https://www.rfc-editor.org/errata/eid5965)).
- **Online guessing** of handles must be infeasible → entropy ≥128 bits ([§5.1.4.2.2](https://www.rfc-editor.org/rfc/rfc6819.html#section-5.1.4.2.2); randomness BCP [RFC 4086](https://www.rfc-editor.org/rfc/rfc4086.html)).
- **Refresh theft / replay** for public clients: rotation + retained relationship so a presented **retired** token detects breach ([RFC 9700 §4.14.2](https://www.rfc-editor.org/rfc/rfc9700.html)).

OAuth 2.1 (draft) keeps refresh tokens **opaque to the client** and allows either a retrieval identifier or encoded information ([draft-ietf-oauth-v2-1](https://datatracker.ietf.org/doc/html/draft-ietf-oauth-v2-1)); it does not specify AS at-rest hashing.

---

## Cross-cutting notes

### Per-token random salt vs server-wide pepper

| Mechanism | Role | Lookup implication |
| --- | --- | --- |
| **Per-token salt** (per-token salt approach) | Stored with the digest; defeats precomputed tables across rows; NIST requires salt for **memorized** secrets ([SP 800-63B](https://pages.nist.gov/800-63-3/sp800-63b.html)) | Digest is **not** recomputable without fetching the row → needs a **separate** lookup key |
| **Server pepper / HMAC key** | Secret known only to the AS (optionally HSM); NIST’s optional “secret salt” for passwords is the same idea | Enables **deterministic** index `HMAC(pepper, token)` without storing plaintext; rotation of pepper must be planned |
| **Both** | Possible but usually redundant for ≥128-bit handles | Extra columns / ops cost |

RFC 6819 §5.1.4.1.3: “If the credential lacks a reasonable entropy level (because it is a user password), an additional salt will harden the storage….” High-entropy opaque tokens are the other case: fast keyed or unkeyed hashes are cryptographically normal (this project already rejects slow KDFs for Token Endpoint predictability).

### Access Token `jti` is a different case

When Access Tokens are JWTs verified by Resource Servers via JWKS (this product’s model — CONTEXT.md), authentication does **not** require looking up the opaque bearer string in Postgres. RFC 7519’s optional `jti` is a unique JWT id for replay prevention / correlation ([RFC 7519 §4.1.7](https://www.rfc-editor.org/rfc/rfc7519.html#section-4.1.7)). A denylist keyed by `jti` (if ever used) stores an **identifier**, not a vaulted secret. Opaque storage research applies to **codes, refresh tokens, and server-side sessions**, not to JWT validation of Access Tokens.

### UUID / public_id entropy if the id is exposed

In split pattern A, `public_id` is a **locator**, not the authenticator:

- Security rests on `secret` entropy (≥128 bits per RFC 6819 §5.1.4.2.2).
- `public_id` must be unique and non-guessable enough to avoid trivial enumeration of rows if the API leaks existence differences; UUIDv4 (122 bits random) is ample for uniqueness and practical unguessability of the id itself, but **does not** replace secret entropy.
- OWASP: session ids ≥128 bits from a CSPRNG ([Session Management Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Session_Management_Cheat_Sheet.html)). NIST session secrets: ≥64 bits from an approved RBG ([SP 800-63-4 session](https://pages.nist.gov/800-63-4/sp800-63b/session/)); OAuth handle guidance is stricter (≥128 bits).
- Do not log full tokens; OWASP recommends logging a **salted hash** of a session id for correlation, not the raw id ([Session Management Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Session_Management_Cheat_Sheet.html)). Logging `public_id` alone is usually acceptable if the secret half never appears in logs.

### Refresh reuse detection and retired rows

RFC 9700 §4.14.2: on rotation, invalidate the previous refresh token but **retain information about the relationship**; reuse of an invalidated token informs the AS of breach and leads to revoking the active refresh token (Auth0 documents the same family-wide invalidation — [Refresh Token Rotation](https://auth0.com/docs/secure/tokens/refresh-tokens/refresh-token-rotation)). Storage implication: the lookup key for a **retired** presentation must still find a row (or family metadata). Split `public_id`, SHA-256(handle), and Fosite signature all support this; deleting the row on rotate without a side index breaks reuse detection.

---

## Approaches

### A. Split opaque token: `{lookup_id}.{secret}`

**How lookup works**

1. At issuance, generate `lookup_id` (e.g. UUIDv4) and high-entropy `secret` (≥128 bits CSPRNG).
2. Issue `lookup_id || "." || secret` (encoding as needed) to the Client or set it as the session cookie value.
3. Persist a row with PK/UNIQUE on `lookup_id`; store salt + HMAC-SHA256 digest of `secret` (per-token salt approach); never store `secret`.
4. On present: split on separator → `SELECT … WHERE lookup_id = $1` (index / PK) → O(1).

**How verification works**

Load salt + algorithm metadata + digest; recompute HMAC over presented `secret`; constant-time compare. Invalid format, missing row, and bad secret all map to the same protocol failure (`invalid_grant` / unauthenticated session).

**What it buys**

- O(1) lookup **without** a server pepper for the index.
- Full compatibility with **per-token random salt**.
- DB leak exposes `lookup_id` + salt + digest, not the bearer secret (RFC 6819 “hashes only”).
- App-known `lookup_id` before insert → easy fixtures / testability.
- Retired refresh rows remain addressable by the same `lookup_id` for reuse detection.

**What it costs**

- Wire format parsing; separator collision rules; slightly longer presented values.
- `lookup_id` is not confidential — security is entirely on `secret` entropy.
- Index stores public ids (size ≈ UUID / opaque id width); digest column separate.
- Pepper rotation is N/A for lookup; salt rotation is per-row at reissue only.
- Compatible with a per-token-salt verifier.

**Primary sources:** RFC 6819 §5.1.4.1.3 / §5.1.4.2.2; RFC 9700 §4.14.2; ADR-0030 / 0076. Closest shipping relative: Fosite’s two-part token (see F) uses a MAC half as lookup rather than a public id.

---

### B. Full-token keyed hash as lookup key (server pepper) ± per-token salt verifier

**How lookup works**

1. Issue a single opaque random token `T` (≥128 bits).
2. Store `lookup = HMAC(server_pepper, T)` (or keyed KDF) as UNIQUE/PK.
3. Optionally also store per-token salt + verifier over `T` (usually redundant if the lookup MAC already binds `T`).
4. On present: recompute `lookup` from `T` → `SELECT` by that key → O(1).

**How verification works**

Finding the row via keyed hash already proves knowledge of `T` **and** possession of the pepper at computation time. A second salted verifier is optional double-checking.

**What it buys**

- Single-string token (no split UX).
- O(1) deterministic index.
- DB dump **without** pepper cannot confirm whether a stolen candidate `T` is still valid (stronger than unsalted SHA-256 membership tests).
- Aligns with Hydra’s stated goal that DB access alone should not suffice without `secrets.system` ([ory/hydra#1831](https://github.com/ory/hydra/issues/1831)).
- NIST optional secret salt / pepper pattern for stored secrets ([SP 800-63B](https://pages.nist.gov/800-63-3/sp800-63b.html)).

**What it costs**

- Server secret lifecycle: generate, store, rotate, agree across instances (Fosite `Validate` walks current + rotated global secrets — [`token/hmac/hmacsha.go`](https://github.com/ory/fosite/blob/master/token/hmac/hmacsha.go)).
- Keeping per-token salt **and** peppered lookup → two digests for one secret (complexity; little gain for high-entropy `T`).
- Tests need the pepper (or double) to predict keys.
- Index size ≈ digest width (e.g. 32 bytes hex/base64).

**Primary sources:** Fosite HMAC strategy; Hydra#1831; NIST SP 800-63B (secret salt).

---

### C. Store only salted hash with no separate lookup key

**How lookup works**

Store `salt` + `HMAC(salt, T)` only. On present, there is no key derivable from `T` alone that points at the row.

**How verification works**

Only by scanning candidate rows (try each salt) or by abandoning random salt (deterministic salt from `T`, which defeats the anti-rainbow purpose of per-token salt).

**What it buys**

Nothing useful for Token Endpoint / session load under O(1) with per-token salt.

**What it costs**

Full-table verify (latency, DoS surface) or workaround that recreates A/B/E:

| Workaround | Effect |
| --- | --- |
| Put salt (or encrypted salt) in the token | Becomes split / self-describing → essentially A or encrypted-id variant |
| Derive salt from `T` | Deterministic → collapses toward E/B |
| Encrypt `T` at rest with server key as PK | Cleartext-equivalent for anyone with the key; different threat model |

**Why it fails for this project:** per-token **random** salt makes the digest non-indexable by the presented secret. This is the core reason “hash like a password” does not transfer to **bearer handles that must be looked up**.

**Primary sources:** logical consequence of RFC 6819’s hash-at-rest guidance + random salt (NIST salt requirements for memorized secrets); no RFC recommends salted-hash-only lookup for OAuth handles.

---

### D. Store plaintext token / session id as PK

**How lookup works**

Presented value is the primary key; `SELECT` by equality.

**How verification works**

Row presence (+ expiry / binding checks). No hash compare.

**What it buys**

- Simplest code; trivial O(1); trivial reuse-detection rows.
- Common historically for short-lived **server sessions** when operators accept “DB read = session theft until expiry.”

**What it costs**

- Conflicts with RFC 6819 §5.1.4.1.3 and ADR-0030 for codes/refresh.
- DB compromise = immediate theft of all valid handles.
- Keycloak-style designs often keep **session ids** as first-class keys while refresh/access are JWTs bound to `sid` (session is the vault; token is signed assertion) — different architecture than opaque hashed refresh ([Keycloak Server Admin — sessions / offline](https://www.keycloak.org/docs/latest/server_admin/#_offline-access)).

**When people accept this**

- Very short TTL; DB encrypted at rest + strict host controls; session table treated as equivalent to memory.
- Still a conscious downgrade vs hashing for long-lived refresh tokens.

**Fit:** Not for Authorization Codes / Refresh Tokens here. Session cookies: only with explicit risk acceptance; A/B/E remain consistent if one policy is preferred.

**Primary sources:** RFC 6819 §5.1.4.1.3; Keycloak admin guide (session-centric refresh); OWASP session storage guidance (secure repository; prefer not exposing raw ids in logs).

---

### E. Hash the full token with a publicly known algorithm (no server secret)

**How lookup works**

1. Issue high-entropy opaque `T`.
2. Store `key = SHA256(T)` (often with a grant-type domain separator) as PK.
3. On present: hash `T` → `SELECT` by `key`.

**How verification works**

Row found via preimage-resistant hash of the full secret = proof of possession for high-entropy `T`.

**What it buys**

- Simple single-string token; O(1); no split format; no per-token salt column.
- High-entropy `T` is not practically invertible (preimage resistance).
- Matches Duende IdentityServer: handle never stored; `Key = SHA256(handle + grant-type metadata)`; stated intent that reading keys does not recover handles ([Reference Tokens](https://docs.duendesoftware.com/identityserver/tokens/reference/), [Persisted Grant Store](https://docs.duendesoftware.com/identityserver/reference/stores/persisted-grant-store/)). ConsumedTime retained for one-time / replay semantics.
- Cloudflare workers-oauth-provider documents SHA-256 hashes of codes / access / refresh at rest ([storage-schema.md](https://github.com/cloudflare/workers-oauth-provider/blob/main/storage-schema.md)).

**What it costs**

- **Rainbow / offline membership:** with only a public hash, an attacker who steals a candidate `T` (logs, device) and a DB dump can test whether that exact token is still valid. With a pepper (B), that check needs the server secret. For ≥128-bit random `T`, building rainbow tables over the full space is infeasible — the realistic residual is **membership testing of stolen candidates**, not offline guessing from the hash alone.
- **Conflicts with a per-token-salt per-token salt** unless the ADR is revised (RFC 6819 already says salt matters mainly for low entropy).
- No integrity MAC on the token string itself (unlike Fosite’s signature half).

**Primary sources:** Duende docs; Cloudflare schema; RFC 6819 §5.1.4.1.3 + §5.1.4.2.2.

---

### F. Notable patterns from major identity servers / RFCs

#### RFC / BCP layer (policy, not schema)

| Source | Storage-relevant point |
| --- | --- |
| [RFC 6749](https://www.rfc-editor.org/rfc/rfc6749.html) | Refresh / codes are opaque strings; refresh may be identifier or encoded data; no at-rest hash schema |
| [RFC 6819](https://www.rfc-editor.org/rfc/rfc6819.html) | No cleartext; hash/encrypt; ≥128-bit handles; DB disclosure threats for codes/refresh/access handles |
| [RFC 9700](https://www.rfc-editor.org/rfc/rfc9700.html) | Public-client refresh: sender-constrain or rotate; retain relationship for reuse detection |
| [OAuth 2.1 draft](https://datatracker.ietf.org/doc/html/draft-ietf-oauth-v2-1) | Opaque refresh; inherits BCP security posture |
| [RFC 7519 `jti`](https://www.rfc-editor.org/rfc/rfc7519.html#section-4.1.7) | JWT id for uniqueness / replay — not opaque secret vaulting |

#### Ory Fosite / Hydra (split + global HMAC; DB by signature)

From [`token/hmac/hmacsha.go`](https://github.com/ory/fosite/blob/master/token/hmac/hmacsha.go) and README:

1. Random `tokenKey` (≥32 bytes; comments cite RFC 6819 §5.1.4.2.2).
2. `signature = HMAC-SHA512/256(globalSecret, tokenKey)`.
3. Issue `base64url(tokenKey) + "." + base64url(signature)`.
4. Persist/lookup by `signature`. Hydra additionally stores `SHA-384(signature)` as the access-token PK when the raw signature would overflow / for uniformity — [`x/sighash.go`](https://raw.githubusercontent.com/ory/hydra/master/x/sighash.go), [`persister_oauth2.go`](https://github.com/ory/hydra/blob/master/persistence/sql/persister_oauth2.go).
5. Validate: recompute HMAC with current/rotated secrets; then DB get.

This is **not** per-token salt: the MAC key is a **server-wide secret**; the “signature” half is both integrity tag and (typically) lookup key. Closest research classification: **hybrid of A (two-part token) and B (keyed material)**.

#### Duende IdentityServer (E)

Opaque handle (32 bytes CSPRNG + version suffix); store SHA-256 of handle∥grant-type; encrypt grant `Data` with ASP.NET Data Protection by default ([Persisted Grant Store](https://docs.duendesoftware.com/identityserver/reference/stores/persisted-grant-store/)).

#### Auth0 (managed rotation; storage opaque to customers)

Documents refresh rotation and **automatic reuse detection** / token families ([Refresh Token Rotation](https://auth0.com/docs/secure/tokens/refresh-tokens/refresh-token-rotation)); does not publish at-rest hash schema. Useful as **behavioral** primary source for RFC 9700-aligned reuse detection, not as a storage blueprint.

#### Keycloak (session-centric JWTs)

Refresh/offline tokens are JWTs bound to user/client sessions; “Revoke Refresh Token” issues a new token and invalidates the prior one ([Server Administration Guide](https://www.keycloak.org/docs/latest/server_admin/#_offline-access)). Persistence centers on **session ids** and session data, not hashed opaque refresh strings — a different decomposition than handle-hash AS designs.

---

## Comparison matrix

| | O(1) | No plaintext | Works with per-token salt (per-token salt approach) | Needs server secret for lookup | DB leak without server secret | Format / ops complexity |
| --- | --- | --- | --- | --- | --- | --- |
| **A Split id.secret** | Yes | Yes (secret hashed) | **Yes** | No | ids + salts + digests; still need `secret` | Medium |
| **B Peppered index** | Yes | Yes | Optional / redundant | **Yes** | Cannot forge / membership-test without pepper | Medium (secret ops) |
| **C Salt only** | **No** | Yes | Yes | No | Unusable design | Low (broken) |
| **D Plaintext PK** | Yes | **No** | N/A | No | **Full token theft** | Lowest |
| **E SHA256(T) PK** | Yes | Yes | **No** (replaces it) | No | Preimage hard; membership test on stolen `T` | Low |
| **F Fosite MAC** | Yes | Yes (sig / hash of sig) | **No** (global MAC) | **Yes** | Designed so DB alone ≠ usable tokens | Medium |

---

## What actually drives the choice here

1. **Per-token salt (per-token salt approach)** is the fork. Right tool for **low-entropy** secrets (RFC 6819 §5.1.4.1.3; NIST memorized secrets). For **≥128-bit handles**, unsalted or peppered full-token hashes (E/B/F) are normal in shipping AS code; keeping random salt then **requires** A or B’s separate lookup key.
2. **Reuse detection** (RFC 9700 / Auth0 behavior / ADR-0007) needs a stable lookup key for **retired** rows — `public_id`, `SHA256(T)`, or Fosite `signature` all work.
3. **Session cookies** share the design space; short TTL may tempt D, but A/B/E keep one policy with codes/refresh.
4. **JWT Access Tokens** do not need this vault for RS auth; `jti` is not a substitute opaque refresh design.

---

## Recommendation for propraetor-oauth

**Decided (ADR-0076 / 0077):** **E** — `SHA-256(full opaque handle)` as `bytea` PK; ≥128-bit CSPRNG; Argon2id for passwords and client secrets only (ADR-0026 / 0053). Deactivate / purge is ADR-0075; CSRF derivation is ADR-0078 (orthogonal to handle storage).

---

## Primary source index

| Topic | Source |
| --- | --- |
| No cleartext credentials; hash or encrypt; salt for low entropy | [RFC 6819 §5.1.4.1.3](https://www.rfc-editor.org/rfc/rfc6819.html#section-5.1.4.1.3) |
| Handle entropy ≥128 bits | [RFC 6819 §5.1.4.2.2](https://www.rfc-editor.org/rfc/rfc6819.html#section-5.1.4.2.2) |
| Authorization code DB: store code hashes (errata) | [RFC 6819 Errata 5965](https://www.rfc-editor.org/errata/eid5965) |
| Randomness BCP | [RFC 4086](https://www.rfc-editor.org/rfc/rfc4086.html) |
| Refresh rotation; retain relationship; reuse detection | [RFC 9700 §4.14](https://www.rfc-editor.org/rfc/rfc9700.html) |
| Opaque refresh / code framework | [RFC 6749](https://www.rfc-editor.org/rfc/rfc6749.html), [OAuth 2.1 draft](https://datatracker.ietf.org/doc/html/draft-ietf-oauth-v2-1) |
| JWT `jti` | [RFC 7519 §4.1.7](https://www.rfc-editor.org/rfc/rfc7519.html#section-4.1.7) |
| Memorized-secret hashing, salt, optional secret salt | [NIST SP 800-63B](https://pages.nist.gov/800-63-3/sp800-63b.html) |
| Session secret length / opaque session id | [NIST SP 800-63-4 session](https://pages.nist.gov/800-63-4/sp800-63b/session/) |
| Session id entropy; log salted hash not raw id | [OWASP Session Management Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Session_Management_Cheat_Sheet.html) |
| Fosite `key.signature` HMAC strategy | [fosite `token/hmac/hmacsha.go`](https://github.com/ory/fosite/blob/master/token/hmac/hmacsha.go) |
| Fosite README token layout | [ory/fosite README](https://github.com/ory/fosite/blob/master/README.md) |
| Hydra signature hash at rest | [hydra `x/sighash.go`](https://raw.githubusercontent.com/ory/hydra/master/x/sighash.go), [persister_oauth2.go](https://github.com/ory/hydra/blob/master/persistence/sql/persister_oauth2.go) |
| Hydra DB vs `secrets.system` rationale | [ory/hydra#1831](https://github.com/ory/hydra/issues/1831) |
| Duende handle vs hashed key | [Reference Tokens](https://docs.duendesoftware.com/identityserver/tokens/reference/), [Persisted Grant Store](https://docs.duendesoftware.com/identityserver/reference/stores/persisted-grant-store/) |
| Auth0 rotation + reuse detection | [Refresh Token Rotation](https://auth0.com/docs/secure/tokens/refresh-tokens/refresh-token-rotation) |
| Keycloak revoke-refresh / offline sessions | [Keycloak Server Admin](https://www.keycloak.org/docs/latest/server_admin/#_offline-access) |
| SHA-256 hashes of codes/tokens at rest | [cloudflare/workers-oauth-provider `storage-schema.md`](https://github.com/cloudflare/workers-oauth-provider/blob/main/storage-schema.md) |
| Project ADRs | [0030](../adr/0030-hashed-codes-and-refresh.md), [0076](../adr/0076-handle-hash-no-per-token-salt.md), [0077](../adr/0077-persistence-keys.md), [0075](../adr/0075-deactivate-and-purge.md), [0078](../adr/0078-csrf-derived-from-signing-key.md), [0007](../adr/0007-refresh-token-lifecycle.md) |
