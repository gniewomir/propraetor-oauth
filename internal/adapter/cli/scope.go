package cli

import "github.com/spf13/cobra"

func newScopeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "scope",
		Short: "Manage the Operator Scope catalog",
		Long:  "Defines Scope names in the shared catalog; assign Scopes to Clients with client allow / disallow (ADR-0015).",
	}
	cmd.AddCommand(
		leaf("create", "Create a Scope in the catalog"),
		leaf("deactivate", "Deactivate a Scope in the catalog"),
		leaf("reactivate", "Reactivate a Scope in the catalog"),
		leaf("list", "List Scopes in the catalog"),
	)
	return cmd
}
