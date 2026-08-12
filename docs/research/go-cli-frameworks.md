# Go CLI argument / command handling for propraetor-oauth

Research notes from **primary sources only** (Go stdlib docs/source, package READMEs / official docs / module `go.mod`, and this repo’s CONTEXT.md + ADRs).  
Retrieved / verified: 2026-08-12.

Scope: pick an argv / subcommand approach for the single Operator-facing binary (`cmd/oauth`), for a first scaffold of **“not implemented” placeholders + auto help**, sized for the admin surface already named in domain docs — not for implementing admin logic yet.

---

## Executive summary / verdict

**Recommend `github.com/urfave/cli/v3` as the first CLI dependency** for the Operator binary scaffold.

Why this project:

- ADR-0019 needs one binary with nested-ish admin verbs (`server` + Client / User / Scope / Consent / purge / `set-not-before`). Auto help and a real command tree matter for Operator discoverability from day one.
- ADR-0035 prefers stdlib and allows external modules only when necessary. That ADR’s concrete examples are crypto / Postgres / JWT and **Resource Owner screens**, not CLI. A dedicated CLI framework is still a deliberate first dependency — justified here by Operator UX and a growing command tree — but the chosen library should stay **runtime-stdlib-light**.
- `urfave/cli` **v3 is the documented default for new development**; v2 is maintenance-only. The project states **no runtime dependencies except the Go standard library** (testify is for the library’s own tests). That is the lightest mature “full CLI” option among the contenders.
- **Do not take Viper** with the scaffold. Cobra’s Viper integration is optional; this product already splits Server Policy (TOML file) from CLI topology flags (ADR-0068) and keeps secrets in the environment.

**Strong runner-up:** stay on **stdlib `flag` + per-command `FlagSet`** if the human wants to keep `go.mod` dependency-free until real admin use cases land. Feasible for placeholders; help, aliases, suggestions, and nested trees become hand-rolled debt as the ADR surface expands.

**Also strong, different tradeoff:** `spf13/cobra` (ecosystem default; requires `pflag`; Viper optional). Prefer it if Operators/maintainers already expect Cobra ergonomics and you accept a non-stdlib flag parser as the first dep.

**Record an ADR** if any non-stdlib library is chosen (see §7).

---

## Project constraints that drive the choice

| Constraint | Source |
| --- | --- |
| One Operator binary; subcommands select `server` vs admin | CONTEXT.md “CLI”; [ADR-0019](../adr/0019-cli-server-and-admin.md) |
| Admin: CLI → domain → Postgres; no running server required | [ADR-0022](../adr/0022-admin-direct-postgres.md), [ADR-0033](../adr/0033-layered-adapters.md) |
| Composition root: `internal/adapter/cli`; thin `cmd/oauth` | `cmd/oauth/main.go`, `internal/adapter/cli/doc.go`, ADR-0033 |
| Stdlib-first; external modules only when necessary; screens are a separate pure-frontend rule | [ADR-0035](../adr/0035-stdlib-and-pure-screens.md) |
| Likely admin surface: Client / User / Scope; Consent revoke; unified `purge`; User/Client `set-not-before`; `server` + Server Policy + production-risk ack | CONTEXT.md; ADR-0015/0020, 0055, 0060, 0069, 0062, 0067, 0068 |
| Today: Go module with **no dependencies**; `main` prints “not implemented” | `go.mod`, `cmd/oauth/main.go` |

---

## Comparison at a glance

