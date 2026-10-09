package main

import (
	"context"
	"log"
	"time"

	"github.com/tlong15306/ttcsn-gym-management/server/internal/config"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/delivery/http/router"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/infrastructure/postgres"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	startupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := postgres.Open(startupCtx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	r, err := router.New(pool)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("PostgreSQL connected; starting API in %s mode", cfg.AppEnv)

	if err := r.Run(cfg.HTTPAddr); err != nil {
		log.Fatal(err)
	}
}
