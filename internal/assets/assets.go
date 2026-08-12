// Package assets exposes embedded static files as fs.FS (ADR-0074).
//
// Callers depend on fs.FS, not embed. Production wires the embedded tree;
// tests may pass a different fs.FS. The package is oblivious to file roles
// (migrations, later templates, …).
package assets

import (
	"embed"
	"io/fs"
)

//go:embed all:migrations
var migrationsRoot embed.FS

// Migrations returns the embedded migrations directory as an fs.FS
// (migration SQL files at the FS root).
func Migrations() fs.FS {
	sub, err := fs.Sub(migrationsRoot, "migrations")
	if err != nil {
		panic("assets: migrations subtree: " + err.Error())
	}
	return sub
}
