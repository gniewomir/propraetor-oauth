package cli

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/gniewomir/propraetor-oauth/internal/adapter/postgres/migrate"
	"github.com/gniewomir/propraetor-oauth/internal/assets"
	"github.com/spf13/cobra"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func newStorageMigrateCommand() *cobra.Command {
	var up, down bool
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Apply or roll back Storage schema migrations",
		Long:  "Requires --up or --down. --down only when OAUTH_STORAGE_ENV is dev or test (ADR-0072, ADR-0074).",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if up == down {
				return fmt.Errorf("storage: migrate requires --up or --down")
			}
			if down && !storageEnvAllowsDown() {
				return fmt.Errorf("storage: migrate --down refused")
			}
			url := strings.TrimSpace(os.Getenv("OAUTH_STORAGE_URL"))
			if url == "" {
				return fmt.Errorf("OAUTH_STORAGE_URL is not set")
			}
			ver, err := storageMigrate(cmd.Context(), url, up)
			if err != nil {
				return mapMigrateErr(err)
			}
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "storage: schema %d\n", ver)
			return err
		},
	}
	cmd.Flags().BoolVar(&up, "up", false, "advance Storage schema to embedded head")
	cmd.Flags().BoolVar(&down, "down", false, "roll back one Storage schema version (dev/test only)")
	return cmd
}

func storageEnvAllowsDown() bool {
	switch strings.TrimSpace(os.Getenv("OAUTH_STORAGE_ENV")) {
	case "dev", "test":
		return true
	default:
		return false
	}
}

func mapMigrateErr(err error) error {
	if errors.Is(err, migrate.ErrMigrationsInvalid) {
		return migrate.ErrMigrationsInvalid
	}
	if errors.Is(err, migrate.ErrMigrateFailed) {
		return migrate.ErrMigrateFailed
	}
	return fmt.Errorf("%w", migrate.ErrMigrateFailed)
}

// storageMigrate opens storage and runs Up or one Down step. Overridable in tests.
var storageMigrate = defaultStorageMigrate

func defaultStorageMigrate(ctx context.Context, storageURL string, up bool) (int, error) {
	db, err := sql.Open("pgx", storageURL)
	if err != nil {
		return 0, fmt.Errorf("%w: %v", migrate.ErrMigrateFailed, err)
	}
	defer func() { _ = db.Close() }()

	if err := db.PingContext(ctx); err != nil {
		return 0, fmt.Errorf("%w: %v", migrate.ErrMigrateFailed, err)
	}

	r := migrate.New(db, assets.Migrations())
	if up {
		return r.Up(ctx)
	}
	return r.Down(ctx)
}

func stubStorageMigrate(fn func(context.Context, string, bool) (int, error)) func() {
	prev := storageMigrate
	storageMigrate = fn
	return func() { storageMigrate = prev }
}
