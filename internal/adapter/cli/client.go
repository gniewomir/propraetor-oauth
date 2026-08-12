package cli

import "github.com/spf13/cobra"

func newClientCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "client",
		Short: "Manage Clients and per-Client Scope allowlists",
	}
	cmd.AddCommand(
		leaf("create", "Create a Client"),
		leaf("delete", "Delete a Client"),
		leaf("list", "List Clients"),
		leaf("show", "Show a Client"),
		leaf("scope-add", "Add a Scope to a Client allowlist"),
		leaf("scope-remove", "Remove a Scope from a Client allowlist"),
		leaf("set-not-before", "Advance Client Not-Before watermark (ADR-0069)"),
	)
	return cmd
}
