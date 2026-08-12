package postgres

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// Ping opens storage at storageURL and runs SELECT 1 (ADR-0072 verify).
func Ping(ctx context.Context, storageURL string) error {
	if storageURL == "" {
		return fmt.Errorf("empty storage URL")
	}
	db, err := sql.Open("pgx", storageURL)
	if err != nil {
		return err
	}
	defer db.Close()

	var one int
	if err := db.QueryRowContext(ctx, "SELECT 1").Scan(&one); err != nil {
		return err
	}
	if one != 1 {
		return fmt.Errorf("unexpected SELECT 1 result: %d", one)
	}
	return nil
}
