package repository_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/identity"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/repository"
	v1service "github.com/tlong15306/ttcsn-gym-management/server/internal/service/v1"
	"github.com/tlong15306/ttcsn-gym-management/server/pkg/auth"
)

func TestIdentityRepositoryIntegration(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("set TEST_DATABASE_URL for ttcsn_gym_test to run PostgreSQL integration tests")
	}
	parsedURL, err := url.Parse(databaseURL)
	if err != nil || strings.TrimPrefix(parsedURL.Path, "/") != "ttcsn_gym_test" {
		t.Fatal("TEST_DATABASE_URL must point exactly to ttcsn_gym_test")
	}

	ctx := context.Background()
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := goose.SetDialect("postgres"); err != nil {
		t.Fatal(err)
	}
	migrations := filepath.Join("..", "db", "migrations")
	if err := goose.DownTo(db, migrations, 0); err != nil && !strings.Contains(err.Error(), "no current version") {
		t.Fatal(err)
	}
	if err := goose.Up(db, migrations); err != nil {
		t.Fatal(err)
	}

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	userRepository := repository.NewIdentityRepository(pool)
	passwords := auth.NewPasswordService(10)
	tokens := auth.NewTokenService("01234567890123456789012345678901", time.Minute)
	service := v1service.NewService(userRepository, passwords, tokens)

	hash, err := passwords.Hash("DemoPass123!")
	if err != nil {
		t.Fatal(err)
	}
	var adminID int64
	if err := pool.QueryRow(ctx, `INSERT INTO users (username, password_hash, email, full_name, role_id) SELECT 'admin.integration', $1, 'admin.integration@gym.local', 'Admin Integration', id FROM roles WHERE code = 'ADMIN' RETURNING id`, hash).Scan(&adminID); err != nil {
		t.Fatal(err)
	}
	admin, err := userRepository.FindByID(ctx, adminID)
	if err != nil {
		t.Fatal(err)
	}

	created, err := service.CreateUser(ctx, admin, identity.CreateInput{Username: "reception.integration", Password: "Password123!", Email: "reception.integration@gym.local", FullName: "Reception Integration", Role: identity.RoleReceptionist}, "integration-request")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.CreateUser(ctx, admin, identity.CreateInput{Username: "reception.integration", Password: "Password123!", Email: "different@gym.local", FullName: "Duplicate", Role: identity.RoleReceptionist}, "integration-request"); !errors.Is(err, identity.ErrConflict) {
		t.Fatalf("duplicate create error = %v, want conflict", err)
	}

	loggedIn, err := service.Login(ctx, created.Username, "Password123!", time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.SetStatus(ctx, admin, created.ID, identity.StatusLocked, "integration lock", "integration-request"); err != nil {
		t.Fatal(err)
	}
	current, err := userRepository.FindByID(ctx, created.ID)
	if err != nil {
		t.Fatal(err)
	}
	claims, err := tokens.Parse(loggedIn.AccessToken)
	if err != nil {
		t.Fatal(err)
	}
	if current.AuthVersion == claims.AuthVersion || current.Status != identity.StatusLocked {
		t.Fatalf("token was not invalidated after lock: user=%+v claims=%+v", current, claims)
	}
	if _, err := service.SetStatus(ctx, admin, admin.ID, identity.StatusLocked, "test last active admin", "integration-request"); !errors.Is(err, identity.ErrLastActiveAdmin) {
		t.Fatalf("lock last admin error = %v, want last active admin", err)
	}
	var auditCount int
	if err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM audit_events WHERE entity_id = $1`, fmt.Sprint(created.ID)).Scan(&auditCount); err != nil {
		t.Fatal(err)
	}
	if auditCount < 2 {
		t.Fatalf("audit count = %d, want at least 2", auditCount)
	}
}
