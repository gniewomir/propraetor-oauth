package cli

import "github.com/spf13/cobra"

func newPurgeCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "purge",
		Short: "Purge expired entities and/or Audit Events",
		Long:  "Requires --target=entities|audit|all and --older-than each run (ADR-0060). Not run automatically by server.",
		Args:  cobra.NoArgs,
		RunE:  notImplemented,
	}
}
