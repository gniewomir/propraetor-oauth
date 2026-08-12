package migrate

import (
	"fmt"
	"io/fs"
	"regexp"
	"sort"
	"strconv"
)

var migrationName = regexp.MustCompile(`^(\d{4})_([^.]+)\.(up|down)\.sql$`)

// ErrMigrationsInvalid is the locked Operator-facing cause for bad embedded pairs (ADR-0074).
var ErrMigrationsInvalid = fmt.Errorf("storage: migrations invalid")

type migrationFiles struct {
	version int
	up      string // path within fsys
	down    string
}

// Head returns the highest contiguous embedded migration version, or 0 if none.
func Head(fsys fs.FS) (int, error) {
	ms, err := loadMigrations(fsys)
	if err != nil {
		return 0, err
	}
	if len(ms) == 0 {
		return 0, nil
	}
	return ms[len(ms)-1].version, nil
}

func loadMigrations(fsys fs.FS) ([]migrationFiles, error) {
	entries, err := fs.ReadDir(fsys, ".")
	if err != nil {
		return nil, fmt.Errorf("%w", ErrMigrationsInvalid)
	}

	type partial struct {
		slug string
		up   string
		down string
	}
	byVer := map[int]*partial{}

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		m := migrationName.FindStringSubmatch(name)
		if m == nil {
			continue
		}
		ver, convErr := strconv.Atoi(m[1])
		if convErr != nil || ver < 1 {
			return nil, ErrMigrationsInvalid
		}
		slug, dir := m[2], m[3]
		p := byVer[ver]
		if p == nil {
			p = &partial{}
			byVer[ver] = p
		}
		if p.slug != "" && p.slug != slug {
			return nil, ErrMigrationsInvalid
		}
		p.slug = slug
		switch dir {
		case "up":
			if p.up != "" {
				return nil, ErrMigrationsInvalid
			}
			p.up = name
		case "down":
			if p.down != "" {
				return nil, ErrMigrationsInvalid
			}
			p.down = name
		}
	}

	if len(byVer) == 0 {
		return nil, nil
	}

	versions := make([]int, 0, len(byVer))
	for v := range byVer {
		versions = append(versions, v)
	}
	sort.Ints(versions)

	out := make([]migrationFiles, 0, len(versions))
	for i, v := range versions {
		if v != i+1 {
			return nil, ErrMigrationsInvalid
		}
		p := byVer[v]
		if p.up == "" || p.down == "" {
			return nil, ErrMigrationsInvalid
		}
		out = append(out, migrationFiles{
			version: v,
			up:      p.up,
			down:    p.down,
		})
	}
	return out, nil
}
