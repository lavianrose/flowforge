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

	// Create editor user
	editorEmail := "editor@flowforge.local"
	editorPassword := "editor123"
	var editorID string

	err = db.Pool.QueryRow(ctx, `SELECT id FROM users WHERE email = $1`, editorEmail).Scan(&editorID)

	if err != nil {
		hash, err := auth.HashPassword(editorPassword)
		if err != nil {
			log.Fatalf("Failed to hash password: %v", err)
		}

		err = db.Pool.QueryRow(ctx, `
			INSERT INTO users (tenant_id, email, password_hash, role)
			VALUES ($1, $2, $3, $4)
			RETURNING id
		`, tenantID, editorEmail, hash, "editor").Scan(&editorID)

		if err != nil {
			log.Fatalf("Failed to create editor user: %v", err)
		}
		fmt.Println("Created new editor user")
	} else {
		fmt.Println("Using existing editor user")
	}

	// Create viewer user
	viewerEmail := "viewer@flowforge.local"
	viewerPassword := "viewer123"
	var viewerID string

	err = db.Pool.QueryRow(ctx, `SELECT id FROM users WHERE email = $1`, viewerEmail).Scan(&viewerID)

	if err != nil {
		hash, err := auth.HashPassword(viewerPassword)
		if err != nil {
			log.Fatalf("Failed to hash password: %v", err)
		}

		err = db.Pool.QueryRow(ctx, `
			INSERT INTO users (tenant_id, email, password_hash, role)
			VALUES ($1, $2, $3, $4)
			RETURNING id
		`, tenantID, viewerEmail, hash, "viewer").Scan(&viewerID)

		if err != nil {
			log.Fatalf("Failed to create viewer user: %v", err)
		}
		fmt.Println("Created new viewer user")
	} else {
		fmt.Println("Using existing viewer user")
	}

	fmt.Printf("\n=== Seed Complete ===\n")
	fmt.Printf("Admin - Email: %s, Password: %s\n", email, password)
	fmt.Printf("Editor - Email: %s, Password: %s\n", editorEmail, editorPassword)
	fmt.Printf("Viewer - Email: %s, Password: %s\n", viewerEmail, viewerPassword)
	fmt.Printf("\nLogin with POST /api/v1/auth/login\n")
}
