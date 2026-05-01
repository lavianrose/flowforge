package main

import (
	"log"

	"github.com/lavianrose/flowforge/internal/config"
	"github.com/lavianrose/flowforge/internal/db"
	"github.com/lavianrose/flowforge/internal/server"
)

func main() {
	cfg := config.Load()

	// Initialize database
	if err := db.Init(cfg); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Setup and start server
	srv := server.New(cfg)
	srv.Setup()

	log.Printf("FlowForge backend starting on port %s", cfg.Port)
	if err := srv.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
