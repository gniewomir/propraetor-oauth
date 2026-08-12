// Package migrate advances the Storage schema using paired SQL files and
// storage_schema_migrations (ADR-0074). The CLI composition root wires a
// database and an fs.FS of migration files (typically assets.Migrations).
package migrate
