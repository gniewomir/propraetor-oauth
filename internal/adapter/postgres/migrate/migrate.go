package migrate

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"sort"
)

// ErrMigrateFailed is the locked Operator-facing cause for apply/state failures (ADR-0074).
var ErrMigrateFailed = fmt.Errorf("storage: migrate failed")

// advisoryLockKey is a process-wide Postgres advisory lock for Storage schema migrate.
const advisoryLockKey int64 = 0x70726f705f6d6967 // "prop_mig"

const ensureRecordSQL = `
CREATE TABLE IF NOT EXISTS storage_schema_migrations (
	version INT PRIMARY KEY,
	applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
`

// Runner advances Storage schema versions against a database (ADR-0074).
type Runner struct {
	db   *sql.DB
	fsys fs.FS
}

// New returns a Runner that reads paired migration SQL from fsys (files at FS root).
func New(db *sql.DB, fsys fs.FS) *Runner {
	return &Runner{db: db, fsys: fsys}
}

// Up applies all pending migrations through embedded head and returns the new current version.
func (r *Runner) Up(ctx context.Context) (int, error) {
	if err := r.ensureRecord(ctx); err != nil {
		return 0, err
	}
	var current int
	err := r.withLock(ctx, func() error {
		var err error
		current, err = r.upLocked(ctx)
		return err
	})
	return current, err
}

// Down rolls back exactly one migration version and returns the new current version.
func (r *Runner) Down(ctx context.Context) (int, error) {
	if err := r.ensureRecord(ctx); err != nil {
		return 0, err
	}
	var current int
	err := r.withLock(ctx, func() error {
		var err error
		current, err = r.downLocked(ctx)
		return err
	})
	return current, err
}

func (r *Runner) upLocked(ctx context.Context) (int, error) {
	ms, applied, err := r.loadState(ctx)
	if err != nil {
		return 0, err
	}
	current := 0
	if len(applied) > 0 {
		current = applied[len(applied)-1]
	}
	for _, m := range ms {
		if m.version <= current {
			continue
		}
		if m.version != current+1 {
			return 0, ErrMigrateFailed
		}
		if err := r.applyOne(ctx, m.version, m.up, true); err != nil {
			return 0, err
		}
		current = m.version
	}
	return current, nil
}

func (r *Runner) downLocked(ctx context.Context) (int, error) {
	ms, applied, err := r.loadState(ctx)
	if err != nil {
		return 0, err
	}
	if len(applied) == 0 {
		return 0, nil
	}
	current := applied[len(applied)-1]
	var m *migrationFiles
	for i := range ms {
		if ms[i].version == current {
			m = &ms[i]
			break
		}
	}
	if m == nil {
		return 0, ErrMigrateFailed
	}
	if err := r.applyOne(ctx, current, m.down, false); err != nil {
		return 0, err
	}
	return current - 1, nil
}

func (r *Runner) loadState(ctx context.Context) ([]migrationFiles, []int, error) {
	ms, err := loadMigrations(r.fsys)
	if err != nil {
		return nil, nil, err
	}
	applied, err := r.appliedVersions(ctx)
	if err != nil {
		return nil, nil, err
	}
	if err := validateAppliedAgainst(ms, applied); err != nil {
		return nil, nil, err
	}
	return ms, applied, nil
}

func (r *Runner) applyOne(ctx context.Context, version int, sqlPath string, up bool) error {
	body, err := fs.ReadFile(r.fsys, sqlPath)
	if err != nil {
		return fmt.Errorf("%w: read %s: %v", ErrMigrateFailed, sqlPath, err)
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrMigrateFailed, err)
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, string(body)); err != nil {
		return fmt.Errorf("%w: %v", ErrMigrateFailed, err)
	}
	if up {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO storage_schema_migrations (version) VALUES ($1)`, version,
		); err != nil {
			return fmt.Errorf("%w: %v", ErrMigrateFailed, err)
		}
	} else {
		if _, err := tx.ExecContext(ctx,
			`DELETE FROM storage_schema_migrations WHERE version = $1`, version,
		); err != nil {
			return fmt.Errorf("%w: %v", ErrMigrateFailed, err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("%w: %v", ErrMigrateFailed, err)
	}
	return nil
}

func (r *Runner) ensureRecord(ctx context.Context) error {
	if _, err := r.db.ExecContext(ctx, ensureRecordSQL); err != nil {
		return fmt.Errorf("%w: %v", ErrMigrateFailed, err)
	}
	return nil
}

func (r *Runner) withLock(ctx context.Context, fn func() error) error {
	if _, err := r.db.ExecContext(ctx, `SELECT pg_advisory_lock($1)`, advisoryLockKey); err != nil {
		return fmt.Errorf("%w: %v", ErrMigrateFailed, err)
	}
	defer func() {
		_, _ = r.db.ExecContext(context.Background(), `SELECT pg_advisory_unlock($1)`, advisoryLockKey)
	}()
	return fn()
}

func (r *Runner) appliedVersions(ctx context.Context) ([]int, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT version FROM storage_schema_migrations ORDER BY version`)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrMigrateFailed, err)
	}
	defer func() { _ = rows.Close() }()

	var out []int
	for rows.Next() {
		var v int
		if err := rows.Scan(&v); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrMigrateFailed, err)
		}
		out = append(out, v)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrMigrateFailed, err)
	}
	return out, nil
}

func validateAppliedAgainst(ms []migrationFiles, applied []int) error {
	head := 0
	if len(ms) > 0 {
		head = ms[len(ms)-1].version
	}
	known := make(map[int]struct{}, len(ms))
	for _, m := range ms {
		known[m.version] = struct{}{}
	}

	if len(applied) == 0 {
		return nil
	}
	sorted := append([]int(nil), applied...)
	sort.Ints(sorted)
	for i, v := range sorted {
		if v < 1 {
			return ErrMigrateFailed
		}
		if _, ok := known[v]; !ok {
			return ErrMigrateFailed
		}
		if v > head {
			return ErrMigrateFailed
		}
		if v != i+1 {
			return ErrMigrateFailed
		}
	}
	return nil
}
