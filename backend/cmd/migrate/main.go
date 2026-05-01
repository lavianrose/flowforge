package main

import (
	"fmt"
	"log"
	"os"

	"github.com/lavianrose/flowforge/internal/config"
	"github.com/lavianrose/flowforge/internal/db"
	"github.com/lavianrose/flowforge/internal/migrate"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: migrate <up|down>")
	}

	cfg := config.Load()

	if err := db.Init(cfg); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	cmd := os.Args[1]

	switch cmd {
	case "up":
		if err := migrate.Up(); err != nil {
			log.Fatalf("Migration failed: %v", err)
		}
		fmt.Println("Migrations applied successfully")

	case "down":
		if err := migrate.Down(); err != nil {
			log.Fatalf("Migration failed: %v", err)
		}
		fmt.Println("Migrations rolled back successfully")

	default:
		log.Fatal("Unknown command. Use 'up' or 'down'")
	}
}
