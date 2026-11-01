package main

import (
	"log/slog"
	"os"

	"github.com/tlong15306/ttcsn-gym-management/server/internal/app"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("load config", "error", err)
		os.Exit(1)
	}

	application, err := app.NewApplication(cfg)
	if err != nil {
		slog.Error("initialize application", "error", err)
		os.Exit(1)
	}
	if err := application.Run(); err != nil {
		slog.Error("application stopped", "error", err)
		os.Exit(1)
	}
}
