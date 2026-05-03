package migrate

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/jackc/pgx/v5"
	"github.com/lavianrose/flowforge/internal/db"
)

// Up runs all *.up.sql migration files from the given directory.
// dir should be the path to the folder containing migration files
// (e.g. "migrations" when CWD is the project root).
func Up(dir string) error {
	ctx := context.Background()
	pattern := filepath.Join(dir, "*.up.sql")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("failed to read migrations: %w", err)
	}

	if len(files) == 0 {
		return fmt.Errorf("no migration files found matching %s", pattern)
	}

	sort.Strings(files)

	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", file, err)
		}

		if _, err := db.Pool.Exec(ctx, string(content)); err != nil {
			return fmt.Errorf("failed to run migration %s: %w", file, err)
		}

		fmt.Printf("Applied migration: %s\n", filepath.Base(file))
	}

	return nil
}

// Down runs all *.down.sql migration files from the given directory in reverse order.
func Down(dir string) error {
	ctx := context.Background()
	pattern := filepath.Join(dir, "*.down.sql")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("failed to read migrations: %w", err)
	}

	if len(files) == 0 {
		return fmt.Errorf("no migration files found matching %s", pattern)
	}

	sort.Sort(sort.Reverse(sort.StringSlice(files)))

	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", file, err)
		}

		if _, err := db.Pool.Exec(ctx, string(content)); err != nil {
			return fmt.Errorf("failed to run migration %s: %w", file, err)
		}

		fmt.Printf("Reversed migration: %s\n", filepath.Base(file))
	}

	return nil
}

func CreateMigrationTable(ctx context.Context, conn *pgx.Conn) error {
	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMP NOT NULL DEFAULT NOW()
		);
	`
	_, err := conn.Exec(ctx, query)
	return err
}
