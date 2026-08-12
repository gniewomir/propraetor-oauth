package integration_test

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gniewomir/propraetor-oauth/internal/adapter/postgres"
)

func TestPingOK(t *testing.T) {
	t.Parallel()

	url := strings.TrimSpace(os.Getenv("OAUTH_STORAGE_URL"))
	if url == "" {
		t.Fatal("OAUTH_STORAGE_URL is required (run via ./scripts/test-integration.sh)")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := postgres.Ping(ctx, url); err != nil {
		t.Fatalf("Ping: %v", err)
	}
}
