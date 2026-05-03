package seed

import (
	"context"
	"fmt"

	"github.com/lavianrose/flowforge/internal/auth"
	"github.com/lavianrose/flowforge/internal/db"
)

type seedUser struct {
	email    string
	password string
	role     string
}

var defaultUsers = []seedUser{
	{email: "admin@flowforge.local", password: "admin123", role: "admin"},
	{email: "editor@flowforge.local", password: "editor123", role: "editor"},
	{email: "viewer@flowforge.local", password: "viewer123", role: "viewer"},
}

// Run seeds the database with default tenant and users.
// It is idempotent — safe to call multiple times.
func Run(ctx context.Context) error {
	// Create default tenant if not exists
	var tenantID string
	err := db.Pool.QueryRow(ctx, `SELECT id FROM tenants WHERE name = $1`, "Default Tenant").Scan(&tenantID)
	if err != nil {
		err = db.Pool.QueryRow(ctx, `
			INSERT INTO tenants (name)
			VALUES ($1)
			RETURNING id
		`, "Default Tenant").Scan(&tenantID)
		if err != nil {
			return fmt.Errorf("failed to create default tenant: %w", err)
		}
		fmt.Println("Seed: created default tenant")
	}

	// Create default users
	for _, u := range defaultUsers {
		var userID string
		err := db.Pool.QueryRow(ctx, `SELECT id FROM users WHERE email = $1`, u.email).Scan(&userID)
		if err != nil {
			hash, err := auth.HashPassword(u.password)
			if err != nil {
				return fmt.Errorf("failed to hash password for %s: %w", u.email, err)
			}
			err = db.Pool.QueryRow(ctx, `
				INSERT INTO users (tenant_id, email, password_hash, role)
				VALUES ($1, $2, $3, $4)
				RETURNING id
			`, tenantID, u.email, hash, u.role).Scan(&userID)
			if err != nil {
				return fmt.Errorf("failed to create user %s: %w", u.email, err)
			}
			fmt.Printf("Seed: created %s user (%s)\n", u.role, u.email)
		}
	}

	return nil
}
