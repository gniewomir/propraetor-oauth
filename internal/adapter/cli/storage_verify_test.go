package cli

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
)

func TestStorageVerifyRequiresURL(t *testing.T) {
	t.Setenv("OAUTH_STORAGE_URL", "")

	root := newRootCommand()
	var stdout, stderr bytes.Buffer
	root.SetOut(&stdout)
	root.SetErr(&stderr)
	root.SetArgs([]string{"storage", "verify"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when OAUTH_STORAGE_URL is unset")
	}
	if !strings.Contains(err.Error(), "OAUTH_STORAGE_URL is not set") {
		t.Fatalf("error = %v, want substring OAUTH_STORAGE_URL is not set", err)
	}
}

func TestStorageVerifyConnectionFailed(t *testing.T) {
	t.Setenv("OAUTH_STORAGE_URL", "postgres://u:p@127.0.0.1:1/db?sslmode=disable")
	restore := stubStoragePing(func(context.Context, string) error {
		return errors.New("dial failed")
	})
	defer restore()

	root := newRootCommand()
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"storage", "verify"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected connection error")
	}
	if !strings.Contains(err.Error(), "storage: connection failed") {
		t.Fatalf("error = %v, want storage: connection failed", err)
	}
}

func TestStorageVerifyOK(t *testing.T) {
	t.Setenv("OAUTH_STORAGE_URL", "postgres://ok")
	restore := stubStoragePing(func(context.Context, string) error {
		return nil
	})
	defer restore()

	root := newRootCommand()
	var stdout bytes.Buffer
	root.SetOut(&stdout)
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"storage", "verify"})

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if got := strings.TrimSpace(stdout.String()); got != "storage: ok" {
		t.Fatalf("stdout = %q, want %q", got, "storage: ok")
	}
}

func TestStorageVerifyWhitespaceURLTreatedAsMissing(t *testing.T) {
	t.Setenv("OAUTH_STORAGE_URL", "   ")

	root := newRootCommand()
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"storage", "verify"})

	err := root.Execute()
	if err == nil || !strings.Contains(err.Error(), "OAUTH_STORAGE_URL is not set") {
		t.Fatalf("error = %v, want missing URL", err)
	}
}
