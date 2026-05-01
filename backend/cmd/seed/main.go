package main

import (
	"context"
	"fmt"
	"log"

	"github.com/lavianrose/flowforge/internal/auth"
	"github.com/lavianrose/flowforge/internal/config"
	"github.com/lavianrose/flowforge/internal/db"
)

func main() {
	cfg := config.Load()

	if err := db.Init(cfg); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	// Create default tenant
	var tenantID string
	err := db.Pool.QueryRow(ctx, `
		INSERT INTO tenants (name)
		VALUES ($1)
		ON CONFLICT (name) DO NOTHING
		RETURNING id
	`, "Default Tenant").Scan(&tenantID)

	if err != nil {
		log.Printf("Warning: Could not create tenant: %v", err)
		// Try to get existing tenant
		err = db.Pool.QueryRow(ctx, `SELECT id FROM tenants WHERE name = $1`, "Default Tenant").Scan(&tenantID)
		if err != nil {
			log.Fatalf("Failed to get tenant: %v", err)
		}
	}

	fmt.Printf("Tenant ID: %s\n", tenantID)

	// Create admin user
	email := "admin@flowforge.local"
	password := "admin123"

	hash, err := auth.HashPassword(password)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	var userID string
	err = db.Pool.QueryRow(ctx, `
		INSERT INTO users (tenant_id, email, password_hash, role)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (tenant_id, email) DO NOTHING
		RETURNING id
	`, tenantID, email, hash, "admin").Scan(&userID)

	if err != nil {
		log.Printf("Warning: Could not create user (may already exist): %v", err)
		// Try to get existing user
		err = db.Pool.QueryRow(ctx, `SELECT id FROM users WHERE email = $1`, email).Scan(&userID)
		if err != nil {
			log.Fatalf("Failed to get user: %v", err)
		}
	}

	fmt.Printf("User ID: %s\n", userID)
	fmt.Printf("\n=== Seed Complete ===\n")
	fmt.Printf("Email: %s\n", email)
	fmt.Printf("Password: %s\n", password)
	fmt.Printf("\nLogin with POST /api/v1/auth/login\n")
}
