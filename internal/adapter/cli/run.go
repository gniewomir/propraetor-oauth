package cli

import (
	"context"

	"github.com/spf13/cobra"
)

// Run is the Operator CLI entrypoint (composition root). args[0] is the
// program name (as in os.Args).
func Run(ctx context.Context, args []string) error {
	root := newRootCommand()
	root.SetContext(ctx)
	if len(args) > 0 {
		root.SetArgs(args[1:])
	}
	return root.ExecuteContext(ctx)
}

func newRootCommand() *cobra.Command {
	root := &cobra.Command{
		Use:           "oauth",
		Short:         "Propraetor OAuth Authorization Server Operator CLI",
		Long:          "Single Operator-facing binary: server runs the Authorization Server; other commands administer Clients, Users, Scopes, and related state (ADR-0019).",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.AddCommand(
		newServerCommand(),
		newPolicyCommand(),
		newClientCommand(),
		newUserCommand(),
		newScopeCommand(),
		newConsentCommand(),
		newPurgeCommand(),
	)

	return root
}
