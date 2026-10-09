package main

import (
	"log"

	"github.com/tlong15306/ttcsn-gym-management/server/internal/config"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/delivery/http/router"
)

func main() {
	cfg := config.Load()
	r, err := router.New()
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("starting API in %s mode", cfg.AppEnv)

	if err := r.Run(cfg.HTTPAddr); err != nil {
		log.Fatal(err)
	}
}
