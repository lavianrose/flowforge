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

	// Try to get existing default tenant first
	var tenantID string
	err := db.Pool.QueryRow(ctx, `SELECT id FROM tenants WHERE name = $1`, "Default Tenant").Scan(&tenantID)

	if err != nil {
		// Tenant doesn't exist, create it
		err = db.Pool.QueryRow(ctx, `
			INSERT INTO tenants (name)
			VALUES ($1)
			RETURNING id
		`, "Default Tenant").Scan(&tenantID)

		if err != nil {
			log.Fatalf("Failed to create tenant: %v", err)
		}
		fmt.Println("Created new tenant")
	} else {
		fmt.Println("Using existing tenant")
	}

	fmt.Printf("Tenant ID: %s\n", tenantID)

	// Try to get existing admin user first
	email := "admin@flowforge.local"
	password := "admin123"
	var userID string

	err = db.Pool.QueryRow(ctx, `SELECT id FROM users WHERE email = $1`, email).Scan(&userID)

	if err != nil {
		// User doesn't exist, create it
		hash, err := auth.HashPassword(password)
		if err != nil {
			log.Fatalf("Failed to hash password: %v", err)
		}

		err = db.Pool.QueryRow(ctx, `
			INSERT INTO users (tenant_id, email, password_hash, role)
			VALUES ($1, $2, $3, $4)
			RETURNING id
		`, tenantID, email, hash, "admin").Scan(&userID)

		if err != nil {
			log.Fatalf("Failed to create user: %v", err)
		}
		fmt.Println("Created new admin user")
	} else {
		fmt.Println("Using existing admin user")
	}

	fmt.Printf("User ID: %s\n", userID)
	fmt.Printf("\n=== Seed Complete ===\n")
	fmt.Printf("Email: %s\n", email)
	fmt.Printf("Password: %s\n", password)
	fmt.Printf("\nLogin with POST /api/v1/auth/login\n")
}
