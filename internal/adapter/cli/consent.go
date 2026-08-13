package cli

import "github.com/spf13/cobra"

func newConsentCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "consent",
		Short: "Manage Consent Grants",
	}
	cmd.AddCommand(
		&cobra.Command{
			Use:   "deactivate",
			Short: "Deactivate Consent Grants for a User, User×Client, or User×Client×Scope",
			Long:  "Deactivates stored Consent Grants so the next Authorization Request may show Consent again (ADR-0055). Not token invalidation and not Not-Before.",
			Args:  cobra.NoArgs,
			RunE:  notImplemented,
		},
		&cobra.Command{
			Use:   "reactivate",
			Short: "Reactivate Consent Grants for a User, User×Client, or User×Client×Scope",
			Long:  "Reactivates previously deactivated Consent Grants (ADR-0055). Distinct from End-User Consent UI grant. Not token invalidation and not Not-Before.",
			Args:  cobra.NoArgs,
			RunE:  notImplemented,
		},
	)
	return cmd
}
