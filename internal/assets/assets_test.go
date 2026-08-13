package assets_test

import (
	"io/fs"
	"testing"

	"github.com/gniewomir/propraetor-oauth/internal/adapter/postgres/migrate"
	"github.com/gniewomir/propraetor-oauth/internal/assets"
)

func TestMigrationsFSHeadIsClients(t *testing.T) {
	t.Parallel()

	fsys := assets.Migrations()
	head, err := migrate.Head(fsys)
	if err != nil {
		t.Fatalf("Head: %v", err)
	}
	if head != 1 {
		t.Fatalf("head = %d, want 1", head)
	}

	for _, name := range []string{"0001_clients.up.sql", "0001_clients.down.sql"} {
		if _, err := fs.Stat(fsys, name); err != nil {
			t.Fatalf("Stat %s: %v", name, err)
		}
	}
}
