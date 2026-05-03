package main

import (
	"context"
	"log"

	"github.com/lavianrose/flowforge/internal/config"
	"github.com/lavianrose/flowforge/internal/db"
	"github.com/lavianrose/flowforge/internal/migrate"
	"github.com/lavianrose/flowforge/internal/seed"
	"github.com/lavianrose/flowforge/internal/server"
)

func main() {
	cfg := config.Load()

	// Initialize database
	if err := db.Init(cfg); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Run migrations
	log.Println("Running database migrations...")
	if err := migrate.Up("migrations"); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Seed database
	log.Println("Seeding database...")
	if err := seed.Run(context.Background()); err != nil {
		log.Fatalf("Failed to seed database: %v", err)
	}

	// Setup and start server
	srv := server.New(cfg)
	srv.Setup()

	log.Printf("FlowForge backend starting on port %s", cfg.Port)
	if err := srv.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
