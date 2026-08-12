package postgres_test

import (
	"strings"
	"testing"

	"github.com/gniewomir/propraetor-oauth/internal/adapter/postgres"
)

func TestGenerateBootstrapSQL(t *testing.T) {
	t.Parallel()

	got, err := postgres.GenerateBootstrap(postgres.BootstrapInput{
		Prefix:   "dev",
		Password: "s3cret-pass",
		Host:     "127.0.0.1",
		Port:     55432,
	})
	if err != nil {
		t.Fatalf("GenerateBootstrap: %v", err)
	}

	if !strings.Contains(got.SQL, `CREATE ROLE "dev_oauth"`) {
		t.Errorf("SQL missing CREATE ROLE:\n%s", got.SQL)
	}
	if !strings.Contains(got.SQL, `PASSWORD 's3cret-pass'`) {
		t.Errorf("SQL missing password literal:\n%s", got.SQL)
	}
	if !strings.Contains(got.SQL, `CREATE DATABASE "dev_oauth" OWNER "dev_oauth"`) {
		t.Errorf("SQL missing CREATE DATABASE:\n%s", got.SQL)
	}

	wantURL := "postgres://dev_oauth:s3cret-pass@127.0.0.1:55432/dev_oauth?sslmode=disable"
	if got.StorageURL != wantURL {
		t.Errorf("StorageURL = %q, want %q", got.StorageURL, wantURL)
	}
}

func TestGenerateBootstrapSQLEscapesPasswordAndRejectsBadPrefix(t *testing.T) {
	t.Parallel()

	got, err := postgres.GenerateBootstrap(postgres.BootstrapInput{
		Prefix:   "test",
		Password: `o'brien`,
		Host:     "localhost",
		Port:     5432,
	})
	if err != nil {
		t.Fatalf("GenerateBootstrap: %v", err)
	}
	if !strings.Contains(got.SQL, `PASSWORD 'o''brien'`) {
		t.Errorf("SQL should escape single quotes in password:\n%s", got.SQL)
	}
	if !strings.Contains(got.StorageURL, "o%27brien") {
		t.Errorf("StorageURL should percent-encode password: %q", got.StorageURL)
	}

	_, err = postgres.GenerateBootstrap(postgres.BootstrapInput{
		Prefix:   "bad-prefix",
		Password: "x",
		Host:     "localhost",
		Port:     5432,
	})
	if err == nil {
		t.Fatal("expected error for prefix with hyphen")
	}
}
