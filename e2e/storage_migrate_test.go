package e2e_test

import (
	"os"
	"strings"
	"testing"
)

func TestMigrateRequiresDirection(t *testing.T) {
	t.Parallel()

	url := strings.TrimSpace(os.Getenv("OAUTH_STORAGE_URL"))
	env := append(baseEnv(), "OAUTH_STORAGE_URL="+url, "OAUTH_STORAGE_ENV=test")
	stdout, stderr, code := runOAuth(t, env, "storage", "migrate")
	if code == 0 {
		t.Fatalf("exit = 0; stdout=%q stderr=%q", stdout, stderr)
	}
	if !strings.Contains(stderr, "storage: migrate requires --up or --down") {
		t.Fatalf("stderr = %q", stderr)
	}
}

func TestMigrateDownRefusedOutsideDevTest(t *testing.T) {
	t.Parallel()

	url := strings.TrimSpace(os.Getenv("OAUTH_STORAGE_URL"))
	if url == "" {
		t.Fatal("OAUTH_STORAGE_URL required")
	}
	env := append(baseEnv(), "OAUTH_STORAGE_URL="+url, "OAUTH_STORAGE_ENV=production")
	stdout, stderr, code := runOAuth(t, env, "storage", "migrate", "--down")
	if code == 0 {
		t.Fatalf("exit = 0; stdout=%q stderr=%q", stdout, stderr)
	}
	if !strings.Contains(stderr, "storage: migrate --down refused") {
		t.Fatalf("stderr = %q", stderr)
	}
}

func TestMigrateUpAndDownClients(t *testing.T) {
	// Serial: mutates Storage schema on shared e2e DB.
	url := strings.TrimSpace(os.Getenv("OAUTH_STORAGE_URL"))
	if url == "" {
		t.Fatal("OAUTH_STORAGE_URL required")
	}
	env := append(baseEnv(), "OAUTH_STORAGE_URL="+url, "OAUTH_STORAGE_ENV=test")

	stdout, stderr, code := runOAuth(t, env, "storage", "migrate", "--up")
	if code != 0 {
		t.Fatalf("up exit = %d stdout=%q stderr=%q", code, stdout, stderr)
	}
	if !strings.Contains(stdout, "storage: schema 1") {
		t.Fatalf("up stdout = %q", stdout)
	}

	stdout, stderr, code = runOAuth(t, env, "storage", "migrate", "--down")
	if code != 0 {
		t.Fatalf("down exit = %d stdout=%q stderr=%q", code, stdout, stderr)
	}
	if !strings.Contains(stdout, "storage: schema 0") {
		t.Fatalf("down stdout = %q", stdout)
	}
}
