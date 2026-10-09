package config

import "testing"

func TestLoad(t *testing.T) {
	t.Setenv("APP_ENV", "test")
	t.Setenv("HTTP_ADDR", ":9090")
	t.Setenv("DATABASE_URL", "postgres://example")

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}

	if cfg.AppEnv != "test" {
		t.Fatalf("AppEnv = %q, want %q", cfg.AppEnv, "test")
	}

	if cfg.HTTPAddr != ":9090" {
		t.Fatalf("HTTPAddr = %q, want %q", cfg.HTTPAddr, ":9090")
	}

	if cfg.DatabaseURL != "postgres://example" {
		t.Fatalf("DatabaseURL = %q, want %q", cfg.DatabaseURL, "postgres://example")
	}
}
