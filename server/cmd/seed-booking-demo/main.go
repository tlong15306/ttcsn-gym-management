// Explicit local business fixtures. Synthetic approval is not a real PT review.
package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/config"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/db"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/facility"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/identity"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/repository"
	service "github.com/tlong15306/ttcsn-gym-management/server/internal/service/v1"
	"github.com/tlong15306/ttcsn-gym-management/server/pkg/auth"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

func main() {
	if err := run(); err != nil {
		log.Fatal("Booking demo fixture failed; existing fixture may need manual review: ", err)
	}
}
func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("invalid config")
	}
	u, err := url.Parse(cfg.DatabaseURL)
	if err != nil || u.Hostname() != "127.0.0.1" || (u.Path != "/ttcsn_gym_dev" && u.Path != "/ttcsn_gym_test") {
		return fmt.Errorf("requires project local dev/test DB")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	pool, err := db.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("database unavailable")
	}
	defer pool.Close()
	actor := identity.User{Role: identity.RoleAdmin}
	if err = pool.QueryRow(ctx, `SELECT u.id FROM users u JOIN roles r ON r.id=u.role_id WHERE username='admin.demo' AND u.status='ACTIVE' AND r.code='ADMIN'`).Scan(&actor.ID); err != nil {
		return fmt.Errorf("run seed first; active admin.demo required")
	}
	hash, err := auth.NewPasswordService(10).Hash("DemoPass123!")
	if err != nil {
		return err
	}
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err = tx.Exec(ctx, `SELECT pg_advisory_xact_lock(841018)`); err != nil {
		return err
	}
	for _, item := range []struct{ username, role, name string }{{"member.flow.demo", "MEMBER", "Hội viên DEMO Booking"}, {"pt.flow.demo", "PT", "PT DEMO Booking"}} {
		var uid int64
		err = tx.QueryRow(ctx, `INSERT INTO users(username,password_hash,email,full_name,role_id) SELECT $1::text,$2::text,$1::text||'@demo.local',$3::text,id FROM roles WHERE code=$4 ON CONFLICT DO NOTHING RETURNING users.id`, item.username, hash, item.name, item.role).Scan(&uid)
		if err != nil {
			if !errors.Is(err, pgx.ErrNoRows) {
				return fmt.Errorf("fixture user insert failed: %s", err)
			}
			// Never attach profiles to a colliding existing account or reset its password.
			var valid bool
			if e := tx.QueryRow(ctx, `SELECT u.id,(u.status='ACTIVE' AND r.code=$2 AND u.email=$1::text||'@demo.local') FROM users u JOIN roles r ON r.id=u.role_id WHERE username=$1`, item.username, item.role).Scan(&uid, &valid); e != nil || !valid {
				return fmt.Errorf("demo account collision or inactive account: %s", item.username)
			}
			var profileOK bool
			query := `SELECT EXISTS(SELECT 1 FROM members WHERE user_id=$1 AND member_code='DEMO-FLOW-M' AND status='ACTIVE' AND membership_expiry>=(now() AT TIME ZONE 'Asia/Ho_Chi_Minh')::date+14)`
			if item.role == "PT" {
				query = `SELECT EXISTS(SELECT 1 FROM personal_trainers WHERE user_id=$1 AND approval_status='APPROVED' AND decision_reason='Synthetic local demo; not actual credential approval')`
			}
			if e := tx.QueryRow(ctx, query, uid).Scan(&profileOK); e != nil || !profileOK {
				return fmt.Errorf("existing %s fixture not ready; no automatic overwrite", item.username)
			}
			continue
		}
		if item.role == "MEMBER" {
			_, err = tx.Exec(ctx, `INSERT INTO members(user_id,member_code,join_date,membership_expiry) VALUES($1,'DEMO-FLOW-M',(now() AT TIME ZONE 'Asia/Ho_Chi_Minh')::date,(now() AT TIME ZONE 'Asia/Ho_Chi_Minh')::date+365)`, uid)
		} else {
			var pt int64
			err = tx.QueryRow(ctx, `INSERT INTO personal_trainers(user_id,specialty,experience_text,approval_status,decided_at,decided_by_user_id,decision_reason) VALUES($1,'Strength demo','Synthetic fixture only','APPROVED',clock_timestamp(),$2,'Synthetic local demo; not actual credential approval') RETURNING id`, uid, actor.ID).Scan(&pt)
			if err == nil {
				_, err = tx.Exec(ctx, `INSERT INTO pt_working_hours(pt_id,day_of_week,start_time,end_time) SELECT $1,d,'07:00'::time,'21:00'::time FROM generate_series(1,7) d`, pt)
			}
		}
		if err != nil {
			return fmt.Errorf("create synthetic profile failed")
		}
		_, err = tx.Exec(ctx, `INSERT INTO audit_events(actor_user_id,action,entity_type,entity_id,reason,request_id) VALUES($1,'demo.FIXTURE_CREATED','user',$2,'Synthetic local business fixture; not real approval','booking-demo-seed')`, actor.ID, fmt.Sprint(uid))
		if err != nil {
			return fmt.Errorf("fixture audit failed")
		}
	}
	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("fixture transaction failed")
	}
	facilities := service.NewFacilityService(repository.NewFacilityRepository(pool))
	calendar := service.NewCalendarService(repository.NewCalendarRepository(pool))
	covers := service.NewRoomCoverService(repository.NewRoomCoverRepository(pool))
	for _, item := range []struct{ name, kind, file string }{{"FLOW DEMO · Strength", "GYM", "strength-demo.png"}, {"FLOW DEMO · Cardio", "CARDIO", "cardio-demo.png"}, {"FLOW DEMO · PT", "PT", "pt-demo.png"}} {
		var id int64
		err = pool.QueryRow(ctx, `SELECT id FROM rooms WHERE name=$1`, item.name).Scan(&id)
		if err != nil {
			room, e := facilities.SaveRoom(ctx, actor, 0, facility.RoomInput{Name: item.name, RoomType: item.kind, DefaultCapacity: 12, Status: "AVAILABLE", Description: "Ảnh và dữ liệu minh họa DEMO, không phải phòng thực tế."}, "booking-demo-seed")
			if e != nil {
				return fmt.Errorf("create fixture room failed")
			}
			id = room.ID
			f, e := os.Open(filepath.Join("..", "docs", "reports", "assets", "booking-portals", "photos", item.file))
			if e != nil {
				return fmt.Errorf("run from server; demo image missing")
			}
			_, e = covers.Upload(ctx, actor, id, f, "Ảnh minh họa phòng DEMO", true, "booking-demo-seed")
			f.Close()
			if e != nil {
				return fmt.Errorf("fixture cover failed")
			}
			snap, e := calendar.Snapshot(ctx, actor, id)
			if e != nil {
				return e
			}
			mutation := facility.CalendarMutation{Kind: "WEEKLY", BaseRevision: snap.Room.CalendarRevision, EffectiveFrom: facility.Today(time.Now()), Days: 28, Reason: "Synthetic local demo opening hours"}
			for d := 1; d <= 7; d++ {
				mutation.Hours = append(mutation.Hours, facility.Hours{Weekday: d, Start: 420, End: 1260})
			}
			preview, e := calendar.Preview(ctx, actor, id, mutation)
			if e != nil {
				return e
			}
			_, e = calendar.Confirm(ctx, actor, id, facility.CalendarConfirm{CalendarMutation: mutation, PreviewID: preview.ID}, fmt.Sprintf("flow-demo-room-%d", id), "booking-demo-seed")
			if e != nil {
				return e
			}
		}
		var ready bool
		err = pool.QueryRow(ctx, `SELECT r.status='AVAILABLE' AND r.cover_image IS NOT NULL AND EXISTS(SELECT 1 FROM room_sessions s WHERE s.room_id=r.id AND s.status='OPEN' AND s.starts_at>now()) FROM rooms r WHERE r.id=$1`, id).Scan(&ready)
		if err != nil || !ready {
			return fmt.Errorf("existing fixture room %s not ready; publish/repair manually", item.name)
		}
	}
	var ready bool
	err = pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM personal_trainers p JOIN users u ON u.id=p.user_id WHERE u.username='pt.flow.demo' AND (SELECT count(DISTINCT day_of_week) FROM pt_working_hours h WHERE h.pt_id=p.id AND h.is_active AND h.start_time<='07:00' AND h.end_time>='21:00')=7)`).Scan(&ready)
	if err != nil || !ready {
		return fmt.Errorf("PT fixture working hours changed; no overwrite")
	}
	log.Print("Business demo ready: member.flow.demo / pt.flow.demo, initial password DemoPass123! (existing passwords preserved). Three illustrative rooms have published sessions. Actual booking/check-in remains explicit.")
	return nil
}
