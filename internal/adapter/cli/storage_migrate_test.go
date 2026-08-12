package cli

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
)

func TestStorageMigrateRequiresDirection(t *testing.T) {
	root := newRootCommand()
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"storage", "migrate"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error without --up/--down")
	}
	if !strings.Contains(err.Error(), "storage: migrate requires --up or --down") {
		t.Fatalf("error = %v", err)
	}
}

func TestStorageMigrateBothFlagsRejected(t *testing.T) {
	root := newRootCommand()
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"storage", "migrate", "--up", "--down"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error with both flags")
	}
	if !strings.Contains(err.Error(), "storage: migrate requires --up or --down") {
		t.Fatalf("error = %v", err)
	}
}

func TestStorageMigrateDownRefusedOutsideDevTest(t *testing.T) {
	t.Setenv("OAUTH_STORAGE_ENV", "production")
	t.Setenv("OAUTH_STORAGE_URL", "postgres://ok")

	root := newRootCommand()
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"storage", "migrate", "--down"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected refuse")
	}
	if err.Error() != "storage: migrate --down refused" {
		t.Fatalf("error = %q", err.Error())
	}
}

func TestStorageMigrateDownRefusedWhenEnvUnset(t *testing.T) {
	t.Setenv("OAUTH_STORAGE_ENV", "")
	t.Setenv("OAUTH_STORAGE_URL", "postgres://ok")

	root := newRootCommand()
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"storage", "migrate", "--down"})

	err := root.Execute()
	if err == nil || err.Error() != "storage: migrate --down refused" {
		t.Fatalf("error = %v", err)
	}
}

func TestStorageMigrateUpOK(t *testing.T) {
	t.Setenv("OAUTH_STORAGE_URL", "postgres://ok")
	restore := stubStorageMigrate(func(context.Context, string, bool) (int, error) {
		return 0, nil
	})
	defer restore()

	root := newRootCommand()
	var stdout bytes.Buffer
	root.SetOut(&stdout)
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"storage", "migrate", "--up"})

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if got := strings.TrimSpace(stdout.String()); got != "storage: schema 0" {
		t.Fatalf("stdout = %q, want %q", got, "storage: schema 0")
	}
}

func TestStorageMigrateDownOKInTestEnv(t *testing.T) {
	t.Setenv("OAUTH_STORAGE_ENV", "test")
	t.Setenv("OAUTH_STORAGE_URL", "postgres://ok")
	restore := stubStorageMigrate(func(_ context.Context, _ string, up bool) (int, error) {
		if up {
			t.Fatal("expected down")
		}
		return 0, nil
	})
	defer restore()

	root := newRootCommand()
	var stdout bytes.Buffer
	root.SetOut(&stdout)
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"storage", "migrate", "--down"})

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if got := strings.TrimSpace(stdout.String()); got != "storage: schema 0" {
		t.Fatalf("stdout = %q", got)
	}
}

func TestStorageMigrateRequiresURL(t *testing.T) {
	t.Setenv("OAUTH_STORAGE_URL", "")
	t.Setenv("OAUTH_STORAGE_ENV", "test")

	root := newRootCommand()
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"storage", "migrate", "--up"})

	err := root.Execute()
	if err == nil || !strings.Contains(err.Error(), "OAUTH_STORAGE_URL is not set") {
		t.Fatalf("error = %v", err)
	}
}

func TestStorageMigrateMapsFailed(t *testing.T) {
	t.Setenv("OAUTH_STORAGE_URL", "postgres://ok")
	restore := stubStorageMigrate(func(context.Context, string, bool) (int, error) {
		return 0, errors.New("boom")
	})
	defer restore()

	root := newRootCommand()
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"storage", "migrate", "--up"})

	err := root.Execute()
	if err == nil || err.Error() != "storage: migrate failed" {
		t.Fatalf("error = %v", err)
	}
}
