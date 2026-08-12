# Go supply-chain mitigations for propraetor-oauth

Research notes from **primary sources only** (go.dev blog + module reference, sum/proxy service pages, Go toolchain / vuln docs, and this repo’s CONTEXT.md + ADR-0035).  
Retrieved / verified: 2026-08-12.

Scope: how Go’s module system and tooling reduce supply-chain risk (especially silent mutation of existing versions), what still remains attackable, and a practical checklist for maintainers of this Authorization Server module.

---

## Executive summary / verdict

Go’s strongest supply-chain property is **content immutability + deliberate upgrades**, not “unsigned packages are safe.” Once a module version is recorded, `go.sum` plus the global checksum database (`sum.golang.org`) make silent retagging / bit-flipping of that version detectable ecosystem-wide. Builds do not auto-update dependencies: only `go get` / `go mod tidy` change the graph. There is **no separate package-registry upload account**—VCS tags are the source of truth; `proxy.golang.org` is a cache/mirror.

**Honest limits:** a compromised maintainer (or account) can still publish a **new** malicious version; the first recording of a never-before-seen version into sumdb is effectively “trust on first observation”; private modules that skip sumdb (`GOPRIVATE` / `GONOSUMDB`) lose global consistency; typosquatting and malicious *new* modules remain social/review problems; toolchain downloads inherit the same proxy/sumdb trust model (and historically had go-command bugs against malicious proxies—keep the **base** toolchain patched).

**For propraetor-oauth:** ADR-0035’s stdlib-first rule is itself a first-class mitigation. Today `go.mod` has **no `require` directives** (`go 1.26`). Keep that surface small; when dependencies land, treat every `go.mod` / `go.sum` delta as a security review artifact and never disable checksum verification in CI/prod.

---

## Project constraints

| Constraint | Source |
| --- | --- |
| Single-tenant first-party OAuth AS; Operator CLI + HTTP adapters | [CONTEXT.md](../../CONTEXT.md) |
| Prefer stdlib; external modules only when necessary | [ADR-0035](../adr/0035-stdlib-and-pure-screens.md) |
| Module currently dependency-free (`require`-less) | `go.mod` (`module github.com/gniewomir/propraetor-oauth`, `go 1.26`) |

---

## 1. How Go mitigates silent mutation and related attacks

### 1.1 Builds are locked — no automatic dependency updates

The version of every dependency in a build is fully determined by the main module’s `go.mod`. Since Go 1.16, build commands (`go build`, `go test`, `go install`, `go run`, …) fail if `go.mod` is incomplete. **Only** `go get` and `go mod tidy` are expected to change `go.mod` / the build graph; those are not intended to run automatically in CI. A compromised upstream publishing a new malicious version does **not** affect anyone until they explicitly update.

Sources:

