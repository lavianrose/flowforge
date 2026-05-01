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

func Up() error {
	ctx := context.Background()
	files, err := filepath.Glob("migrations/*.up.sql")
	if err != nil {
		return fmt.Errorf("failed to read migrations: %w", err)
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

func Down() error {
	ctx := context.Background()
	files, err := filepath.Glob("migrations/*.down.sql")
	if err != nil {
		return fmt.Errorf("failed to read migrations: %w", err)
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
