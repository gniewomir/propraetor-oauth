package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/gniewomir/propraetor-oauth/internal/adapter/postgres"
)

func TestPingRejectsEmptyURL(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	err := postgres.Ping(ctx, "")
	if err == nil {
		t.Fatal("expected error for empty URL")
	}
}

func TestPingFailsUnreachable(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Valid URL shape, nothing listening.
	err := postgres.Ping(ctx, "postgres://u:p@127.0.0.1:1/db?sslmode=disable")
	if err == nil {
		t.Fatal("expected connection error")
	}
}
