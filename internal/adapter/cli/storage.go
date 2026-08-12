package cli

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/gniewomir/propraetor-oauth/internal/adapter/postgres"
	"github.com/spf13/cobra"
)

func newStorageCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "storage",
		Short: "Postgres storage maintenance",
		Long:  "Postgres app-role bootstrap, verify, and Storage schema migrate for Operators (ADR-0072, ADR-0074).",
	}
	cmd.AddCommand(newStorageBootstrapSQLCommand())
	cmd.AddCommand(newStorageVerifyCommand())
	cmd.AddCommand(newStorageMigrateCommand())
	return cmd
}

func newStorageVerifyCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "verify",
		Short: "Check OAUTH_STORAGE_URL and storage connectivity",
		Long:  "Requires OAUTH_STORAGE_URL; runs a connectivity check (SELECT 1). Does not ensure Storage schema migrations yet (ADR-0072).",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			url := strings.TrimSpace(os.Getenv("OAUTH_STORAGE_URL"))
			if url == "" {
				return fmt.Errorf("OAUTH_STORAGE_URL is not set")
			}
			if err := storagePing(cmd.Context(), url); err != nil {
				return fmt.Errorf("storage: connection failed")
			}
			_, err := fmt.Fprintln(cmd.OutOrStdout(), "storage: ok")
			return err
		},
	}
}

func newStorageBootstrapSQLCommand() *cobra.Command {
	var (
		prefix  string
		host    string
		port    int
		sslmode string
	)
	cmd := &cobra.Command{
		Use:   "bootstrap-sql",
		Short: "Print superuser SQL to create the app role and database",
		Long:  "Emits CREATE ROLE / CREATE DATABASE SQL on stdout and OAUTH_STORAGE_URL on stderr (secret-bearing; ADR-0072).",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			password, err := bootstrapPassword()
			if err != nil {
				return fmt.Errorf("generate password: %w", err)
			}
			result, err := postgres.GenerateBootstrap(postgres.BootstrapInput{
				Prefix:   prefix,
				Password: password,
				Host:     host,
				Port:     port,
				SSLMode:  sslmode,
			})
			if err != nil {
				return err
			}
			if _, err := io.WriteString(cmd.OutOrStdout(), result.SQL); err != nil {
				return err
			}
			_, err = fmt.Fprintf(cmd.ErrOrStderr(), "OAUTH_STORAGE_URL=%s\n", result.StorageURL)
			return err
		},
	}
	cmd.Flags().StringVar(&prefix, "prefix", "", "instance discriminator (database and role become {prefix}_oauth)")
	cmd.Flags().StringVar(&host, "host", "localhost", "host embedded in OAUTH_STORAGE_URL")
	cmd.Flags().IntVar(&port, "port", 5432, "port embedded in OAUTH_STORAGE_URL")
	cmd.Flags().StringVar(&sslmode, "sslmode", "disable", "sslmode query param embedded in OAUTH_STORAGE_URL")
	_ = cmd.MarkFlagRequired("prefix")
	return cmd
}

var bootstrapPassword = defaultBootstrapPassword

func defaultBootstrapPassword() (string, error) {
	var b [24]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b[:]), nil
}

func stubBootstrapPassword(fn func() (string, error)) func() {
	prev := bootstrapPassword
	bootstrapPassword = fn
	return func() { bootstrapPassword = prev }
}

var storagePing = postgres.Ping

func stubStoragePing(fn func(context.Context, string) error) func() {
	prev := storagePing
	storagePing = fn
	return func() { storagePing = prev }
}
