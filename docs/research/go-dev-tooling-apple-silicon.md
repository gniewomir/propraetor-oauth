# Go development tooling on Apple Silicon Macs

Research notes from **primary sources only** (go.dev, pkg.go.dev, official GitHub repos / first-party docs).  
Retrieved / verified: 2026-08-12.

**Apple Silicon specifics** appear where the owning docs call them out (`darwin/arm64` downloads, Homebrew bottles, race detector ports, Delve macOS notes). Where a tool is arch-agnostic once installed, that is stated.

---

## 1. Minimum / baseline tooling

What you need to install and use Go productively on an Apple Silicon Mac: toolchain, modules, format/vet, editor + language server.

### 1.1 Install the Go toolchain (`darwin/arm64`)

**Official downloads** publish a dedicated Apple Silicon build:

- Label: “Apple macOS (ARM64)” — macOS 12 or later, Apple 64-bit processor  
- Artifacts (example at time of research): `go1.26.5.darwin-arm64.pkg` (installer) and `go1.26.5.darwin-arm64.tar.gz` (archive)  
- Source: [https://go.dev/dl/](https://go.dev/dl/)

**Install steps** (official install guide):

1. Remove any previous install under `/usr/local/go` if present.  
2. Install via the macOS package, or extract the archive into `/usr/local` so you get `/usr/local/go`.  
3. Put `/usr/local/go/bin` on `PATH`.  
4. Also put the Go bin directory (default `$HOME/go/bin`, or `$(go env GOPATH)/bin`) on `PATH` for tools installed with `go install`.  
5. Verify: `go version`.

Sources: [https://go.dev/doc/install](https://go.dev/doc/install), [https://go.dev/dl/](https://go.dev/dl/)

**Homebrew alternative** (first-party Homebrew formula, bottles for Apple Silicon):

```bash
brew install go
```

- Formula documents bottles for macOS on Apple Silicon (tahoe / sequoia / sonoma).  
- Requires macOS ≥ 12.  
- Current stable noted on formulae.brew.sh at research time: 1.26.5.  
- Source: [https://formulae.brew.sh/formula/go](https://formulae.brew.sh/formula/go)

**Verify you are on arm64** (expected on Apple Silicon):

```bash
go env GOOS GOARCH
# expect: darwin / arm64
```

`darwin` + `arm64` is a supported `GOOS`/`GOARCH` combination in the official source install docs.  
Source: [https://go.dev/doc/install/source](https://go.dev/doc/install/source)

### 1.2 Toolchain switching (Go 1.21+)

Starting in Go 1.21, the distribution is a `go` command plus a bundled toolchain. With default `GOTOOLCHAIN=auto`, the `go` command can download/use a newer toolchain when `go.mod` / `go.work` require it.

- Manage preferred/min versions via `go` / `toolchain` lines and `go get go@…`.  
- Force a version for a command: `GOTOOLCHAIN=go1.21rc3 go test` (example from docs).  

Source: [https://go.dev/doc/toolchain](https://go.dev/doc/toolchain)

Multiple side-by-side versions can also be installed with `go install golang.org/dl/go1.X.Y@latest` then `go1.X.Y download`.  
Source: [https://go.dev/doc/manage-install](https://go.dev/doc/manage-install)

### 1.3 Modules (default dependency model)

Baseline workflow from official tutorials / “How to Write Go Code”:

| Command | Role |
| --- | --- |
| `go mod init <module-path>` | Create `go.mod`, enable dependency tracking |
| `go mod tidy` | Add missing module requirements / sums; drop unused |
| `go get <module>@<version>` | Add or change a dependency |
| `go run .` / `go build` / `go install` | Run, compile, install |
| `go test` | Run tests in `*_test.go` |

Modules do **not** require the project to live under `GOPATH`. Default `GOPATH` is `$HOME/go` if unset.  
Sources:

- [https://go.dev/doc/tutorial/getting-started](https://go.dev/doc/tutorial/getting-started)  
- [https://go.dev/doc/code](https://go.dev/doc/code)  
- [https://go.dev/wiki/SettingGOPATH](https://go.dev/wiki/SettingGOPATH)  
- Modules reference: [https://go.dev/ref/mod](https://go.dev/ref/mod)

Optional multi-module local work: `go work init` / `go work use` (`go.work`).  
Source: [https://go.dev/doc/tutorial/workspaces](https://go.dev/doc/tutorial/workspaces)

### 1.4 Format and vet (stdlib tools)

**Formatting**

- `gofmt -w file.go` or `go fmt <package>`  
- Canonical style: easier to write/read/maintain; reduces formatting noise in diffs.  
- Source: [https://go.dev/blog/gofmt](https://go.dev/blog/gofmt)

**Static checks (`go vet`)**

- Official diagnostics page groups debugging/profiling; `gopls` and editors surface **build, vet, and analysis** diagnostics.  
- golangci-lint’s default `govet` linter is documented as roughly the same as `go vet` and using its passes.  
- Practical baseline command: `go vet ./...` (package patterns same as other `go` tools).  
- Sources: [https://go.dev/doc/diagnostics](https://go.dev/doc/diagnostics), [https://golangci-lint.run/docs/welcome/quick-start/](https://golangci-lint.run/docs/welcome/quick-start/), [https://go.dev/gopls/](https://go.dev/gopls/)

**What they guard against**

| Tool | Guards against |
| --- | --- |
| `gofmt` / `go fmt` | Non-canonical layout, style debates, noisy diffs |
| `go vet` | Suspicious constructs the compiler does not reject (via vet passes) |

### 1.5 Testing (built-in)

- Files ending in `_test.go`; functions `TestXxx(t *testing.T)`.  
- `go test` / `go test -v`  
- Source: [https://go.dev/doc/tutorial/add-a-test](https://go.dev/doc/tutorial/add-a-test), [https://go.dev/doc/code](https://go.dev/doc/code)

### 1.6 Editor + LSP (`gopls`)

**`gopls`** is the official language server from the Go team.

```bash
go install golang.org/x/tools/gopls@latest
```

- Provides LSP features: hover, completion, diagnostics, navigation, formatting, rename, organize imports, etc.  
- VS Code may install/update gopls for you.  
- Supports module, multi-module, and GOPATH layouts.  
- While running, gopls uses the `go` on `$PATH`; it follows the Go release policy (two most recent major releases officially).  
- Source: [https://go.dev/gopls/](https://go.dev/gopls/), feature index: [https://go.dev/gopls/features/](https://go.dev/gopls/features/)

**VS Code + official Go extension** (`golang.go`)

Requirements (extension README): VS Code 1.90+, Go 1.21+.

Quick start:

1. Install Go.  
2. Install the Go extension.  
3. Open a `.go` / `go.mod` file; extension activates.  
4. Depends on `go`, `gopls`, and optional tools; missing `gopls` is installed automatically.  

Feature highlights called out by the extension: IntelliSense, navigation, formatting/import organization, diagnostics (build/vet/lint), testing, debugging.  
Source: [https://github.com/golang/vscode-go](https://github.com/golang/vscode-go) (README), tools wiki: [https://github.com/golang/vscode-go/wiki/tools](https://github.com/golang/vscode-go/wiki/tools)

**Other editors** (mentioned by official getting-started tutorial as popular): VS Code (free), GoLand (paid), Vim (free).  
Source: [https://go.dev/doc/tutorial/getting-started](https://go.dev/doc/tutorial/getting-started)

### 1.7 Baseline checklist (Apple Silicon)

```bash
# Install Go: official darwin-arm64 .pkg from https://go.dev/dl/
#   or: brew install go

export PATH="/usr/local/go/bin:$PATH"                 # if using official pkg/tar
export PATH="$PATH:$(go env GOPATH)/bin"              # go install tools

go version
go env GOOS GOARCH                                    # darwin arm64

go install golang.org/x/tools/gopls@latest

# In a project:
go mod init example.com/myapp
go test ./...
go vet ./...
go fmt ./...
```

Plus an LSP-capable editor with gopls (VS Code Go extension is the first-party documented path).

---

## 2. DX improvements (beyond the minimum)

Additions that improve day-to-day development: debugging, richer editor UX, live reload, package managers, profiling.

### 2.1 Debugging with Delve (`dlv`)

Official Go diagnostics docs recommend **Delve** as the Go-oriented debugger (GDB works but is “not ideal”).  
Source: [https://go.dev/doc/diagnostics](https://go.dev/doc/diagnostics)

**Install** (Delve project docs; works on macOS):

```bash
go install github.com/go-delve/delve/cmd/dlv@latest
```

Also: prebuilt binaries on GitHub Releases; Homebrew: `brew install delve` (Apple Silicon bottles documented).  
Sources:

- [https://github.com/go-delve/delve/blob/master/Documentation/installation/README.md](https://github.com/go-delve/delve/blob/master/Documentation/installation/README.md)  
- [https://formulae.brew.sh/formula/delve](https://formulae.brew.sh/formula/delve)

**macOS-specific** (Delve install docs):

1. Install CLT: `xcode-select --install`  
2. Without Developer Mode, macOS may ask to authorize the debugger every use. Enable:  
   `sudo /usr/sbin/DevToolsSecurity -enable`  
3. Optionally add your user to `_developer`:  
   `sudo dscl . append /Groups/_developer GroupMembership $(whoami)`  
4. Docs note the macOS **native** backend is unnecessary and “has known problems”; default install path is fine for normal use.

**VS Code integration**

- Extension uses Delve; `dlv-dap` is the default for local debugging.  
- First debug session may prompt to install/update `dlv`; or use **Go: Install/Update Tools**.  
- Source: [https://github.com/golang/vscode-go/blob/master/docs/debugging.md](https://github.com/golang/vscode-go/blob/master/docs/debugging.md)

**Tip from Go diagnostics:** for clearer debugging of optimized code, build with  
`go build -gcflags=all="-N -l"`.  
Source: [https://go.dev/doc/diagnostics](https://go.dev/doc/diagnostics)

### 2.2 Editor / IDE feature upgrades

| Addition | What it improves | Source |
| --- | --- | --- |
| gopls full feature set | Hover, signature help, inlay hints, semantic tokens, call/type hierarchy, extract/inline, add test, etc. | [gopls features](https://go.dev/gopls/features/) |
| VS Code Go tools (`gotests`, `impl`, `goplay`, …) | Generate tests/stubs; playground; via Go: Install/Update Tools | [vscode-go tools wiki](https://github.com/golang/vscode-go/wiki/tools) |
| Semantic tokens in VS Code | Better highlighting via gopls `ui.semanticTokens` | [vscode-go README](https://raw.githubusercontent.com/golang/vscode-go/master/README.md) |
| GoLand | Listed as popular Go IDE in official tutorial; golangci-lint docs note built-in golangci-lint support from 2025.1 | [getting started](https://go.dev/doc/tutorial/getting-started), [golangci integrations](https://golangci-lint.run/docs/welcome/integrations/) |

### 2.3 Testing UX

Beyond CLI `go test`:

- VS Code Go: enhanced **testing** and **debugging** of tests in the editor (extension README / debugging docs).  
- Race-aware tests (also a quality guardrail — see §3): `go test -race`. Supported on **`darwin/arm64`**.  
  Source: [https://go.dev/doc/articles/race_detector](https://go.dev/doc/articles/race_detector)

### 2.4 Live reload (Air)

**Air** (`github.com/air-verse/air`) — live-reloading CLI for Go app **development** (explicitly not hot-deploy for production).

```bash
# go 1.25+ per Air README
go install github.com/air-verse/air@latest
# or: brew install go-air
air init   # writes .air.toml
air
```

Platform-specific overrides include `[build.darwin]`. Can wire Delve into `build.entrypoint` for headless debug.  
Source: [https://github.com/air-verse/air](https://github.com/air-verse/air) (README)

### 2.5 Package managers / install channels (Apple Silicon)

| Channel | Notes | Source |
| --- | --- | --- |
| Official `.pkg` / `.tar.gz` `darwin-arm64` | Canonical Go install | [go.dev/dl](https://go.dev/dl/) |
| Homebrew `go`, `delve`, `golangci-lint`, `go-air` | Bottles for Apple Silicon where documented | [formulae.brew.sh/formula/go](https://formulae.brew.sh/formula/go), etc. |
| `go install …@version` | Standard for gopls, dlv, govulncheck, staticcheck | Various official tool docs |

**Caution (golangci-lint + Homebrew):** maintainers note Homebrew may build with an unexpected Go version; they recommend their binaries or ensuring the build Go version. Prefer official formula over old taps.  
Source: [https://golangci-lint.run/docs/welcome/install/local/](https://golangci-lint.run/docs/welcome/install/local/)

### 2.6 Profiling / tracing (when investigating performance)

Built into the Go toolchain:

- `go test` / `runtime/pprof` / `net/http/pprof` + `go tool pprof`  
- Execution tracer: `go tool trace`  
- On macOS, diagnostics docs mention **Instruments** for profiling Go programs  

Source: [https://go.dev/doc/diagnostics](https://go.dev/doc/diagnostics)

### 2.7 Practical DX set (optional installs)

```bash
go install github.com/go-delve/delve/cmd/dlv@latest
go install github.com/air-verse/air@latest   # if you want live reload
# VS Code: install golang.go, then "Go: Install/Update Tools"
```

---

## 3. Local deterministic quality guardrails

Tools and hooks that catch bugs, style issues, and known vulnerabilities **before** merge — runnable the same way locally and in CI.

### 3.1 Built-in / first-party checks

| Check | Command | Guards against | Source |
| --- | --- | --- | --- |
| Format | `gofmt -l .` / `go fmt` | Drift from canonical style | [gofmt blog](https://go.dev/blog/gofmt) |
| Vet | `go vet ./...` | Suspicious constructs | [golangci govet note](https://golangci-lint.run/docs/welcome/quick-start/), [diagnostics](https://go.dev/doc/diagnostics) |
| Tests | `go test ./...` | Functional regressions | [add a test](https://go.dev/doc/tutorial/add-a-test) |
| Race detector | `go test -race ./...` | Data races (runtime); **supported on darwin/arm64** | [race detector](https://go.dev/doc/articles/race_detector) |
| Module hygiene | `go mod tidy` + commit `go.sum` | Missing/extra deps; checksum authenticity | [getting started](https://go.dev/doc/tutorial/getting-started), [ref/mod](https://go.dev/ref/mod) |

Race detector requirements (official): cgo enabled; on Darwin, ports include `darwin/arm64`. Overhead: memory ~5–10×, time ~2–20× typical.  
Source: [https://go.dev/doc/articles/race_detector](https://go.dev/doc/articles/race_detector)

### 3.2 `govulncheck` (Go vulnerability DB)

Official CLI from the Go security / vuln project:

```bash
go install golang.org/x/vuln/cmd/govulncheck@latest
govulncheck ./...
```

- Surfaces **known** vulnerabilities that **actually affect** your call graph (lower noise than dependency-list-only scanners).  
- Can analyze codebases and binaries.  
- Integrated with pkg.go.dev and VS Code Go extension (per Go vuln docs / blog).  
- GitHub Action exists for CI (mentioned in govulncheck blog).  

Sources:

- [https://go.dev/doc/security/vuln/](https://go.dev/doc/security/vuln/)  
- [https://go.dev/blog/govulncheck](https://go.dev/blog/govulncheck)  
- Module: [https://pkg.go.dev/golang.org/x/vuln/cmd/govulncheck](https://pkg.go.dev/golang.org/x/vuln/cmd/govulncheck)

### 3.3 Staticcheck

First-party docs at staticcheck.io:

```bash
go install honnef.co/go/tools/cmd/staticcheck@latest
staticcheck ./...
```

- Finds bugs, performance issues, simplifications, style; designed for low false positives / CI use.  
- “Behaves just like the official `go` tool”; intended alongside `go vet ./...`.  
- Check classes include **SA** (bugs / stdlib misuse), plus other prefixes documented on the checks page (150+ checks claimed on welcome page).  

Sources:

- [https://staticcheck.io/docs/](https://staticcheck.io/docs/)  
- [https://staticcheck.io/docs/getting-started/](https://staticcheck.io/docs/getting-started/)  
- [https://staticcheck.io/docs/checks/](https://staticcheck.io/docs/checks/)

VS Code Go wiki: **staticcheck is the default lint tool**; alternatives include golangci-lint / revive.  
Source: [https://github.com/golang/vscode-go/wiki/tools](https://github.com/golang/vscode-go/wiki/tools)

### 3.4 golangci-lint (meta-linter)

Fast runner that runs many linters in parallel; YAML config; IDE integrations; “over a hundred linters” (project README).  
Source: [https://github.com/golangci/golangci-lint](https://github.com/golangci/golangci-lint)

**Install (recommended: version-pinned binary)**

```bash
# Example from official local install docs (pin the version you want):
curl -sSfL https://golangci-lint.run/install.sh | sh -s -- -b $(go env GOPATH)/bin v2.12.2

# or Homebrew (Apple Silicon bottles exist; see Homebrew note in §2.5):
brew install golangci-lint
```

**Do not rely on `go install` for production/CI** unless you accept the documented risks (local Go version, dependency upgrades, untested binaries). Maintainers recommend binary install.  
Source: [https://golangci-lint.run/docs/welcome/install/local/](https://golangci-lint.run/docs/welcome/install/local/)

**Run**

```bash
golangci-lint run          # same as ./...
golangci-lint fmt          # formatting via golangci-lint
```

**Default-enabled linters** (zero-config), from quick start:

| Linter | Role (per `golangci-lint help linters` text in docs) |
| --- | --- |
| `errcheck` | Unchecked errors |
| `govet` | Same family as `go vet` |
| `ineffassign` | Ineffectual assignments |
| `staticcheck` | Staticcheck rule set |
| `unused` | Unused consts/vars/funcs/types |

Source: [https://golangci-lint.run/docs/welcome/quick-start/](https://golangci-lint.run/docs/welcome/quick-start/)

**CI reproducibility:** pin a specific release; avoid `linters.default: all` if you need stable CI. Prefer their GitHub Action for GitHub projects.  
Source: [https://golangci-lint.run/docs/welcome/install/ci/](https://golangci-lint.run/docs/welcome/install/ci/)

**Editor:** VS Code settings documented (use `--fast-only` in-editor to avoid freezes). GoLand 2025.1+ has built-in support.  
Source: [https://golangci-lint.run/docs/welcome/integrations/](https://golangci-lint.run/docs/welcome/integrations/)

### 3.5 Pre-commit hooks

**Generic framework:** [pre-commit](https://pre-commit.com/) — install the package manager, add `.pre-commit-config.yaml`, run `pre-commit install`, optionally `pre-commit run --all-files`.  
Source: [https://pre-commit.com/#install](https://pre-commit.com/#install)

**golangci-lint official hooks** (from the project’s `.pre-commit-hooks.yaml`):

| Hook id | Behavior |
| --- | --- |
| `golangci-lint` | `golangci-lint run --new-from-rev HEAD --fix` on Go changes (note: not full-repo; `unused`-style linters may be incomplete) |
| `golangci-lint-full` | `golangci-lint run --fix` on all files (use if pre-commit runs in CI) |
| `golangci-lint-config-verify` | `golangci-lint config verify` when config files change |

Source: [https://raw.githubusercontent.com/golangci/golangci-lint/master/.pre-commit-hooks.yaml](https://raw.githubusercontent.com/golangci/golangci-lint/master/.pre-commit-hooks.yaml)

**Historical gofmt git hook:** the gofmt blog mentions `misc/git/pre-commit` in the Go repository for preventing incorrectly formatted commits (location/availability may change over time — verify in the tree you use).  
Source: [https://go.dev/blog/gofmt](https://go.dev/blog/gofmt)

### 3.6 Suggested local “CI-equivalent” script

Deterministic, no network required beyond module cache / vuln DB update policies of the tools themselves:

```bash
#!/usr/bin/env bash
set -euo pipefail

go test ./...
go test -race ./...          # slower; keep in CI and/or pre-push
go vet ./...
staticcheck ./...            # or rely on golangci-lint's embedded staticcheck
golangci-lint run
govulncheck ./...
# optional: test -z "$(gofmt -l .)" 
```

Pin `golangci-lint` and Go toolchain versions in CI the same as locally (`GOTOOLCHAIN`, `go.mod` `go`/`toolchain` lines).

---

## Summary matrix

| Layer | Install / enable | Primary purpose |
| --- | --- | --- |
| **Baseline** | `darwin-arm64` Go (pkg or Homebrew), `PATH`, modules, `gofmt`/`go vet`/`go test`, gopls + editor | Write, build, navigate, format, basic check |
| **DX** | Delve (+ macOS DevToolsSecurity), VS Code Go tools, Air, pprof/trace | Debug, generate tests, live reload, profile |
| **Guardrails** | Race tests, staticcheck, golangci-lint (pinned), govulncheck, pre-commit | Bugs, unused code, unchecked errors, known vulns, style |

---

## Uncertainties / limits of this research

- Exact **macOS tab** wording on [go.dev/doc/install](https://go.dev/doc/install) is tabbed/JS-rendered; Apple Silicon artifact names and macOS 12+ requirement are taken from [go.dev/dl](https://go.dev/dl/). Generic PATH / `/usr/local/go` steps are from the same install doc.  
- **`goimports`**: still published under `golang.org/x/tools`, but gopls already formats and organizes imports for LSP users; this note does not claim a separate mandatory baseline install.  
- **golangci-lint website** returned intermittent 500s during research; install/quick-start/integrations content above was successfully retrieved from `golangci-lint.run` docs URLs and the GitHub repo.  
- Version numbers (Go 1.26.5, golangci-lint v2.12.2, etc.) are **point-in-time** from primary pages on 2026-08-12 — always re-check release pages before pinning.
