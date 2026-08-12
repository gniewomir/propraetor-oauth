package e2e_test

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func oauthBin(t *testing.T) string {
	t.Helper()
	if b := os.Getenv("OAUTH_BIN"); b != "" {
		return b
	}
	candidates := []string{
		filepath.Join("bin", "oauth"),
		filepath.Join("..", "bin", "oauth"),
	}
	for _, c := range candidates {
		if st, err := os.Stat(c); err == nil && !st.IsDir() {
			return c
		}
	}
	t.Fatal("oauth binary not found; set OAUTH_BIN or run ./scripts/test-e2e.sh")
	return ""
}

func runOAuth(t *testing.T, env []string, args ...string) (stdout, stderr string, exitCode int) {
	t.Helper()
	cmd := exec.Command(oauthBin(t), args...)
	cmd.Env = env
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	err := cmd.Run()
	if err != nil {
		var ee *exec.ExitError
		if !errors.As(err, &ee) {
			t.Fatalf("run oauth: %v", err)
		}
		return outBuf.String(), errBuf.String(), ee.ExitCode()
	}
	return outBuf.String(), errBuf.String(), 0
}

func baseEnv() []string {
	out := make([]string, 0, len(os.Environ()))
	for _, e := range os.Environ() {
		if strings.HasPrefix(e, "OAUTH_STORAGE_URL=") {
			continue
		}
		out = append(out, e)
	}
	return out
}

func TestVerifyMissingURL(t *testing.T) {
	t.Parallel()

	stdout, stderr, code := runOAuth(t, baseEnv(), "storage", "verify")
	if code == 0 {
		t.Fatalf("exit = 0, want non-zero; stdout=%q stderr=%q", stdout, stderr)
	}
	if !strings.Contains(stderr, "OAUTH_STORAGE_URL is not set") {
		t.Fatalf("stderr = %q, want OAUTH_STORAGE_URL is not set", stderr)
	}
}

func TestVerifyBadConnection(t *testing.T) {
	t.Parallel()

	env := append(baseEnv(), "OAUTH_STORAGE_URL=postgres://u:p@127.0.0.1:1/db?sslmode=disable")
	stdout, stderr, code := runOAuth(t, env, "storage", "verify")
	if code == 0 {
		t.Fatalf("exit = 0, want non-zero; stdout=%q stderr=%q", stdout, stderr)
	}
	if !strings.Contains(stderr, "storage: connection failed") {
		t.Fatalf("stderr = %q, want storage: connection failed", stderr)
	}
}

func TestVerifyOK(t *testing.T) {
	t.Parallel()

	url := strings.TrimSpace(os.Getenv("OAUTH_STORAGE_URL"))
	if url == "" {
		t.Fatal("OAUTH_STORAGE_URL is required for TestVerifyOK (run via ./scripts/test-e2e.sh)")
	}
	env := append(baseEnv(), "OAUTH_STORAGE_URL="+url)
	stdout, stderr, code := runOAuth(t, env, "storage", "verify")
	if code != 0 {
		t.Fatalf("exit = %d, want 0; stdout=%q stderr=%q", code, stdout, stderr)
	}
	if !strings.Contains(stdout, "storage: ok") {
		t.Fatalf("stdout = %q, want storage: ok", stdout)
	}
}