- [https://go.dev/blog/supply-chain](https://go.dev/blog/supply-chain) (“All builds are locked”)
- [https://go.dev/ref/mod](https://go.dev/ref/mod) (module-aware commands; `-mod=readonly` default behavior when no vendor)

### 1.2 `go.sum` — local cryptographic hashes

`go.sum` lists cryptographic hashes (`h1:` SHA-256) of each contributing dependency’s zip contents and `go.mod` files. Incomplete `go.sum` causes errors on builds that need those modules; only deliberate dependency commands rewrite it. Downloads whose hash disagrees with `go.sum` are rejected and not installed in the module cache.

Sources:

- [https://go.dev/blog/supply-chain](https://go.dev/blog/supply-chain) (“Version contents never change”)
- [https://go.dev/ref/mod#authenticating](https://go.dev/ref/mod#authenticating) / [https://go.dev/ref/mod#go-sum-files](https://go.dev/ref/mod#go-sum-files)
- [https://go.dev/doc/modules/managing-dependencies](https://go.dev/doc/modules/managing-dependencies)

### 1.3 Checksum database (sumdb) — append-only global consistency

Beyond per-repo lockfiles, Go uses a **global, append-only, cryptographically verifiable** list of `go.sum` entries (Transparent Log / Merkle tree, Trillian-backed), served by default at `sum.golang.org`. When the `go` command needs a hash not already in `go.sum`, it looks it up in sumdb and verifies **inclusion** and **consistency** proofs against signed tree heads before accepting the hash. Consequences:

- Everyone using e.g. `example.com/module@v1.9.2` is guaranteed the **same bits**.
- Compromised origins or proxies cannot target specific dependents with different content for an already-logged version without detection.
- Authors cannot quietly move tags / rewrite version contents after the fact without the change being detected against the log.
- No per-author key management is required; it fits decentralized VCS-based modules.

If the version is not yet in the log, sumdb fetches it from the origin before replying (`/lookup`). Clients must authenticate lookup data against the signed tree hash timeline—not trust raw lookup text alone.

Sources:

- [https://go.dev/blog/supply-chain](https://go.dev/blog/supply-chain)
- [https://go.dev/blog/module-mirror-launch](https://go.dev/blog/module-mirror-launch)
- [https://go.dev/ref/mod#checksum-database](https://go.dev/ref/mod#checksum-database)
- [https://sum.golang.org/](https://sum.golang.org/)

### 1.4 Module proxy / mirror (`proxy.golang.org`)

Default `GOPROXY` is `https://proxy.golang.org,direct`. The mirror:

- Caches module zip / mod / metadata (availability if origin disappears—“left-pad”-style protection).
- Is **not** a separate upload registry: authors do not register an account or push artifacts; the proxy runs the same fetch logic as `go mod download`.
- Combined with sumdb, everyone using the proxy or fetching direct should see the same authenticated content for a given version.
- Sandboxes VCS on the proxy side; client defaults restrict public downloads to `git` / `hg` (`GOVCS`), reducing client-side VCS attack surface for uncommon tools.

Sources:

- [https://go.dev/blog/supply-chain](https://go.dev/blog/supply-chain) (“The VCS is the source of truth”)
- [https://proxy.golang.org/](https://proxy.golang.org/) / [https://sum.golang.org/](https://sum.golang.org/)
- [https://go.dev/ref/mod#goproxy-protocol](https://go.dev/ref/mod#goproxy-protocol), [https://go.dev/ref/mod#private-module-privacy](https://go.dev/ref/mod#private-module-privacy), [https://go.dev/ref/mod#vcs-govcs](https://go.dev/ref/mod#vcs-govcs)

### 1.5 VCS as source of truth (no registry upload account)

Import paths embed how to fetch from VCS; versions are tags. There is no second “publisher account” whose compromise can inject code that never appeared in the repository (as in ecosystems with separate upload pipelines). Malicious divergence between “source on GitHub” and “what the registry serves” is structurally harder—though a **compromised VCS host / maintainer account** can still tag malicious commits (see §2).

Source: [https://go.dev/blog/supply-chain](https://go.dev/blog/supply-chain)

### 1.6 Minimal version selection (MVS)

When adding a dependency, transitive requirements come from that dependency’s `go.mod` at the selected version—not an automatic leap to “latest of everything.” MVS picks the **minimum** version that satisfies the module graph. That reduces surprise upgrades and keeps upgrades deliberate.

Sources:

- [https://go.dev/blog/supply-chain](https://go.dev/blog/supply-chain)
- [https://go.dev/ref/mod#minimal-version-selection](https://go.dev/ref/mod#minimal-version-selection)

### 1.7 Fetch / build does not execute package code (no post-install hooks)

Design goal: neither fetching nor building runs the downloaded package’s code. There are no install-time scripts. Mitigation is partial: once you run tests or a binary, `init` in packages that participate in that build can run; packages not linked into a given build have no impact on it.

Source: [https://go.dev/blog/supply-chain](https://go.dev/blog/supply-chain) (“Building code doesn’t execute it”)

### 1.8 Culture + rich stdlib (“a little copying is better than a little dependency”)

Go’s proverb and ecosystem norms favor small trees; stdlib / `golang.org/x/...` cover many common needs. Tooling cannot eliminate trust in reused code—the smallest tree remains the strongest mitigation.

Sources:

- [https://go.dev/blog/supply-chain](https://go.dev/blog/supply-chain)
- This repo: [ADR-0035](../adr/0035-stdlib-and-pure-screens.md)

---

## 2. What still remains attackable (honest limits)

| Residual risk | Why Go’s defaults don’t fully close it | Primary notes |
| --- | --- | --- |
| **New malicious versions after account / VCS takeover** | Sumdb / `go.sum` pin *existing* versions. A new tag is a new version; victims must `go get` / tidy into it. Review still required. | [supply-chain blog](https://go.dev/blog/supply-chain) |
| **First observation / race before sumdb entry** | Local `go.sum` alone is trust-on-first-use. Sumdb makes the *first successfully logged* hash global. If origin is already malicious (or racing) when first recorded, that content becomes the ecosystem canonical hash for that version. | [module-mirror-launch](https://go.dev/blog/module-mirror-launch); [ref/mod checksum DB](https://go.dev/ref/mod#checksum-database) (`/lookup` may fetch origin if missing) |
| **Compromised or mis-set `GOPROXY` / `GOSUMDB`** | `GOSUMDB=off` (or `go get -insecure`) accepts unrecognized modules without sumdb. Proxies may mirror sumdb so the client never talks to `sum.golang.org` directly—trust shifts to that proxy’s sumdb responses. **GO-2026-4984** (CVE-2026-42501): a malicious proxy could bypass sumdb validation in affected `cmd/go` versions (toolchain + module paths); **upgrade the base toolchain** (fixes: Go **1.25.10** / **1.26.3**); re-tidy/`go mod verify` if you used an untrusted proxy. | [ref/mod Privacy / GOSUMDB](https://go.dev/ref/mod#private-module-privacy); [vuln.go.dev GO-2026-4984](https://vuln.go.dev/ID/GO-2026-4984.json); [announce](https://groups.google.com/g/golang-announce/c/qcCIEXso47M) |
| **Private modules / `GOPRIVATE` / `GONOSUMDB`** | Matching modules skip the public sumdb (and often the public proxy). Authentication falls back to whatever is already in `go.sum`, or **accept-on-first-download** into `go.sum` with no global cross-check. Org must supply its own integrity story (private sumdb mirror, careful PR review, optional vendor). | [ref/mod Private modules](https://go.dev/ref/mod#private-modules); [sum.golang.org](https://sum.golang.org/) (public services can’t see private code) |
| **Typosquatting / malicious new modules** | Decentralized import paths; no registry “name ownership” beyond DNS/VCS. Choosing the wrong module path is a review failure, not a sumdb failure. | Implication of VCS-based paths ([supply-chain blog](https://go.dev/blog/supply-chain)); not a sumdb guarantee |
| **Compromised toolchain downloads** | Newer toolchains download as modules `golang.org/toolchain@…`, proxied and sumdb-checked; downloads **fail** if `GOSUMDB=off`. `GOPRIVATE`/`GONOSUMDB` do **not** apply to toolchains. Still: you trust proxy+sumdb+`cmd/go` validation; keep base install patched (see GO-2026-4984). Optional: verify published archives with [`gorebuild`](https://go.dev/blog/rebuild). | [go.dev/doc/toolchain](https://go.dev/doc/toolchain); [go.dev/blog/rebuild](https://go.dev/blog/rebuild) |
| **`init` / test / runtime execution** | No post-install hooks ≠ no code execution after you build/test/run. | [supply-chain blog](https://go.dev/blog/supply-chain) |

**Not claimed by Go:** publisher-signed module attestations as a first-party required feature. “Authenticity” in-tree means **content hash continuity** (local `go.sum` + global sumdb), not author identity signatures. Vendoring (`go mod vendor`, `-mod=vendor`) copies source into the tree for offline/reproducible builds; it does not replace hash verification of what you vendored.

Sources: [https://go.dev/ref/mod#vendoring](https://go.dev/ref/mod#vendoring), sections above.

---

## 3. Operator tooling worth wiring in

### 3.1 `go mod verify`

Checks that module zip files and extracted dirs in the **module cache** still match hashes recorded when first downloaded. Complements (does not replace) `go.sum` checks on download. Useful in CI after restore of caches.

Source: [https://go.dev/ref/mod#go-mod-verify](https://go.dev/ref/mod#go-mod-verify)

### 3.2 `govulncheck` + Go vulnerability database

Curated OSV reports at [https://vuln.go.dev](https://vuln.go.dev); `govulncheck` reports vulns that are **actually reachable** from your code (call-graph aware), including stdlib/toolchain where applicable.

```bash
go install golang.org/x/vuln/cmd/govulncheck@latest
govulncheck ./...
```

Sources:

- [https://go.dev/security/vuln/](https://go.dev/security/vuln/)
- [https://go.dev/security/vuln/database](https://go.dev/security/vuln/database)
- [https://pkg.go.dev/golang.org/x/vuln/cmd/govulncheck](https://pkg.go.dev/golang.org/x/vuln/cmd/govulncheck)
- [https://go.dev/doc/tutorial/govulncheck](https://go.dev/doc/tutorial/govulncheck)

### 3.3 Inventory / reproducibility aids (first-party)

| Mechanism | What it gives you |
| --- | --- |
| Committed `go.mod` + `go.sum` | Deterministic dependency graph + hashes |
| `go list -m -json all` | Machine-readable module inventory for review/SBOM inputs |
| `go version -m <binary>` / `debug/buildinfo` | Embedded module versions in built binaries |
| `toolchain` / `go` lines + `GOTOOLCHAIN` | Explicit toolchain selection; downloads verified via sumdb |
| `go mod vendor` | Optional in-tree source snapshot for air-gapped / audit builds |
| `gorebuild` (x/build) | Independent bit-for-bit verification of official toolchain archives |

There is **no** mandatory first-party SBOM/provenance release format equivalent to “must ship SPDX + Sigstore.” Prefer `go list` / buildinfo as Go-native inventory; treat third-party SBOM generators as optional org policy, not Go core.

Sources: [ref/mod](https://go.dev/ref/mod), [toolchain doc](https://go.dev/doc/toolchain), [rebuild blog](https://go.dev/blog/rebuild)

### 3.4 Dependency change hygiene

Prefer explicit versions (`go get module@vX.Y.Z`) over casual `@latest` on production paths. Use `go list -m -u all` to discover upgrades, then bump deliberately. Keep PRs that touch `go.mod`/`go.sum` small and reviewed.

Source: [https://go.dev/doc/modules/managing-dependencies](https://go.dev/doc/modules/managing-dependencies)

---

## 4. Fit for propraetor-oauth

- **Empty require graph today** maximizes the cultural mitigation; every future dependency (Postgres driver, JWT, Argon2, CLI framework, etc.) should clear ADR-0035’s “only when necessary” bar and land with reviewed `go.sum` lines.
- Public module `github.com/gniewomir/propraetor-oauth`: default public proxy + sumdb apply; no need for `GOPRIVATE` unless private replaces appear.
- CI should treat unexpected `go.sum` changes as **incident-shaped** (compromise or mistaken bump), not noise.
- Pin / patch the **developer and CI base Go** at or above the fixed lines for known `cmd/go` proxy bugs (see GO-2026-4984: ≥ **1.26.3** on the 1.26 line).

---

## 5. Practical checklist

### Everyday development

- [ ] Keep `go.mod` and `go.sum` committed; never delete `go.sum` “to clean up” without regenerating and reviewing the diff.
- [ ] Do **not** set `GOSUMDB=off` (or use `go get -insecure`) on developer machines used for this repo.
- [ ] Prefer `go get module@vX.Y.Z` over `@latest` when adding or bumping runtime deps.
- [ ] Before adding a module: confirm import path (typosquat check), license, maintenance, and whether ADR-0035 allows it (stdlib / `golang.org/x` first).
- [ ] After any module change: run tests and skim the dependency’s release notes / diff for the version you selected.
- [ ] Leave default `GOPROXY` / `GOSUMDB` alone unless you have an org proxy; if you use a corporate proxy, treat it as trusted infrastructure.
- [ ] Keep the **installed** Go toolchain patched (not only `GOTOOLCHAIN=` pins).

### CI/CD

- [ ] Fail the build if `go.mod` would need updates (`-mod=readonly` / default without `-mod=mod`).
- [ ] Run `go mod verify` (especially when restoring a module cache).
- [ ] Optionally assert tidy: e.g. `go mod tidy` then `git diff --exit-code go.mod go.sum`.
- [ ] Run `govulncheck ./...` on a schedule or every PR; triage reachable findings.
- [ ] Never set `GOSUMDB=off` in CI or production image builds.
- [ ] Pin CI’s Go version explicitly; track Go security announcements ([golang-announce](https://groups.google.com/g/golang-announce)).
- [ ] Optional inventory artifact: `go list -m -json all` attached to builds; for release binaries, record `go version -m`.

### Dependency updates

- [ ] One concern per PR: dependency bumps isolated from feature work.
- [ ] Review the PR diff of **both** `go.mod` and `go.sum` (unexpected new modules in `go.sum` are a red flag).
- [ ] Prefer minimal bumps (MVS-friendly); avoid drive-by major upgrades.
- [ ] For high-impact deps: read upstream tag / compare source for that version; don’t trust the version label alone.
- [ ] After bumps: `go test ./...` and `govulncheck ./...`.
- [ ] Document why each new direct dependency is necessary (ADR-0035 alignment).

### Private modules (if/when any)

- [ ] Set `GOPRIVATE` (or tighter `GONOPROXY` / `GONOSUMDB`) only for true private path prefixes.
- [ ] Understand that those modules **skip** public sumdb: integrity is `go.sum` + your VCS/proxy controls.
- [ ] Prefer an internal module proxy that also mirrors checksums if the org needs sumdb-like guarantees for private code.
- [ ] Consider `go mod vendor` for release branches if auditors need an in-tree source snapshot.
- [ ] Never send private module paths to the public proxy/sumdb unintentionally (misconfigured `GOPRIVATE`).

### Incident / compromise response

- [ ] Unexpected `go.sum` churn without a deliberate bump → stop merges; compare with sumdb / `go mod download -json` hashes; treat as possible proxy or cache compromise.
- [ ] If an untrusted `GOPROXY` was used on a vulnerable toolchain: upgrade base Go first; then re-validate (`rm go.sum` is destructive—only in a clean worktree) via tidy + `go mod verify` as described in [GO-2026-4984](https://vuln.go.dev/ID/GO-2026-4984.json).
- [ ] Malicious **new** upstream version: pin/exclude/retract locally (`exclude` / replace with known-good / remove require); upgrade only after upstream fix; notify consumers if you published affected builds.
- [ ] Compromised maintainer account on a dependency: remove or fork/replace; audit binaries built since the bad version entered your graph (`go version -m`).
- [ ] Report malicious public modules to [security@golang.org](https://sum.golang.org/) (as directed on the module services page).
- [ ] Toolchain suspicion: reinstall from [https://go.dev/dl/](https://go.dev/dl/) or verify with `gorebuild`; avoid `GOSUMDB=off` during recovery.

---

## Primary source index

| Topic | URL |
| --- | --- |
| Supply-chain overview | [https://go.dev/blog/supply-chain](https://go.dev/blog/supply-chain) |
| Module reference (proxy, sumdb, env, verify, vendor, MVS) | [https://go.dev/ref/mod](https://go.dev/ref/mod) |
| Module mirror / sumdb / index services | [https://sum.golang.org/](https://sum.golang.org/), [https://proxy.golang.org/](https://proxy.golang.org/) |
| Mirror + sumdb launch (first-use / global log) | [https://go.dev/blog/module-mirror-launch](https://go.dev/blog/module-mirror-launch) |
| Managing dependencies | [https://go.dev/doc/modules/managing-dependencies](https://go.dev/doc/modules/managing-dependencies) |
| Toolchains + download verification | [https://go.dev/doc/toolchain](https://go.dev/doc/toolchain) |
| Reproducible toolchain archives | [https://go.dev/blog/rebuild](https://go.dev/blog/rebuild) |
| Vulnerability management / govulncheck | [https://go.dev/security/vuln/](https://go.dev/security/vuln/), [https://pkg.go.dev/golang.org/x/vuln/cmd/govulncheck](https://pkg.go.dev/golang.org/x/vuln/cmd/govulncheck) |
| Vuln DB API | [https://go.dev/security/vuln/database](https://go.dev/security/vuln/database), [https://vuln.go.dev](https://vuln.go.dev) |
| Malicious proxy / sumdb bypass (cmd/go) | [https://vuln.go.dev/ID/GO-2026-4984.json](https://vuln.go.dev/ID/GO-2026-4984.json) |
| Repo constraints | [CONTEXT.md](../../CONTEXT.md), [ADR-0035](../adr/0035-stdlib-and-pure-screens.md), `go.mod` |
