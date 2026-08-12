// Command oauth is the single Operator-facing binary.
//
// Subcommands select mode (ADR-0019): server runs the Authorization Server;
// other commands perform administration. Admin and server entrypoints are
// adapters over the domain (CLI → domain → persistence), not a second rules
// engine (ADR-0022, ADR-0033).
package main

import (
	"fmt"
	"os"
)

func main() {
	// Composition root will live in internal/adapter/cli; this main stays thin.
	fmt.Fprintln(os.Stderr, "oauth: not implemented")
	os.Exit(1)
}
