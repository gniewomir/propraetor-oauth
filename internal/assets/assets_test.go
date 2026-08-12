package assets_test

import (
	"io/fs"
	"testing"

	"github.com/gniewomir/propraetor-oauth/internal/adapter/postgres/migrate"
	"github.com/gniewomir/propraetor-oauth/internal/assets"
)

func TestMigrationsFSHeadIsZero(t *testing.T) {
	t.Parallel()

	head, err := migrate.Head(assets.Migrations())
	if err != nil {
		t.Fatalf("Head: %v", err)
	}
	if head != 0 {
		t.Fatalf("head = %d, want 0 (no versioned SQL yet)", head)
	}

	// Ensure the FS is readable (embed of empty/keep tree).
	_, err = fs.ReadDir(assets.Migrations(), ".")
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
}
