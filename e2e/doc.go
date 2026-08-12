// Package e2e_test drives bin/oauth as a subprocess (ADR-0073).
//
// Suite runner (scripts/test-e2e.sh): ensure → truncate → migrate (when available).
// Per test: parallel by default; unique IDs for fixtures; no harness transactions;
// no mid-suite truncate. Prefer Operator CLI for setup, not direct SQL.
package e2e_test
