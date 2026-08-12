package integration_test

import (
	"context"
	"database/sql"
	"os"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/gniewomir/propraetor-oauth/internal/adapter/postgres/migrate"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func openStorage(t *testing.T) *sql.DB {
	t.Helper()
	url := strings.TrimSpace(os.Getenv("OAUTH_STORAGE_URL"))
	if url == "" {
		t.Fatal("OAUTH_STORAGE_URL is required (run via ./scripts/test-integration.sh)")
	}
	db, err := sql.Open("pgx", url)
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("Ping: %v", err)
	}
	return db
}

func migrationPairFS() fstest.MapFS {
	return fstest.MapFS{
		"0001_demo.up.sql":   &fstest.MapFile{Data: []byte(`CREATE TABLE migrate_demo (id INT PRIMARY KEY);`)},
		"0001_demo.down.sql": &fstest.MapFile{Data: []byte(`DROP TABLE IF EXISTS migrate_demo;`)},
		"0002_demo.up.sql":   &fstest.MapFile{Data: []byte(`ALTER TABLE migrate_demo ADD COLUMN label TEXT;`)},
		"0002_demo.down.sql": &fstest.MapFile{Data: []byte(`ALTER TABLE migrate_demo DROP COLUMN label;`)},
	}
}

func cleanupMigrateArtifacts(t *testing.T, db *sql.DB) {
	t.Helper()
	ctx := context.Background()
	_, _ = db.ExecContext(ctx, `DROP TABLE IF EXISTS migrate_demo`)
	_, _ = db.ExecContext(ctx, `DROP TABLE IF EXISTS storage_schema_migrations`)
}

func TestMigrateUpDownRoundTrip(t *testing.T) {
	db := openStorage(t)
	cleanupMigrateArtifacts(t, db)
	t.Cleanup(func() { cleanupMigrateArtifacts(t, db) })

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	r := migrate.New(db, migrationPairFS())

	ver, err := r.Up(ctx)
	if err != nil {
		t.Fatalf("Up: %v", err)
	}
	if ver != 2 {
		t.Fatalf("after Up version = %d, want 2", ver)
	}

	var labelExists bool
	err = db.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM information_schema.columns
			WHERE table_name = 'migrate_demo' AND column_name = 'label'
		)`).Scan(&labelExists)
	if err != nil || !labelExists {
		t.Fatalf("label column missing after Up: exists=%v err=%v", labelExists, err)
	}

	ver, err = r.Down(ctx)
	if err != nil {
		t.Fatalf("Down: %v", err)
	}
	if ver != 1 {
		t.Fatalf("after Down version = %d, want 1", ver)
	}

	ver, err = r.Down(ctx)
	if err != nil {
		t.Fatalf("Down to 0: %v", err)
	}
	if ver != 0 {
		t.Fatalf("after second Down version = %d, want 0", ver)
	}

	var tableExists bool
	err = db.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM information_schema.tables
			WHERE table_schema = 'public' AND table_name = 'migrate_demo'
		)`).Scan(&tableExists)
	if err != nil || tableExists {
		t.Fatalf("migrate_demo should be gone: exists=%v err=%v", tableExists, err)
	}

	var recordExists bool
	err = db.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM information_schema.tables
			WHERE table_schema = 'public' AND table_name = 'storage_schema_migrations'
		)`).Scan(&recordExists)
	if err != nil || !recordExists {
		t.Fatalf("storage_schema_migrations should remain: exists=%v err=%v", recordExists, err)
	}
}

func TestMigrateUpWhenAlreadyAtHead(t *testing.T) {
	db := openStorage(t)
	cleanupMigrateArtifacts(t, db)
	t.Cleanup(func() { cleanupMigrateArtifacts(t, db) })

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	r := migrate.New(db, migrationPairFS())
	if _, err := r.Up(ctx); err != nil {
		t.Fatalf("Up: %v", err)
	}
	ver, err := r.Up(ctx)
	if err != nil {
		t.Fatalf("second Up: %v", err)
	}
	if ver != 2 {
		t.Fatalf("version = %d, want 2", ver)
	}
}

func TestMigrateDownAtZero(t *testing.T) {
	db := openStorage(t)
	cleanupMigrateArtifacts(t, db)
	t.Cleanup(func() { cleanupMigrateArtifacts(t, db) })

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	r := migrate.New(db, fstest.MapFS{})
	ver, err := r.Down(ctx)
	if err != nil {
		t.Fatalf("Down: %v", err)
	}
	if ver != 0 {
		t.Fatalf("version = %d, want 0", ver)
	}
}

func TestMigrateRejectsAppliedAheadOfHead(t *testing.T) {
	db := openStorage(t)
	cleanupMigrateArtifacts(t, db)
	t.Cleanup(func() { cleanupMigrateArtifacts(t, db) })

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	r := migrate.New(db, migrationPairFS())
	if _, err := r.Up(ctx); err != nil {
		t.Fatalf("Up: %v", err)
	}

	// Binary with only version 1 while DB is at 2.
	smaller := fstest.MapFS{
		"0001_demo.up.sql":   &fstest.MapFile{Data: []byte(`CREATE TABLE migrate_demo (id INT PRIMARY KEY);`)},
		"0001_demo.down.sql": &fstest.MapFile{Data: []byte(`DROP TABLE IF EXISTS migrate_demo;`)},
	}
	r2 := migrate.New(db, smaller)
	_, err := r2.Up(ctx)
	if err == nil {
		t.Fatal("expected error when applied ahead of head")
	}
	if !strings.Contains(err.Error(), "storage: migrate failed") {
		t.Fatalf("error = %v, want storage: migrate failed", err)
	}
}