| Library | Latest stance (research time) | Subcommands / help | Flag scoping | Runtime deps | Fit for this binary |
| --- | --- | --- | --- | --- | --- |
| stdlib `flag` | Go 1 stdlib; `FlagSet` documents subcommands | Manual dispatch + manual help | Per-`FlagSet`; no built-in global/cascade | None | Good for zero-dep placeholders; weak as the tree grows |
| `urfave/cli` **v3** | **Recommended for new work**; v2 security/bugfix only | Nested `Commands`, aliases, prefix match, auto help, shell completion | Flags on commands; persistent-style via parent/root flags | Stdlib only at runtime (testify for library tests) | **Best default** for this scaffold |
| `spf13/cobra` | Actively released (e.g. v1.10.2); v1 line | Nested commands, aliases, suggestions, auto help, completion, man pages | Persistent vs local flags; optional Viper bind | **`pflag` required**; `mousetrap` (Windows); doc gens use other modules | Excellent UX; heavier first dep |
| `peterbourgon/ff` | **v3.4.0 stable**; **v4 still pre-release/beta** | v3: `ffcli.Command` tree; v4: `ff.Command` | Parent flag sets / shared flags | Core + optional yaml/toml parsers | Good stdlib-aligned alternative; version-line caution on v4 |
| `alecthomas/kong` | Stable **v1.x** (1.0 release cut) | Struct-tagged commands; auto `--help` | Flags/commands from struct tags | Runtime stdlib; assert/repr for library tests | Mature; different (declarative struct) style |
| `alecthomas/kingpin` | v2 stable but **contributions-only**; author uses Kong | Nested commands, help, completion | Fluent API | Several modules | **Avoid for new work** |

---

## 1. Stdlib `flag`

### What it is

Package `flag` implements command-line flag parsing. Top-level helpers and the `FlagSet` type are the supported API. The package comment states that **`FlagSet` exists so one can define independent sets of flags, such as to implement subcommands**.

Sources:

