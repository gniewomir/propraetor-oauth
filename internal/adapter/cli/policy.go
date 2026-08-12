package cli

import "github.com/spf13/cobra"

func newPolicyCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "policy",
		Short: "Server Policy file helpers",
		Long:  "Generate a starter Server Policy TOML for Operator review (ADR-0071). Does not supply runtime defaults for server start (ADR-0068).",
	}
	cmd.AddCommand(&cobra.Command{
		Use:   "init",
		Short: "Write a starter Server Policy TOML document",
		Args:  cobra.NoArgs,
		RunE:  notImplemented,
	})
	return cmd
}
