package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	command := "up"
	if len(os.Args) > 1 {
		command = os.Args[1]
	}
	if err := validateCommand(command, cfg.DatabaseURL); err != nil {
		log.Fatal(err)
	}
	db, err := sql.Open("pgx", cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatal(err)
	}
	if err := goose.Run(command, db, "internal/db/migrations"); err != nil {
		log.Fatal(err)
	}
}

func validateCommand(command, databaseURL string) error {
	switch command {
	case "up", "status", "version":
		return nil
	case "down", "reset":
		parsed, err := url.Parse(databaseURL)
		if err != nil || strings.TrimPrefix(parsed.Path, "/") != "ttcsn_gym_test" {
			return fmt.Errorf("migration rollback/reset only accepts ttcsn_gym_test")
		}
		return nil
	default:
		return fmt.Errorf("unsupported migration command; use up, status, version, down or reset")
	}
}
