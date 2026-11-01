package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	t.Setenv("APP_ENV", "test")
	t.Setenv("HTTP_ADDR", ":9090")
	t.Setenv("DATABASE_URL", "postgres://example")
	t.Setenv("JWT_SECRET", "01234567890123456789012345678901")
	t.Setenv("JWT_TTL", "30m")
	t.Setenv("BCRYPT_COST", "10")

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
	if cfg.JWTTTL.String() != "30m0s" {
		t.Fatalf("JWTTTL = %s, want 30m", cfg.JWTTTL)
	}
}

func TestLoadRejectsMissingSecret(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://example")
	t.Setenv("JWT_SECRET", "")

	if _, err := Load(); err == nil {
		t.Fatal("Load() error = nil, want missing JWT_SECRET error")
	}
}

func TestLoadLocalEnvironmentPrefersLocalFileButNotShell(t *testing.T) {
	dir := t.TempDir()
	previousDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(previousDir) })

	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte("CONFIG_FILE_ORDER=base\nCONFIG_SHELL_PRIORITY=file\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".env.local"), []byte("CONFIG_FILE_ORDER=local\nCONFIG_SHELL_PRIORITY=local\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CONFIG_SHELL_PRIORITY", "shell")

	if err := loadLocalEnvironment(); err != nil {
		t.Fatal(err)
	}
	if got := os.Getenv("CONFIG_FILE_ORDER"); got != "local" {
		t.Fatalf("CONFIG_FILE_ORDER = %q, want local", got)
	}
	if got := os.Getenv("CONFIG_SHELL_PRIORITY"); got != "shell" {
		t.Fatalf("CONFIG_SHELL_PRIORITY = %q, want shell", got)
	}
}
