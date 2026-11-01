// Optional illustrative rooms only; never changes existing passwords, schedules or covers.
package main

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/config"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/db"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/facility"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/identity"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/repository"
	service "github.com/tlong15306/ttcsn-gym-management/server/internal/service/v1"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	parsed, err := url.Parse(cfg.DatabaseURL)
	if err != nil || parsed.Hostname() != "127.0.0.1" || (parsed.Path != "/ttcsn_gym_dev" && parsed.Path != "/ttcsn_gym_test") {
		log.Fatal("room demo seed requires project local dev/test database")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	pool, err := db.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal("database unavailable")
	}
	defer pool.Close()
	var actor identity.User
	actor.Role = identity.RoleAdmin
	if err = pool.QueryRow(ctx, `SELECT u.id FROM users u JOIN roles r ON r.id=u.role_id WHERE u.username='admin.demo' AND u.status='ACTIVE' AND r.code='ADMIN'`).Scan(&actor.ID); err != nil {
		log.Fatal("run seed first; active admin.demo required")
	}
	facilities := service.NewFacilityService(repository.NewFacilityRepository(pool))
	covers := service.NewRoomCoverService(repository.NewRoomCoverRepository(pool))
	for _, item := range []struct{ name, kind, file, description string }{{"DEMO · Strength Studio", "GYM", "strength-demo.png", "Không gian sức mạnh minh họa, không phải ảnh phòng thực tế."}, {"DEMO · Cardio Club", "CARDIO", "cardio-demo.png", "Không gian cardio minh họa, không phải ảnh phòng thực tế."}, {"DEMO · Personal Training", "PT", "pt-demo.png", "Phòng huấn luyện cá nhân minh họa, không phải ảnh phòng thực tế."}} {
		var id int64
		err = pool.QueryRow(ctx, `SELECT id FROM rooms WHERE lower(name)=lower($1)`, item.name).Scan(&id)
		if err == pgx.ErrNoRows {
			r, e := facilities.SaveRoom(ctx, actor, 0, facility.RoomInput{Name: item.name, RoomType: item.kind, DefaultCapacity: 12, Description: item.description, Status: "AVAILABLE"}, "room-demo-seed")
			if e != nil {
				log.Fatal("create demo room failed")
			}
			id = r.ID
		} else if err != nil {
			log.Fatal("read demo room failed")
		}
		r, e := facilities.Room(ctx, id)
		if e != nil {
			log.Fatal("read room failed")
		}
		if r.HasCover {
			continue
		}
		f, e := os.Open(filepath.Join("..", "docs", "reports", "assets", "booking-portals", "photos", item.file))
		if e != nil {
			log.Fatal("demo image missing; run from server directory")
		}
		_, e = covers.Upload(ctx, actor, id, f, item.description, true, "room-demo-seed")
		f.Close()
		if e != nil {
			log.Fatal("upload demo cover failed")
		}
	}
	log.Print("Illustrative rooms ready. Existing covers and calendars preserved; publish opening hours before booking.")
}
