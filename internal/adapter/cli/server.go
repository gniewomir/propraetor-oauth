package cli

import "github.com/spf13/cobra"

func newServerCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "server",
		Short: "Run the Authorization Server process",
		Long:  "Requires a Server Policy file and process topology flags (ADR-0068). Production-risk allowances need --i-understand-production-risk (ADR-0067).",
		Args:  cobra.NoArgs,
		RunE:  notImplemented,
	}
}
