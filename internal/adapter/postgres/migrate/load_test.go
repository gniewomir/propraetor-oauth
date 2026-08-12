package migrate_test

import (
	"testing"
	"testing/fstest"

	"github.com/gniewomir/propraetor-oauth/internal/adapter/postgres/migrate"
)

func TestHeadEmptyFS(t *testing.T) {
	t.Parallel()

	head, err := migrate.Head(fstest.MapFS{})
	if err != nil {
		t.Fatalf("Head: %v", err)
	}
	if head != 0 {
		t.Fatalf("head = %d, want 0", head)
	}
}

func TestHeadReadsHighestPairedVersion(t *testing.T) {
	t.Parallel()

	fsys := fstest.MapFS{
		"0001_init.up.sql":   &fstest.MapFile{Data: []byte("SELECT 1;")},
		"0001_init.down.sql": &fstest.MapFile{Data: []byte("SELECT 1;")},
		"0002_more.up.sql":   &fstest.MapFile{Data: []byte("SELECT 1;")},
		"0002_more.down.sql": &fstest.MapFile{Data: []byte("SELECT 1;")},
	}
	head, err := migrate.Head(fsys)
	if err != nil {
		t.Fatalf("Head: %v", err)
	}
	if head != 2 {
		t.Fatalf("head = %d, want 2", head)
	}
}

func TestHeadMissingTwinFatal(t *testing.T) {
	t.Parallel()

	fsys := fstest.MapFS{
		"0001_init.up.sql": &fstest.MapFile{Data: []byte("SELECT 1;")},
	}
	_, err := migrate.Head(fsys)
	if err == nil {
		t.Fatal("expected error for missing down twin")
	}
	if got := err.Error(); got != "storage: migrations invalid" {
		t.Fatalf("error = %q, want %q", got, "storage: migrations invalid")
	}
}

func TestHeadGapInVersionsFatal(t *testing.T) {
	t.Parallel()

	fsys := fstest.MapFS{
		"0001_a.up.sql":   &fstest.MapFile{Data: []byte("SELECT 1;")},
		"0001_a.down.sql": &fstest.MapFile{Data: []byte("SELECT 1;")},
		"0003_c.up.sql":   &fstest.MapFile{Data: []byte("SELECT 1;")},
		"0003_c.down.sql": &fstest.MapFile{Data: []byte("SELECT 1;")},
	}
	_, err := migrate.Head(fsys)
	if err == nil {
		t.Fatal("expected error for version gap")
	}
	if got := err.Error(); got != "storage: migrations invalid" {
		t.Fatalf("error = %q, want %q", got, "storage: migrations invalid")
	}
}

func TestHeadIgnoresNonMigrationFiles(t *testing.T) {
	t.Parallel()

	fsys := fstest.MapFS{
		".gitkeep":        &fstest.MapFile{Data: []byte{}},
		"README.md":       &fstest.MapFile{Data: []byte("x")},
		"0001_a.up.sql":   &fstest.MapFile{Data: []byte("SELECT 1;")},
		"0001_a.down.sql": &fstest.MapFile{Data: []byte("SELECT 1;")},
	}
	head, err := migrate.Head(fsys)
	if err != nil {
		t.Fatalf("Head: %v", err)
	}
	if head != 1 {
		t.Fatalf("head = %d, want 1", head)
	}
}

func TestHeadIgnoresNonFourDigitPrefix(t *testing.T) {
	t.Parallel()

	fsys := fstest.MapFS{
		"1_init.up.sql":   &fstest.MapFile{Data: []byte("SELECT 1;")},
		"1_init.down.sql": &fstest.MapFile{Data: []byte("SELECT 1;")},
	}
	head, err := migrate.Head(fsys)
	if err != nil {
		t.Fatalf("Head: %v", err)
	}
	if head != 0 {
		t.Fatalf("head = %d, want 0 (non-NNNN names ignored)", head)
	}
}
