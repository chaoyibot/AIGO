package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"

	"github.com/aigo/internal/config"
	"github.com/aigo/internal/repository"
)

func main() {
	cfg := config.Load()

	db, err := repository.Connect(cfg.Database.URL)
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	defer db.Close()

	action := "up"
	if len(os.Args) > 1 {
		action = os.Args[1]
	}

	switch action {
	case "up":
		runMigrations(db, "migrations")
	case "reset":
		resetDB(db)
		runMigrations(db, "migrations")
	default:
		fmt.Printf("Usage: migrate [up|reset]\n")
	}
}

func runMigrations(db *sql.DB, dir string) {
	files, err := filepath.Glob(filepath.Join(dir, "*.sql"))
	if err != nil {
		log.Fatalf("list migrations: %v", err)
	}
	sort.Strings(files)

	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			log.Fatalf("read %s: %v", f, err)
		}
		if _, err := db.Exec(string(data)); err != nil {
			log.Fatalf("exec %s: %v", f, err)
		}
		fmt.Printf("Applied: %s\n", filepath.Base(f))
	}
}

func resetDB(db *sql.DB) {
	tables := []string{
		"broadcast_reads", "broadcasts", "webhook_logs", "webhooks",
		"withdrawal_orders", "recharge_orders", "transactions",
		"escrows", "orders", "listings", "products",
		"api_keys", "wallets", "users", "events",
	}
	for _, t := range tables {
		db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE", t))
	}
	fmt.Println("Database reset completed")
}
