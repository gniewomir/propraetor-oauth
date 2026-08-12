# Public Client token handling — v2 crossroads

Deferred decisions from v1 grilling. v1 locks **Refresh Tokens for all Public Clients** with RFC 6749 wire format (JSON response + `refresh_token` body parameter) per ADR-0050.

## v2 shortlist (active)

When browser posture beyond v1 JSON refresh is required, choose among **A**, **B**, or **D** (not mutually exclusive in analysis—D may subsume or combine with A/B):

| ID | Path | Primary gain |
|---|---|---|
| **A** | HttpOnly refresh cookie on AS | Reduce refresh **exfiltration** |
| **B** | AS as token-mediating backend | Refresh server-only; no customer BFF |
| **D** | BFF or AS API proxy | Reduce **access token** exfiltration |

**Not on v2 shortlist:** **C** (no refresh for browser profile), **E** (profile split / split delivery)—set aside unless A/B/D prove insufficient.

---

## Crossroad A — HttpOnly refresh cookie on the Authorization Server

**Idea:** Browser clients do not store refresh in JS; AS sets refresh in an HttpOnly Secure cookie on the AS origin; Token Endpoint accepts cookie instead of (or in addition to) body `refresh_token`.

**Pros:** Reduces refresh **exfiltration** via XSS vs `localStorage`/memory.

**Cons:** RFC 6749 §6 literal wire format is a **documented extension**; requires Token Endpoint CSRF, CORS, Origin binding; cross-origin SPA (`app` → `auth`) complexity; refresh still **usable in-browser** during XSS (session extension, not offline exfil).

**Spec fit:** Within RFC 9700 rotation/binding goals; not strict 6749 without a profile ADR.

---

## Crossroad B — AS as token-mediating backend (no customer BFF)

**Idea:** Browser Public Clients get refresh **only server-side** on the AS; browser holds an opaque **OAuth session** cookie; SPA obtains new Access Tokens via a small AS session API; refresh never in JSON or cookie value.

**Pros:** No refresh exfiltration; operators need not deploy a BFF; “remember me” without refresh in the browser.

**Cons:** Custom AS surface beyond `/authorize`, `/token`, JWKS; CORS/CSRF on session endpoints; Access Token still exposed to JS if SPA calls RS directly with Bearer.

**Spec fit:** Similar security story to [OAuth 2.0 for Browser-Based Apps](https://datatracker.ietf.org/doc/html/draft-ietf-oauth-browser-based-apps) §6.2 (TMB), with AS replacing the backend component.

**Relation to A:** B avoids putting refresh **value** in any browser cookie; A keeps refresh grant at `/token` with cookie transport. B is strictly stronger for refresh confidentiality.

---

## Crossroad D — BFF or AS API proxy for browser consumers

**Idea:** Browser never holds Access or Refresh Tokens; authenticated API calls go through a server that attaches Bearer (customer BFF **or** AS proxy to Resource Server).

**Pros:** Best protection against **access token exfiltration**; aligns with browser BCP BFF pattern (§6.1).

**Cons:** Customer BFF adds operator burden; AS proxy blurs AS/RS boundary and is large scope; XSS can still **proxy requests in-browser** during session.

**Relation to A/B:** D addresses access-token exfil; A and B do not, unless combined with D (proxy) or very short access TTL only.

---

## Set aside (not v2 shortlist)

### Crossroad C — No Refresh Tokens for browser-profile Public Clients

Native keeps refresh; browser gets Access Token only; “remember me” via AS End-User login session + re-authorize. **Set aside** in favor of A/B/D.

### Crossroad E — Client profile split without dropping refresh

Unified semantics, split delivery (native JSON vs browser cookie/session API). **Set aside**—A/B already cover split delivery without a separate “E” track; native v1 wire format unchanged under 0050.

---

## v1 compensating controls (until a crossroad ships)

- Short Access Token TTL (warn >5m, refuse start >15m — ADR-0047)
- Rotating opaque refresh + family reuse detection (ADR-0007)
- Operator Not-Before on User or Client (`set-not-before` — ADR-0069)
- No client-facing RFC 7009 (ADR-0003)
- Public Clients: PKCE S256 + state (ADR-0023, 0004)

## Decision triggers

- **A or B:** SPA first-class; JSON refresh exfil unacceptable; operators won't run BFF → prefer **B** if AS should own mediation; **A** if minimal delta from v1 `/token`.
- **D:** Bar rises to **no exfiltratable access tokens** → BFF or AS proxy required; evaluate whether customer BFF or in-product AS proxy fits operator model.
