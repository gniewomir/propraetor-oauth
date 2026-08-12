package cli

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
)

// ErrNotImplemented is returned by placeholder command Actions until the
// corresponding Operator use case is wired.
var ErrNotImplemented = errors.New("not implemented")

func notImplemented(cmd *cobra.Command, _ []string) error {
	return fmt.Errorf("%s: %w", cmd.CommandPath(), ErrNotImplemented)
}

func leaf(use, short string) *cobra.Command {
	return &cobra.Command{
		Use:   use,
		Short: short,
		Args:  cobra.NoArgs,
		RunE:  notImplemented,
	}
}
