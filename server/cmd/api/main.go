package main

import (
	"log"

	"github.com/tlong15306/ttcsn-gym-management/server/internal/delivery/http/router"
)

func main() {
	r := router.New()

	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
