// Command oauth is the single Operator-facing binary.
//
// Subcommands select mode (ADR-0019): server runs the Authorization Server;
// other commands perform administration. Admin and server entrypoints are
// adapters over the domain (CLI → domain → persistence), not a second rules
// engine (ADR-0022, ADR-0033). Argv parsing uses Cobra (ADR-0070).
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/gniewomir/propraetor-oauth/internal/adapter/cli"
)

func main() {
	err := cli.Run(context.Background(), os.Args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
