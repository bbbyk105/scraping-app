package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	direction := flag.String("direction", "up", "migration direction: up or down")
	flag.Parse()

	// Construct database URL from env vars
	host := getEnv("POSTGRES_HOST", "localhost")
	port := getEnv("POSTGRES_PORT", "5432")
	user := getEnv("POSTGRES_USER", "pricecompare")
	password := getEnv("POSTGRES_PASSWORD", "password")
	dbname := getEnv("POSTGRES_DB", "pricecompare")
	sslmode := getEnv("POSTGRES_SSLMODE", "disable")
	databaseURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		user, password, host, port, dbname, sslmode)

	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}
	defer db.Close()

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Fatal("Failed to create driver:", err)
	}

	// Get migrations path
	// In Docker, migrations are in /app/migrations
	// Locally, they're relative to the working directory
	migrationsPath := "file://migrations"
	if _, err := os.Stat("/app/migrations"); err == nil {
		migrationsPath = "file:///app/migrations"
	}
	
	m, err := migrate.NewWithDatabaseInstance(
		migrationsPath,
		"postgres",
		driver,
	)
	if err != nil {
		log.Fatal("Failed to create migrate instance:", err)
	}

	if *direction == "up" {
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			log.Fatal("Migration up failed:", err)
		}
		log.Println("Migration up completed")
	} else {
		if err := m.Down(); err != nil && err != migrate.ErrNoChange {
			log.Fatal("Migration down failed:", err)
		}
		log.Println("Migration down completed")
	}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

