// Package integration_test holds in-process Postgres tests (ADR-0073).
//
// Suite runner (scripts/test-integration.sh): ensure → truncate → migrate (when available).
// Per test: parallel by default; wrap work in a transaction and roll back so the DB
// stays empty (Ping opens its own connection, so this first case only needs a live URL).
// Prefer real internal APIs for fixtures; direct SQL only with a strong reason.
package integration_test
