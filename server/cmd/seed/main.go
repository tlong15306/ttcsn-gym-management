package main

import (
	"context"
	"log"

	"github.com/tlong15306/ttcsn-gym-management/server/internal/config"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/db"
	"github.com/tlong15306/ttcsn-gym-management/server/pkg/auth"
)

type seedUser struct {
	username string
	fullName string
	email    string
	role     string
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	pool, err := db.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	hasher := auth.NewPasswordService(cfg.BcryptCost)
	hash, err := hasher.Hash("DemoPass123!")
	if err != nil {
		log.Fatal(err)
	}
	users := []seedUser{
		{"admin.demo", "Quản trị viên Demo", "admin.demo@gym.local", "ADMIN"},
		{"manager.demo", "Quản lý Demo", "manager.demo@gym.local", "MANAGER"},
		{"reception.demo", "Lễ tân Demo", "reception.demo@gym.local", "RECEPTIONIST"},
		{"pt.demo", "Huấn luyện viên Demo", "pt.demo@gym.local", "PT"},
		{"member.demo", "Hội viên Demo", "member.demo@gym.local", "MEMBER"},
	}
	for _, user := range users {
		_, err := pool.Exec(ctx, `INSERT INTO users (username, password_hash, email, full_name, role_id) SELECT $1, $2, $3, $4, id FROM roles WHERE code = $5 ON CONFLICT (LOWER(username)) DO NOTHING`, user.username, hash, user.email, user.fullName, user.role)
		if err != nil {
			log.Fatalf("seed %s: %v", user.username, err)
		}
	}
	log.Print("Demo accounts seeded. Password for newly created accounts: DemoPass123!")
}
