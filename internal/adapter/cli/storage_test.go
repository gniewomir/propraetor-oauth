package cli

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func TestStorageBootstrapSQLRequiresPrefix(t *testing.T) {
	t.Parallel()

	root := newRootCommand()
	root.SetArgs([]string{"storage", "bootstrap-sql"})
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error without --prefix")
	}
}

func TestStorageBootstrapSQLEmitsSQLAndStorageURL(t *testing.T) {
	restore := stubBootstrapPassword(func() (string, error) {
		return "fixed-pass", nil
	})
	defer restore()

	root := newRootCommand()
	var stdout, stderr bytes.Buffer
	root.SetOut(&stdout)
	root.SetErr(&stderr)
	root.SetArgs([]string{
		"storage", "bootstrap-sql",
		"--prefix", "dev",
		"--host", "127.0.0.1",
		"--port", "55432",
	})

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	sql := stdout.String()
	if !strings.Contains(sql, `CREATE ROLE "dev_oauth"`) {
		t.Errorf("stdout SQL missing role:\n%s", sql)
	}
	if strings.Contains(sql, "OAUTH_STORAGE_URL=") {
		t.Errorf("URL must not be on stdout (so SQL can pipe to psql):\n%s", sql)
	}

	errOut := stderr.String()
	wantLine := "OAUTH_STORAGE_URL=postgres://dev_oauth:fixed-pass@127.0.0.1:55432/dev_oauth?sslmode=disable"
	if !strings.Contains(errOut, wantLine) {
		t.Errorf("stderr missing URL line %q:\n%s", wantLine, errOut)
	}
}

func TestStorageBootstrapSQLSSLMode(t *testing.T) {
	restore := stubBootstrapPassword(func() (string, error) {
		return "fixed-pass", nil
	})
	defer restore()

	root := newRootCommand()
	var stderr bytes.Buffer
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&stderr)
	root.SetArgs([]string{
		"storage", "bootstrap-sql",
		"--prefix", "dev",
		"--sslmode", "require",
	})

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !strings.Contains(stderr.String(), "sslmode=require") {
		t.Errorf("stderr should honor --sslmode: %s", stderr.String())
	}
}

func TestStorageBootstrapSQLPasswordError(t *testing.T) {
	restore := stubBootstrapPassword(func() (string, error) {
		return "", errors.New("rng failed")
	})
	defer restore()

	root := newRootCommand()
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"storage", "bootstrap-sql", "--prefix", "dev"})

	if err := root.Execute(); err == nil {
		t.Fatal("expected error when password generation fails")
	}
}
