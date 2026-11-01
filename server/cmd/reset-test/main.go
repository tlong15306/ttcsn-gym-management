package main

import (
	"database/sql"
	"log"
	"net/url"
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
	databaseURL, err := url.Parse(cfg.DatabaseURL)
	if err != nil || strings.TrimPrefix(databaseURL.Path, "/") != "ttcsn_gym_test" {
		log.Fatal("reset-test only accepts a DATABASE_URL for ttcsn_gym_test")
	}
	db, err := sql.Open("pgx", cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatal(err)
	}
	if err := goose.DownTo(db, "internal/db/migrations", 0); err != nil {
		log.Fatal(err)
	}
	if err := goose.Up(db, "internal/db/migrations"); err != nil {
		log.Fatal(err)
	}
	log.Print("ttcsn_gym_test reset and migrated")
}
