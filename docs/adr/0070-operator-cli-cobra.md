# Operator CLI uses Cobra (no Viper)

The Operator binary parses argv and subcommands with `github.com/spf13/cobra` (and its required `pflag`). Chosen over stdlib `flag` (no maintained command-tree/help model for the planned admin surface) and over `urfave/cli` (lighter, but Cobra’s ecosystem and Operator UX were preferred once the trade-off was accepted). This is an allowed external module under ADR-0035: CLI parsing is necessary for ADR-0019’s single-binary command tree; Resource Owner screen purity is unchanged.

Composition root and command tree live in `internal/adapter/cli`; `cmd/oauth` stays a thin `main` (ADR-0033). Do **not** adopt Viper: Server Policy remains a required closed TOML file for `server`, topology stays on CLI flags, and secrets stay in the environment (ADR-0068, 0022). Flag spellings and admin verbs remain governed by existing CLI ADRs (e.g. 0060, 0067, 0069), not by this choice of library.