- Package overview (stdlib source): [https://cs.opensource.google/go/go/+/master:src/flag/flag.go](https://cs.opensource.google/go/go/+/master:src/flag/flag.go) (same text as `go doc flag`)
- Effective Go (flags as pointers / `flag.Arg`): [https://go.dev/doc/effective_go#flag](https://go.dev/doc/effective_go#flag)

### Subcommand model

There is **no** first-class command tree. The usual pattern is: read `os.Args[1]` (or walk args), select a `FlagSet` named for that subcommand, call `Parse` on the remainder, then run handler code. Help (`-h` / `-help`) is per `FlagSet` via the package’s help handling (`ErrHelp`); composing a root “list of commands” help screen is application code.

### Flag scoping

Each `FlagSet` is independent. “Global” flags mean either parsing a parent set first or registering the same flags on every set. No cascading / persistent-flag abstraction.

### Dependency weight

Zero. Already available with Go 1.25.x in this module.

### Fit for propraetor-oauth

Adequate for a few top-level verbs and placeholder `Run` stubs. Poor match once Operator help must explain `server` policy/risk flags (ADR-0067/0068), `purge --target=… --older-than`, and nested Client/User/Scope verbs without hand-written usage glue.

### Testing

No dedicated “CLI test harness” beyond constructing a `FlagSet`, calling `Parse` with a `[]string`, and asserting. Fully under your control; no framework docs required.

---

## 2. `spf13/cobra` (+ `pflag`, optional `viper`)

### Maintenance / version

- Library module: `github.com/spf13/cobra` (Apache-2.0). Research-time release example: **v1.10.2** ([https://github.com/spf13/cobra/releases/tag/v1.10.2](https://github.com/spf13/cobra/releases/tag/v1.10.2), [https://pkg.go.dev/github.com/spf13/cobra](https://pkg.go.dev/github.com/spf13/cobra)).
- README positions Cobra as used by Kubernetes, Hugo, GitHub CLI, etc.  
  Source: [https://github.com/spf13/cobra/blob/main/README.md](https://github.com/spf13/cobra/blob/main/README.md)

### Subcommand model

First-class nested commands via `AddCommand`. Features claimed by the project README / user guide:

- Nested subcommands  
- Automatic help (`-h` / `--help`)  
- Command **aliases** (`Aliases []string` on `cobra.Command`)  
- Intelligent suggestions on unknown commands  
- Shell completion; optional man/Markdown generation via `cobra/doc`  

Sources: README above; User Guide: [https://github.com/spf13/cobra/blob/main/site/content/user_guide.md](https://github.com/spf13/cobra/blob/main/site/content/user_guide.md); `Aliases` field in [https://pkg.go.dev/github.com/spf13/cobra#Command](https://pkg.go.dev/github.com/spf13/cobra#Command)

### Flag scoping

- **Persistent flags** — available on the command and every child.  
- **Local flags** — only that command (`Flags()`).  
- Optional `TraverseChildren` to parse parent local flags when walking to the leaf.  
- Optional **Viper** binding (`viper.BindPFlag`) — documented as optional integration, not required to use Cobra.

Sources: User Guide “Persistent Flags” / “Local Flags” / “Bind Flags with Config”; README: “Optional seamless integration with viper”.

**pflag (required):** Flag functionality is provided by `github.com/spf13/pflag`, described as a fork of stdlib `flag` that keeps a similar interface while adding POSIX/GNU-style flags (short/long, combined shorts, etc.).  
Sources: Cobra README; [https://github.com/spf13/pflag/blob/master/README.md](https://github.com/spf13/pflag/blob/master/README.md)

**Viper (optional):** Separate configuration library (files, env, remote, …). Not needed for argv parsing.  
Source: [https://github.com/spf13/viper/blob/master/README.md](https://github.com/spf13/viper/blob/master/README.md)

### Dependency weight

From Cobra’s `go.mod` (runtime-relevant):

- `github.com/spf13/pflag` — used from the main Cobra package  
- `github.com/inconshreveable/mousetrap` — Windows double-click guard (`command_win.go`)  
- `github.com/cpuguy83/go-md2man/v2` and `go.yaml.in/yaml/v3` — used from the **`cobra/doc`** subpackage (man/yaml doc generation), not from ordinary command execution imports  

Source: [https://github.com/spf13/cobra/blob/main/go.mod](https://github.com/spf13/cobra/blob/main/go.mod) and package imports under that module.

So: adopting Cobra means accepting **pflag** (and typically mousetrap in the module graph) as the first real CLI stack — still far smaller than “Cobra + Viper”, but not stdlib-only.

### Fit for propraetor-oauth

Excellent for a single binary with `server` + a deep admin tree, POSIX flag UX, suggestions, and completion. Slightly heavier than needed for “placeholders + help”, and the first dependency is a **non-stdlib flag parser**. Package name `cobra` does not collide with `package cli` in `internal/adapter/cli`.

### Testing

User Guide demonstrates `rootCmd.SetArgs([]string{…})` then `Execute()` to drive hooks/commands without a real process argv — suitable as a unit-test pattern.  
Source: User Guide “PreRun / … hooks” example with `SetArgs`.

---

## 3. `urfave/cli` (v2 vs v3)

### Maintenance / version stance (authoritative)

Official site:

- **v3** — `go get github.com/urfave/cli/v3@latest`; **“recommended version for all new development.”**  
- **v2** — “receiving security and bug fixes only” via `v2-maint`; **“should not be used in new development.”**  
- **v1** — security fixes only; not for new development.

Source: [https://cli.urfave.org/](https://cli.urfave.org/)

Research-time module version example: **v3.10.1** ([https://pkg.go.dev/github.com/urfave/cli/v3](https://pkg.go.dev/github.com/urfave/cli/v3)).

v3 migration notes (API shape): root is `cli.Command` (not `cli.App`); nested commands field is `Commands`; handlers take `(context.Context, *cli.Command)`; doc generation and altsrc moved to optional sibling modules; suggestions no longer need `xrash/smetrics`.  
Source: [https://cli.urfave.org/migrate-v2-to-v3/](https://cli.urfave.org/migrate-v2-to-v3/)

### Features (project claims)

From README / welcome page:

- Commands and subcommands with **alias and prefix match**  
- Flexible help system  
- Dynamic shell completion (bash, zsh, fish, powershell)  
- **“no dependencies except Go standard library”** (runtime)  
- Compound short flags  
- Optional docs module `urfave/cli-docs`; optional altsrc module `urfave/cli-altsrc`

Sources: [https://github.com/urfave/cli/blob/main/README.md](https://github.com/urfave/cli/blob/main/README.md), [https://cli.urfave.org/](https://cli.urfave.org/)

Migration guide clarifies the dependency story: *“v3's minimal dependency tree (only `github.com/stretchr/testify` in tests)”*.  
Source: [https://cli.urfave.org/migrate-v2-to-v3/](https://cli.urfave.org/migrate-v2-to-v3/)

Getting started: an empty `(&cli.Command{}).Run(ctx, os.Args)` already prints help.  
Source: [https://cli.urfave.org/v3/getting-started/](https://cli.urfave.org/v3/getting-started/)

### Flag scoping

Flags are declared on each `Command`. Parent/root flags act as the global surface; leaf commands carry command-local flags. v3 exposes helpers such as `LocalFlagNames` (migration table) for distinguishing local vs inherited visibility. Persistent-flag *naming* is Cobra’s vocabulary; urfave’s model is still parent/child command flags + help categories.

### Dependency weight

Module `go.mod` lists `testify` because the library’s own tests import it; non-`_test.go` library sources do not import third-party packages (verified against module source at v3.10.1). Consumer binaries that only import `urfave/cli/v3` therefore stay **stdlib-linked for CLI parsing**.

### Fit for propraetor-oauth

Strong match: nested admin tree, auto help for placeholders, completion later, no Viper, minimal dep cost relative to ADR-0035.  

**Naming caveat:** the imported package name is `cli`, same as `internal/adapter/cli`. Use an import alias (e.g. `ucli` / `cliframe`) inside the adapter package.

### Testing

Official docs emphasize building apps via `Command.Run(ctx, args)`. Library itself is heavily tested with testify. Application tests can call `Run` with a synthetic `[]string` (same pattern as production entry). No separate “SetArgs” API is required because argv is an explicit argument to `Run`.

---

## 4. Other contenders (brief)

### `peterbourgon/ff` (v3 stable / v4 pre-release)

- Philosophy: **flags-first** configuration; `myprogram -h` should show the full config surface.  
- **v3** (`github.com/peterbourgon/ff/v3`): stable latest release **v3.4.0**; subcommands via **`ffcli.Command`** (stdlib `flag.FlagSet` + optional `ff.Parse` for env/config). YAML/TOML parsers live in subpackages (`ffyaml`, `fftoml`) with their own deps — omit them if unused.  
- **v4** (`github.com/peterbourgon/ff/v4`): README explicitly describes **pre-release** v4; proxy versions include `v4.0.0-alpha.*` and `v4.0.0-beta.1`. Prefer v3 if choosing ff today.

Sources:

- v3/v4 README (pkg.go.dev README text): [https://pkg.go.dev/github.com/peterbourgon/ff/v3](https://pkg.go.dev/github.com/peterbourgon/ff/v3), [https://pkg.go.dev/github.com/peterbourgon/ff/v4](https://pkg.go.dev/github.com/peterbourgon/ff/v4)  
- `ffcli` package docs in module source (`ffcli/command.go`)

Fit: closest “framework” to stdlib while still offering a command tree. Slightly more assembly than urfave/Cobra for rich help/completion. Sensible if the human wants FlagSet purity and can accept less “batteries included” Operator UX.

### `alecthomas/kong`

- Command lines expressed as **Go structs + tags**; `kong.Parse`; auto `--help`; nested `cmd:""` structs; `Run(...) error` methods.  
- **v1.0.0 release** announced in README (“stable for a long time”). Research-time module: **v1.16.1**.  
- `assert` / `repr` appear in `go.mod` for the library’s tests, not as runtime imports of non-test sources.

Source: [https://github.com/alecthomas/kong/blob/master/README.md](https://github.com/alecthomas/kong/blob/master/README.md), [https://pkg.go.dev/github.com/alecthomas/kong](https://pkg.go.dev/github.com/alecthomas/kong)

Fit: mature and elegant for struct-driven CLIs. Slightly more “magic” than this repo’s current explicit-adapter style; fine technically, less aligned with “thin main + obvious command constructors” than urfave/Cobra/ffcli.

### `alecthomas/kingpin/v2`

- Fluent nested commands, help, completion.  
- README banner: **“CONTRIBUTIONS ONLY”**; author no longer uses Kingpin personally and now uses **Kong**.

Source: [https://github.com/alecthomas/kingpin/blob/master/README.md](https://github.com/alecthomas/kingpin/blob/master/README.md)

Fit: **not recommended** for a greenfield Operator CLI.

---

## 5. Recommendation for *this* first scaffold

### Choice

Adopt **`github.com/urfave/cli/v3`** in `internal/adapter/cli`, keep `cmd/oauth/main.go` as a thin `os.Exit`-wrapping caller, and register placeholder commands whose `Action` returns a clear “not implemented” error (or prints and exits non-zero), while relying on generated help for discoverability.

### Rationale mapped to constraints

| Concern | How urfave/cli v3 answers it |
| --- | --- |
| ADR-0019 single binary / subcommands | Native nested `Commands` (`server`, `client`, …) |
| ADR-0035 stdlib preference | Runtime stdlib-only CLI core; still a conscious first module — document via ADR |
| Zero deps today | Smallest “full CLI” dep among mature options; no pflag/Viper |
| Operator UX | Help from empty/partial trees; aliases/prefix match; completion available later |
| Future admin growth | Same library scales to purge/consent/server flag groups without rewriting dispatch |
| ADR-0068 / 0067 flags | Per-command flags + help text on dangerous flags; no need for Viper |

### When to pick something else instead

- **Stdlib `flag` only** — if the human wants the scaffold committed with **no** `go.mod` require line, accepting hand-written root help until a later ADR.  
- **Cobra** — if team familiarity / ecosystem norms outweigh stdlib-light preference.  
- **ff v3 `ffcli`** — if you want FlagSet-centric design and are willing to assemble more UX yourself; stay off v4 until it is out of pre-release per upstream README.

### Explicit non-choices for the scaffold

- Do **not** add Viper. Server Policy is a required TOML file for `server` (ADR-0068); admin uses env for Postgres secrets (ADR-0022).  
- Do **not** start on urfave **v2** or Kingpin.  
- Do **not** put the command tree under a Cobra-style top-level `cmd/` package that becomes the composition root — that would fight ADR-0033 / existing `internal/adapter/cli` docs.

---

## 6. Recommended package layout (names only)

Keep the binary entry thin; put the tree and wiring in the driving adapter:

```text
cmd/oauth/
  main.go                 # parse nothing; call adaptercli.Run(ctx, os.Args); map error → exit

internal/adapter/cli/
  doc.go                  # existing package docs
  run.go                  # root Command construction + Run
  server.go               # placeholder: server
  client.go               # placeholder group: client … (incl. set-not-before)
  user.go                 # placeholder group: user … (incl. set-not-before)
  scope.go                # placeholder group: scope …
  consent.go              # placeholder group: consent …
  purge.go                # placeholder: purge
```

Import the framework with an alias inside this package, e.g.:

```go
import ucli "github.com/urfave/cli/v3"
```

### Minimal command tree (domain vocabulary; placeholders)

Binary name as shipped: **`oauth`** (ADRs often say `cli` generically).

```text
oauth
├── server                          # run AS; requires Server Policy (ADR-0068); risk flags (ADR-0067)
├── client
│   ├── create
│   ├── delete
│   ├── list                        # optional for v1 scaffold discoverability
│   ├── show
│   ├── scope-add                   # assign Scope to Client (ADR-0015/0020) — exact verb TBD
│   ├── scope-remove
│   └── set-not-before              # Client Not-Before; --at optional, now only (ADR-0069)
├── user                            # CLI synonym for End-User (CONTEXT.md)
│   ├── create
│   ├── delete
│   ├── list
│   ├── show
│   ├── set-password                # named in ADR-0062; does not advance Not-Before
│   └── set-not-before              # User Not-Before; --at optional, now only (ADR-0069)
├── scope
│   ├── create
│   ├── delete
│   └── list
├── consent
│   └── revoke                      # per End-User / End-User×Client / End-User×Client×Scope (ADR-0055)
└── purge                           # --target=entities|audit|all --older-than (ADR-0060)
```

Notes:

- Exact flag spellings for `server` (listen, Issuer URL, PEM paths, `--policy`, `--allow-http-issuer`, `--allow-cleartext-listen`, `--i-understand-production-risk`) are constrained by ADR-0067/0068 but need not be implemented in the first scaffold — help stubs can mention them when those commands grow.  
- Prefer CONTEXT.md terms (`user` not “end-user” on the CLI; `client`; `scope`; `consent`; `purge`; `server`).
- Operator compromise response is **Not-Before** (`set-not-before`), not a `refresh-token invalidate` verb (ADR-0069).

---

## 7. ADR decision to record

If the scaffold adds **any** non-stdlib CLI library, add a short ADR (suggested title: **“Operator CLI uses urfave/cli v3”** — or Cobra/ff/stdlib, matching the pick) covering:

1. **Decision:** library + major version module path.  
2. **Context:** ADR-0019 single binary; composition root in `internal/adapter/cli`; Operator help/discoverability.  
3. **Relation to ADR-0035:** CLI parsing is an allowed external module because stdlib `flag` lacks a maintained command-tree/help model adequate for the planned admin surface; Resource Owner screen purity is unchanged.  
4. **Non-goals:** no Viper; no remote admin HTTP API (ADR-0019); admin still goes domain → Postgres (ADR-0022).  
5. **Consequences:** first `go.mod` require; import-alias if package name is `cli`; future commands stay in the adapter package (or clear subpackages), not a second rules engine.

If choosing **stdlib only**, a lighter ADR (or a note in an existing CLI ADR) should still say “subcommand dispatch is hand-rolled `flag.FlagSet` until further notice” so agents do not casually add Cobra later.

---

## 8. Open questions / decisions for the human

1. **Accept first dependency now?** Prefer urfave/cli v3 for placeholders+help, or stay zero-dep with stdlib until the first real admin command?  
2. **Cobra instead?** If Operators/docs authors already standardize on Cobra elsewhere, that may outweigh stdlib-light.  
3. **Client scope verbs:** `client scope-add` vs `client allow-scope` vs nested `client scope add` — pick one nesting style before the tree hardens.  
4. **Root help groups:** whether to group commands in help as “runtime” (`server`) vs “administration” (everything else) — both Cobra and urfave support categorisation-style UX; exact API differs.  
5. **Completion / man pages in v1?** Optional; neither is required for the placeholder scaffold.

Settled elsewhere: User/Client `set-not-before` (ADR-0069); no `refresh-token invalidate` verb.

---

## Primary source index

| Topic | URL |
| --- | --- |
| Go `flag` package (FlagSet / subcommands) | [https://cs.opensource.google/go/go/+/master:src/flag/flag.go](https://cs.opensource.google/go/go/+/master:src/flag/flag.go) |
| Effective Go — command-line flags | [https://go.dev/doc/effective_go#flag](https://go.dev/doc/effective_go#flag) |
| Cobra README | [https://github.com/spf13/cobra/blob/main/README.md](https://github.com/spf13/cobra/blob/main/README.md) |
| Cobra User Guide | [https://github.com/spf13/cobra/blob/main/site/content/user_guide.md](https://github.com/spf13/cobra/blob/main/site/content/user_guide.md) |
| Cobra module / versions | [https://pkg.go.dev/github.com/spf13/cobra](https://pkg.go.dev/github.com/spf13/cobra) |
| pflag README | [https://github.com/spf13/pflag/blob/master/README.md](https://github.com/spf13/pflag/blob/master/README.md) |
| Viper README (optional vs Cobra) | [https://github.com/spf13/viper/blob/master/README.md](https://github.com/spf13/viper/blob/master/README.md) |
| urfave/cli welcome (v3 vs v2 vs v1) | [https://cli.urfave.org/](https://cli.urfave.org/) |
| urfave/cli v3 getting started | [https://cli.urfave.org/v3/getting-started/](https://cli.urfave.org/v3/getting-started/) |
| urfave/cli migrate v2→v3 | [https://cli.urfave.org/migrate-v2-to-v3/](https://cli.urfave.org/migrate-v2-to-v3/) |
| urfave/cli README | [https://github.com/urfave/cli/blob/main/README.md](https://github.com/urfave/cli/blob/main/README.md) |
| urfave/cli/v3 on pkg.go.dev | [https://pkg.go.dev/github.com/urfave/cli/v3](https://pkg.go.dev/github.com/urfave/cli/v3) |
| ff v3 / v4 | [https://pkg.go.dev/github.com/peterbourgon/ff/v3](https://pkg.go.dev/github.com/peterbourgon/ff/v3), [https://pkg.go.dev/github.com/peterbourgon/ff/v4](https://pkg.go.dev/github.com/peterbourgon/ff/v4) |
| Kong README | [https://github.com/alecthomas/kong/blob/master/README.md](https://github.com/alecthomas/kong/blob/master/README.md) |
| Kingpin README (contributions-only) | [https://github.com/alecthomas/kingpin/blob/master/README.md](https://github.com/alecthomas/kingpin/blob/master/README.md) |
| Repo CONTEXT + ADRs cited above | `CONTEXT.md`, `docs/adr/0019`, `0022`, `0033`, `0035`, `0007`, `0015`, `0055`, `0060`, `0062`, `0067`, `0068` |
