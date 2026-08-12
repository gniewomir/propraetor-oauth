package cli

import "github.com/spf13/cobra"

func newUserCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "user",
		Short: "Manage Users (End-Users)",
	}
	cmd.AddCommand(
		leaf("create", "Create a User"),
		leaf("delete", "Delete a User"),
		leaf("list", "List Users"),
		leaf("show", "Show a User"),
		leaf("set-password", "Set a User password (ADR-0062)"),
		leaf("set-not-before", "Advance User Not-Before watermark (ADR-0069)"),
	)
	return cmd
}
