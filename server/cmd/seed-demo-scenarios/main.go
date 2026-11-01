// Command seed-demo-scenarios creates an idempotent, local-only business demo.
//
// It intentionally rejects every database except local ttcsn_gym_dev. Existing
// accounts, rooms, schedules, covers, and booking fixtures are validated rather
// than overwritten. New accounts use DemoPass123!, while passwords on accounts
// that already exist are never changed.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/config"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/db"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/booking"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/facility"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/identity"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/repository"
	service "github.com/tlong15306/ttcsn-gym-management/server/internal/service/v1"
	"github.com/tlong15306/ttcsn-gym-management/server/pkg/auth"
)

const (
	demoPassword       = "DemoPass123!"
	demoRoomMarker     = "Dữ liệu DEMO cục bộ do cmd/seed-demo-scenarios tạo; không dùng để vận hành thật."
	demoTrainerMarker  = "Synthetic local demo; not actual credential approval"
	demoAuditAction    = "demo.booking_fixture.created"
	demoRequestID      = "seed-demo-scenarios-v1"
	localDemoDatabase  = "ttcsn_gym_dev"
	localVietnamTZName = "Asia/Ho_Chi_Minh"
)

type memberFixture struct {
	Username string
	FullName string
	Email    string
	Code     string
	Status   string
	Locked   bool
	Expired  bool
	ID       int64
	UserID   int64
}

type trainerFixture struct {
	Username   string
	FullName   string
	Email      string
	Specialty  string
	Experience string
	Approval   string
	ID         int64
	UserID     int64
}

type roomFixture struct {
	Name     string
	RoomType string
	Capacity int
	Status   string
	Cover    string
	Alt      string
	ID       int64
}

type bookingFixture struct {
	Code               string
	Status             string
	DayOffset          int
	RoomIndex          int
	MemberIndex        int
	TrainerIndex       int
	MemberAttendance   bool
	TrainerAttendance  bool
	CreateReview       bool
	CancellationSource string
	CancellationRole   string
	CancellationReason string
}

type seedCounts struct {
	MembersCreated   int
	TrainersCreated  int
	RoomsCreated     int
	BookingsCreated  int
	BookingsExisting int
}

func main() {
	refreshPending := flag.Bool("refresh-pending", false, "create fresh demo PT requests only for completed expired/rejected/cancelled demo requests")
	flag.Parse()
	if err := run(*refreshPending); err != nil {
		log.Fatal("Demo scenario seed failed; existing data was not reset: ", err)
	}
}

func run(refreshPending bool) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load local configuration: %w", err)
	}
	if err := validateLocalDevURL(cfg.DatabaseURL); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	pool, err := db.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("open local demo database: %w", err)
	}
	defer pool.Close()

	var currentDatabase string
	if err := pool.QueryRow(ctx, `SELECT current_database()`).Scan(&currentDatabase); err != nil {
		return fmt.Errorf("verify current database: %w", err)
	}
	if currentDatabase != localDemoDatabase {
		return fmt.Errorf("refuse to seed %q; this command only accepts %q", currentDatabase, localDemoDatabase)
	}

	adminID, err := requireActiveRole(ctx, pool, "admin.demo", "ADMIN")
	if err != nil {
		return err
	}
	receptionID, err := requireActiveRole(ctx, pool, "reception.demo", "RECEPTIONIST")
	if err != nil {
		return err
	}
	actor := identity.User{ID: adminID, Role: identity.RoleAdmin}

	hash, err := auth.NewPasswordService(cfg.BcryptCost).Hash(demoPassword)
	if err != nil {
		return fmt.Errorf("hash demo password: %w", err)
	}

	counts := seedCounts{}
	members, trainers, err := ensureProfiles(ctx, pool, adminID, hash, &counts)
	if err != nil {
		return err
	}
	if err := ensureUsableSlots(ctx, pool, actor); err != nil {
		return err
	}
	rooms, err := ensureRooms(ctx, pool, actor, &counts)
	if err != nil {
		return err
	}
	bookingsCreated, bookingsExisting, err := ensureBookingFixtures(ctx, pool, adminID, receptionID, members, trainers, rooms, refreshPending)
	if err != nil {
		return err
	}
	counts.BookingsCreated = bookingsCreated
	counts.BookingsExisting = bookingsExisting
	if err := ensureCapacityScenarios(ctx, pool, members, rooms[4]); err != nil {
		return fmt.Errorf("prepare demo capacity scenarios: %w", err)
	}
	manifestPath, err := writeManifest(ctx, pool, members, trainers, rooms)
	if err != nil {
		return fmt.Errorf("write demo manifest: %w", err)
	}

	log.Printf("DEMO scenarios ready in %s: %d new members, %d new trainers, %d new rooms, %d new bookings (%d already present).", localDemoDatabase, counts.MembersCreated, counts.TrainersCreated, counts.RoomsCreated, counts.BookingsCreated, counts.BookingsExisting)
	log.Printf("Demo scenario manifest: %s", manifestPath)
	log.Printf("New demo passwords are %q; passwords for previously existing accounts were preserved. Log in with demo.member.01, demo.pt.strength, manager.demo, or reception.demo.", demoPassword)
	return nil
}

// Create separately tagged self-training holds for full and nearly-full
// sessions. Existing user actions are preserved on repeat runs.
func ensureCapacityScenarios(ctx context.Context, pool *pgxpool.Pool, members []memberFixture, room roomFixture) error {
	var today time.Time
	if err := pool.QueryRow(ctx, `SELECT (clock_timestamp() AT TIME ZONE $1)::date`, localVietnamTZName).Scan(&today); err != nil {
		return err
	}
	repo := repository.NewBookingRepository(pool)
	for _, scenario := range []struct {
		name          string
		offset, holds int
	}{{"FULL", 1, room.Capacity}, {"ONE-LEFT", 2, room.Capacity - 1}} {
		date := today.AddDate(0, 0, scenario.offset).Format("2006-01-02")
		prefix := fmt.Sprintf("DEMO-CAPACITY-%s@%s#", scenario.name, date)
		var sessionID int64
		err := pool.QueryRow(ctx, `SELECT b.room_session_id FROM audit_events a JOIN bookings b ON b.id::text=a.entity_id WHERE a.action=$1 AND a.reason LIKE $2 ORDER BY a.id LIMIT 1`, demoAuditAction, prefix+"%").Scan(&sessionID)
		if errors.Is(err, pgx.ErrNoRows) {
			err = pool.QueryRow(ctx, `SELECT s.id FROM room_sessions s WHERE s.room_id=$1 AND s.session_date=$2::date AND s.status='OPEN' AND s.starts_at>clock_timestamp()
				AND NOT EXISTS (SELECT 1 FROM bookings b WHERE b.room_session_id=s.id AND b.status IN ('PENDING_PT','CONFIRMED')) ORDER BY s.starts_at DESC LIMIT 1`, room.ID, date).Scan(&sessionID)
		}
		if err != nil {
			return fmt.Errorf("find unused %s session on %s: %w", scenario.name, date, err)
		}
		processed := 0
		for _, member := range members {
			if member.Expired || member.Locked {
				continue
			}
			code := fmt.Sprintf("%s%d", prefix, member.ID)
			var existing bool
			if err := pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM audit_events WHERE action=$1 AND reason=$2)`, demoAuditAction, code).Scan(&existing); err != nil {
				return err
			}
			if existing {
				processed++
				if processed >= scenario.holds {
					break
				}
				continue
			}
			var overlap bool
			if err := pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM bookings b JOIN room_sessions old ON old.id=b.room_session_id JOIN room_sessions target ON target.id=$2 WHERE b.member_id=$1 AND b.status IN ('PENDING_PT','CONFIRMED') AND tstzrange(old.starts_at,old.ends_at,'[)') && tstzrange(target.starts_at,target.ends_at,'[)'))`, member.ID, sessionID).Scan(&overlap); err != nil {
				return err
			}
			if overlap {
				continue
			}
			created, err := repo.Create(ctx, identity.User{ID: member.UserID, Role: identity.RoleMember}, 0, booking.CreateInput{RoomSessionID: sessionID}, "demo:"+code, demoRequestID)
			if err != nil {
				return fmt.Errorf("create %s for %s: %w", scenario.name, member.Username, err)
			}
			if _, err = pool.Exec(ctx, `INSERT INTO audit_events(actor_user_id,action,entity_type,entity_id,reason,request_id) VALUES($1,$2,'booking',$3,$4,$5)`, member.UserID, demoAuditAction, fmt.Sprint(created.ID), code, demoRequestID); err != nil {
				return err
			}
			processed++
			if processed >= scenario.holds {
				break
			}
		}
		if processed < scenario.holds {
			return fmt.Errorf("not enough eligible demo members for %s on %s", scenario.name, date)
		}
		log.Printf("DEMO capacity scenario %s: date=%s session=%d expected holds=%d; existing user actions are preserved", scenario.name, date, sessionID, scenario.holds)
	}
	return nil
}

// writeManifest records the current state, including fixtures changed by users
// or lifecycle jobs. It contains no passwords, tokens or private contact data.
func writeManifest(ctx context.Context, pool *pgxpool.Pool, members []memberFixture, trainers []trainerFixture, rooms []roomFixture) (string, error) {
	type account struct {
		ID       int64  `json:"id"`
		Username string `json:"username"`
	}
	type roomRecord struct {
		ID     int64  `json:"id"`
		Name   string `json:"name"`
		Status string `json:"status"`
	}
	type bookingRecord struct {
		Code      string    `json:"code"`
		ID        int64     `json:"id"`
		SessionID int64     `json:"session_id"`
		Room      string    `json:"room"`
		Member    string    `json:"member"`
		PT        string    `json:"pt"`
		Status    string    `json:"status"`
		StartsAt  time.Time `json:"starts_at"`
		EndsAt    time.Time `json:"ends_at"`
	}
	manifest := struct {
		GeneratedAt time.Time       `json:"generated_at"`
		Database    string          `json:"database"`
		Members     []account       `json:"members"`
		Trainers    []account       `json:"trainers"`
		Rooms       []roomRecord    `json:"rooms"`
		Bookings    []bookingRecord `json:"bookings"`
	}{GeneratedAt: time.Now(), Database: localDemoDatabase, Bookings: []bookingRecord{}}
	for _, member := range members {
		manifest.Members = append(manifest.Members, account{member.ID, member.Username})
	}
	for _, trainer := range trainers {
		manifest.Trainers = append(manifest.Trainers, account{trainer.ID, trainer.Username})
	}
	for _, room := range rooms {
		manifest.Rooms = append(manifest.Rooms, roomRecord{room.ID, room.Name, room.Status})
	}
	rows, err := pool.Query(ctx, `SELECT a.reason,b.id,s.id,r.name,mu.username,COALESCE(pu.username,''),b.status,s.starts_at,s.ends_at
		FROM audit_events a JOIN bookings b ON b.id::text=a.entity_id
		JOIN room_sessions s ON s.id=b.room_session_id JOIN rooms r ON r.id=s.room_id
		JOIN members m ON m.id=b.member_id JOIN users mu ON mu.id=m.user_id
		LEFT JOIN personal_trainers p ON p.id=b.pt_id LEFT JOIN users pu ON pu.id=p.user_id
		WHERE a.action=$1 ORDER BY s.starts_at,b.id`, demoAuditAction)
	if err != nil {
		return "", err
	}
	defer rows.Close()
	for rows.Next() {
		var item bookingRecord
		if err := rows.Scan(&item.Code, &item.ID, &item.SessionID, &item.Room, &item.Member, &item.PT, &item.Status, &item.StartsAt, &item.EndsAt); err != nil {
			return "", err
		}
		manifest.Bookings = append(manifest.Bookings, item)
	}
	if err := rows.Err(); err != nil {
		return "", err
	}
	directory := filepath.Join("..", ".artifact_work", "demo-manifests")
	if err := os.MkdirAll(directory, 0700); err != nil {
		return "", err
	}
	path := filepath.Join(directory, "demo-"+time.Now().Format("20060102-150405.000000000")+".json")
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return "", err
	}
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(manifest); err != nil {
		file.Close()
		return "", err
	}
	if err := file.Close(); err != nil {
		return "", err
	}
	return path, nil
}

func validateLocalDevURL(databaseURL string) error {
	u, err := url.Parse(databaseURL)
	if err != nil {
		return fmt.Errorf("invalid DATABASE_URL: %w", err)
	}
	if u.Hostname() != "127.0.0.1" && u.Hostname() != "localhost" {
		return fmt.Errorf("refuse non-local database host %q", u.Hostname())
	}
	if strings.TrimPrefix(u.EscapedPath(), "/") != localDemoDatabase {
		return fmt.Errorf("refuse database %q; this command only accepts local %s", u.Path, localDemoDatabase)
	}
	return nil
}

func requireActiveRole(ctx context.Context, pool *pgxpool.Pool, username, role string) (int64, error) {
	var id int64
	err := pool.QueryRow(ctx, `
		SELECT u.id
		FROM users u
		JOIN roles r ON r.id=u.role_id
		WHERE LOWER(u.username)=LOWER($1) AND u.status='ACTIVE' AND r.code=$2
	`, username, role).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, fmt.Errorf("run go run .\\cmd\\seed first; active %s with role %s is required", username, role)
	}
	if err != nil {
		return 0, fmt.Errorf("look up %s: %w", username, err)
	}
	return id, nil
}

func ensureProfiles(ctx context.Context, pool *pgxpool.Pool, adminID int64, passwordHash string, counts *seedCounts) ([]memberFixture, []trainerFixture, error) {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return nil, nil, err
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(841029)`); err != nil {
		return nil, nil, err
	}

	var today time.Time
	if err := tx.QueryRow(ctx, `SELECT (clock_timestamp() AT TIME ZONE $1)::date`, localVietnamTZName).Scan(&today); err != nil {
		return nil, nil, fmt.Errorf("read Vietnam date: %w", err)
	}
	members := memberFixtures()
	for i := range members {
		userID, created, err := ensureUser(ctx, tx, members[i].Username, members[i].FullName, members[i].Email, "MEMBER", map[bool]string{true: "LOCKED", false: "ACTIVE"}[members[i].Locked], passwordHash)
		if err != nil {
			return nil, nil, err
		}
		members[i].UserID = userID
		memberID, profileCreated, err := ensureMember(ctx, tx, userID, members[i], today)
		if err != nil {
			return nil, nil, err
		}
		members[i].ID = memberID
		if created || profileCreated {
			counts.MembersCreated++
		}
	}

	trainers := trainerFixtures()
	for i := range trainers {
		userID, created, err := ensureUser(ctx, tx, trainers[i].Username, trainers[i].FullName, trainers[i].Email, "PT", "ACTIVE", passwordHash)
		if err != nil {
			return nil, nil, err
		}
		trainers[i].UserID = userID
		ptID, profileCreated, err := ensureTrainer(ctx, tx, userID, adminID, trainers[i])
		if err != nil {
			return nil, nil, err
		}
		trainers[i].ID = ptID
		if err := ensureTrainerHours(ctx, tx, ptID, profileCreated); err != nil {
			return nil, nil, err
		}
		if created || profileCreated {
			counts.TrainersCreated++
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, nil, fmt.Errorf("commit demo profiles: %w", err)
	}
	return members, trainers, nil
}

func ensureUser(ctx context.Context, tx pgx.Tx, username, fullName, email, role, status, passwordHash string) (int64, bool, error) {
	var id int64
	err := tx.QueryRow(ctx, `
		INSERT INTO users(username,password_hash,email,full_name,role_id,status)
		SELECT $1,$2,$3,$4,id,$5 FROM roles WHERE code=$6
		ON CONFLICT (LOWER(username)) DO NOTHING
		RETURNING users.id
	`, username, passwordHash, email, fullName, status, role).Scan(&id)
	if err == nil {
		return id, true, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return 0, false, fmt.Errorf("create demo user %s: %w", username, err)
	}

	var actualName, actualEmail, actualRole, actualStatus string
	err = tx.QueryRow(ctx, `
		SELECT u.id,u.full_name,u.email,r.code,u.status
		FROM users u JOIN roles r ON r.id=u.role_id
		WHERE LOWER(u.username)=LOWER($1)
	`, username).Scan(&id, &actualName, &actualEmail, &actualRole, &actualStatus)
	if err != nil {
		return 0, false, fmt.Errorf("read existing demo user %s: %w", username, err)
	}
	if actualName != fullName || actualEmail != email || actualRole != role || actualStatus != status {
		return 0, false, fmt.Errorf("existing account %s differs from the DEMO fixture; no overwrite was made", username)
	}
	return id, false, nil
}

func ensureMember(ctx context.Context, tx pgx.Tx, userID int64, fixture memberFixture, today time.Time) (int64, bool, error) {
	var id int64
	var code, status string
	var expiry time.Time
	err := tx.QueryRow(ctx, `SELECT id,member_code,status,membership_expiry FROM members WHERE user_id=$1`, userID).Scan(&id, &code, &status, &expiry)
	if err == nil {
		validExpiry := expiry.Before(today)
		if !fixture.Expired {
			validExpiry = !expiry.Before(today.AddDate(0, 0, 14))
		}
		if code != fixture.Code || status != fixture.Status || !validExpiry {
			return 0, false, fmt.Errorf("existing member profile for %s differs from the DEMO fixture; no overwrite was made", fixture.Username)
		}
		return id, false, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return 0, false, fmt.Errorf("read member %s: %w", fixture.Username, err)
	}

	joinDate := today.AddDate(-1, 0, 0)
	expiry = today.AddDate(1, 0, 0)
	if fixture.Expired {
		expiry = today.AddDate(0, 0, -3)
		joinDate = today.AddDate(-1, 0, 0)
	}
	err = tx.QueryRow(ctx, `
		INSERT INTO members(user_id,member_code,join_date,membership_expiry,status)
		VALUES($1,$2,$3::date,$4::date,$5) RETURNING id
	`, userID, fixture.Code, joinDate.Format("2006-01-02"), expiry.Format("2006-01-02"), fixture.Status).Scan(&id)
	if err != nil {
		return 0, false, fmt.Errorf("create member profile %s: %w", fixture.Username, err)
	}
	return id, true, nil
}

func ensureTrainer(ctx context.Context, tx pgx.Tx, userID, adminID int64, fixture trainerFixture) (int64, bool, error) {
	var id int64
	var approval string
	var decisionReason *string
	err := tx.QueryRow(ctx, `SELECT id,approval_status,decision_reason FROM personal_trainers WHERE user_id=$1`, userID).Scan(&id, &approval, &decisionReason)
	if err == nil {
		if approval != fixture.Approval || (fixture.Approval == "APPROVED" && (decisionReason == nil || *decisionReason != demoTrainerMarker)) {
			return 0, false, fmt.Errorf("existing PT profile for %s differs from the DEMO fixture; no overwrite was made", fixture.Username)
		}
		return id, false, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return 0, false, fmt.Errorf("read PT %s: %w", fixture.Username, err)
	}

	if fixture.Approval == "APPROVED" {
		err = tx.QueryRow(ctx, `
			INSERT INTO personal_trainers(user_id,specialty,experience_text,approval_status,decision_reason,decided_by_user_id,decided_at)
			VALUES($1,$2,$3,'APPROVED',$4,$5,clock_timestamp()) RETURNING id
		`, userID, fixture.Specialty, fixture.Experience, demoTrainerMarker, adminID).Scan(&id)
	} else {
		err = tx.QueryRow(ctx, `
			INSERT INTO personal_trainers(user_id,specialty,experience_text,approval_status)
			VALUES($1,$2,$3,'PENDING') RETURNING id
		`, userID, fixture.Specialty, fixture.Experience).Scan(&id)
	}
	if err != nil {
		return 0, false, fmt.Errorf("create PT profile %s: %w", fixture.Username, err)
	}
	return id, true, nil
}

func ensureTrainerHours(ctx context.Context, tx pgx.Tx, ptID int64, newlyCreated bool) error {
	var coveredDays int
	err := tx.QueryRow(ctx, `
		SELECT count(DISTINCT day_of_week)
		FROM pt_working_hours
		WHERE pt_id=$1 AND is_active AND start_time<='06:00'::time AND end_time>='21:00'::time
	`, ptID).Scan(&coveredDays)
	if err != nil {
		return err
	}
	if coveredDays == 7 {
		return nil
	}
	if !newlyCreated {
		return fmt.Errorf("existing DEMO PT %d does not have all-week 06:00–21:00 hours; no overwrite was made", ptID)
	}
	_, err = tx.Exec(ctx, `
		INSERT INTO pt_working_hours(pt_id,day_of_week,start_time,end_time)
		SELECT $1,d,'06:00'::time,'21:00'::time FROM generate_series(1,7) d
	`, ptID)
	return err
}

func ensureUsableSlots(ctx context.Context, pool *pgxpool.Pool, actor identity.User) error {
	var total int
	var adjacent bool
	err := pool.QueryRow(ctx, `
		SELECT (SELECT count(*) FROM time_slots WHERE is_active),
		       EXISTS(SELECT 1 FROM time_slots a JOIN time_slots b ON a.end_time=b.start_time
		              WHERE a.is_active AND b.is_active AND a.start_time>='06:00'::time AND b.end_time<='21:00'::time)
	`).Scan(&total, &adjacent)
	if err != nil {
		return fmt.Errorf("read time slots: %w", err)
	}
	if adjacent {
		return nil
	}
	if total > 0 {
		var start int
		err = pool.QueryRow(ctx, `SELECT h FROM generate_series(6,19) h
			WHERE NOT EXISTS (SELECT 1 FROM time_slots t WHERE t.is_active
			AND t.start_time < time '00:00' + make_interval(hours => h+2)
			AND t.end_time > time '00:00' + make_interval(hours => h)) ORDER BY h LIMIT 1`).Scan(&start)
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("no free two-hour interval for adjacent DEMO slots; existing slots were not changed")
		}
		if err != nil {
			return fmt.Errorf("find free demo slots: %w", err)
		}
		facilities := service.NewFacilityService(repository.NewFacilityRepository(pool))
		for hour := start; hour < start+2; hour++ {
			from, to := fmt.Sprintf("%02d:00", hour), fmt.Sprintf("%02d:00", hour+1)
			if _, err := facilities.CreateSlot(ctx, actor, facility.SlotInput{Name: "DEMO · " + from + "–" + to, StartTime: from, EndTime: to, Reason: "Tạo ca liền kề cho DEMO"}, demoRequestID); err != nil {
				return fmt.Errorf("create adjacent demo slot %s: %w", from, err)
			}
		}
		return nil
	}

	facilities := service.NewFacilityService(repository.NewFacilityRepository(pool))
	for _, slot := range []struct{ name, start, end string }{
		{"DEMO · 06:00–07:00", "06:00", "07:00"},
		{"DEMO · 07:00–08:00", "07:00", "08:00"},
		{"DEMO · 08:00–09:00", "08:00", "09:00"},
		{"DEMO · 09:00–10:00", "09:00", "10:00"},
		{"DEMO · 17:00–18:00", "17:00", "18:00"},
		{"DEMO · 18:00–19:00", "18:00", "19:00"},
	} {
		if _, err := facilities.CreateSlot(ctx, actor, facility.SlotInput{Name: slot.name, StartTime: slot.start, EndTime: slot.end, Reason: "Tạo ca cho dữ liệu DEMO cục bộ"}, demoRequestID); err != nil {
			return fmt.Errorf("create demo time slot %s: %w", slot.name, err)
		}
	}
	return nil
}

func ensureRooms(ctx context.Context, pool *pgxpool.Pool, actor identity.User, counts *seedCounts) ([]roomFixture, error) {
	facilities := service.NewFacilityService(repository.NewFacilityRepository(pool))
	covers := service.NewRoomCoverService(repository.NewRoomCoverRepository(pool))
	calendar := service.NewCalendarService(repository.NewCalendarRepository(pool))
	rooms := roomFixtures()
	for i := range rooms {
		var id int64
		var roomType, description, status string
		var capacity int
		err := pool.QueryRow(ctx, `SELECT id,room_type,default_capacity,COALESCE(description,''),status FROM rooms WHERE LOWER(name)=LOWER($1)`, rooms[i].Name).Scan(&id, &roomType, &capacity, &description, &status)
		created := false
		if errors.Is(err, pgx.ErrNoRows) {
			room, createErr := facilities.SaveRoom(ctx, actor, 0, facility.RoomInput{
				Name: rooms[i].Name, RoomType: rooms[i].RoomType, DefaultCapacity: rooms[i].Capacity,
				Status: rooms[i].Status, Description: demoRoomDescription(rooms[i].Name),
			}, demoRequestID)
			if createErr != nil {
				return nil, fmt.Errorf("create demo room %s: %w", rooms[i].Name, createErr)
			}
			id, created = room.ID, true
			counts.RoomsCreated++
		} else if err != nil {
			return nil, fmt.Errorf("read demo room %s: %w", rooms[i].Name, err)
		} else if roomType != rooms[i].RoomType || capacity != rooms[i].Capacity || status != rooms[i].Status || description != demoRoomDescription(rooms[i].Name) {
			return nil, fmt.Errorf("existing room %s differs from the DEMO fixture; no overwrite was made", rooms[i].Name)
		}
		rooms[i].ID = id

		var hasCover bool
		if err := pool.QueryRow(ctx, `SELECT cover_image IS NOT NULL FROM rooms WHERE id=$1`, id).Scan(&hasCover); err != nil {
			return nil, fmt.Errorf("read cover for %s: %w", rooms[i].Name, err)
		}
		if !hasCover {
			file, err := openDemoCover(rooms[i].Cover)
			if err != nil {
				return nil, err
			}
			_, uploadErr := covers.Upload(ctx, actor, id, file, rooms[i].Alt, true, demoRequestID)
			closeErr := file.Close()
			if uploadErr != nil {
				return nil, fmt.Errorf("upload demo cover for %s: %w", rooms[i].Name, uploadErr)
			}
			if closeErr != nil {
				return nil, fmt.Errorf("close demo cover for %s: %w", rooms[i].Name, closeErr)
			}
		}
		if rooms[i].Status != "AVAILABLE" {
			continue
		}
		if err := ensureRoomCalendar(ctx, pool, calendar, actor, id, created); err != nil {
			return nil, fmt.Errorf("prepare calendar for %s: %w", rooms[i].Name, err)
		}
	}
	return rooms, nil
}

func ensureRoomCalendar(ctx context.Context, pool *pgxpool.Pool, calendar *service.CalendarService, actor identity.User, roomID int64, newlyCreated bool) error {
	var hasSchedule bool
	if err := pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM room_schedule_versions WHERE room_id=$1 AND status='ACTIVE')`, roomID).Scan(&hasSchedule); err != nil {
		return err
	}
	if !hasSchedule {
		snapshot, err := calendar.Snapshot(ctx, actor, roomID)
		if err != nil {
			return err
		}
		mutation := facility.CalendarMutation{
			Kind: "WEEKLY", BaseRevision: snapshot.Room.CalendarRevision,
			EffectiveFrom: facility.Today(time.Now()), Days: 28,
			Reason: "Lịch mở cửa cho dữ liệu DEMO cục bộ",
		}
		for weekday := 1; weekday <= 7; weekday++ {
			mutation.Hours = append(mutation.Hours, facility.Hours{Weekday: weekday, Start: 360, End: 1260})
		}
		preview, err := calendar.Preview(ctx, actor, roomID, mutation)
		if err != nil {
			return err
		}
		_, err = calendar.Confirm(ctx, actor, roomID, facility.CalendarConfirm{CalendarMutation: mutation, PreviewID: preview.ID}, fmt.Sprintf("demo-scenarios-room-%d", roomID), demoRequestID)
		if err != nil {
			return err
		}
	}
	var prepared bool
	if err := pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM room_sessions WHERE room_id=$1 AND status='OPEN' AND session_date>=((clock_timestamp() AT TIME ZONE $2)::date+14))`, roomID, localVietnamTZName).Scan(&prepared); err != nil {
		return err
	}
	if !prepared {
		snapshot, err := calendar.Snapshot(ctx, actor, roomID)
		if err != nil {
			return err
		}
		mutation := facility.CalendarMutation{Kind: "EXTEND", BaseRevision: snapshot.Room.CalendarRevision, Days: 28, Reason: "Gia hạn phiên cho DEMO cục bộ"}
		preview, err := calendar.Preview(ctx, actor, roomID, mutation)
		if err != nil {
			return fmt.Errorf("preview demo session extension: %w", err)
		}
		key := fmt.Sprintf("demo-extend-%d-%s-%d", roomID, facility.Today(time.Now()), snapshot.Room.CalendarRevision)
		if _, err := calendar.Confirm(ctx, actor, roomID, facility.CalendarConfirm{CalendarMutation: mutation, PreviewID: preview.ID}, key, demoRequestID); err != nil {
			return fmt.Errorf("extend demo sessions: %w", err)
		}
	}
	var hasFutureSession bool
	if err := pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM room_sessions WHERE room_id=$1 AND status='OPEN' AND ends_at>clock_timestamp())`, roomID).Scan(&hasFutureSession); err != nil {
		return err
	}
	if !hasFutureSession {
		if newlyCreated {
			return fmt.Errorf("new DEMO room has no generated future session; inspect active time slots")
		}
		return fmt.Errorf("existing DEMO room has no future session; no calendar was overwritten")
	}
	return nil
}

func canRefreshPending(fixture bookingFixture, status string, requested bool) bool {
	return requested && fixture.Status == "PENDING_PT" && (status == "EXPIRED" || status == "REJECTED" || status == "CANCELLED")
}

func ensureBookingFixtures(ctx context.Context, pool *pgxpool.Pool, adminID, receptionID int64, members []memberFixture, trainers []trainerFixture, rooms []roomFixture, refreshPending bool) (int, int, error) {
	var today time.Time
	if err := pool.QueryRow(ctx, `SELECT (clock_timestamp() AT TIME ZONE $1)::date`, localVietnamTZName).Scan(&today); err != nil {
		return 0, 0, err
	}

	created, existing := 0, 0
	bookingRepo := repository.NewBookingRepository(pool)
	// Commit historical session fixtures before business booking operations.
	// Keeping their FK locks in a live-creation transaction blocks room locks.
	for _, live := range []bool{false, true} {
		tx, err := pool.Begin(ctx)
		if err != nil {
			return created, existing, err
		}
		defer tx.Rollback(ctx)
		if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(841030)`); err != nil {
			return created, existing, err
		}
		for _, fixture := range bookingFixtures() {
			if live != (fixture.DayOffset >= 0) {
				continue
			}
			if fixture.RoomIndex >= len(rooms) || fixture.MemberIndex >= len(members) || fixture.TrainerIndex >= len(trainers) {
				return 0, 0, fmt.Errorf("invalid in-code booking fixture %s", fixture.Code)
			}
			date := today.AddDate(0, 0, fixture.DayOffset)
			fixtureKey := fixture.Code
			if fixture.DayOffset >= 0 {
				fixtureKey += "@" + date.Format("2006-01-02")
			}
			var previousID int64
			var previousStatus string
			err := tx.QueryRow(ctx, `
			SELECT b.id,b.status
			FROM audit_events a
			JOIN bookings b ON b.id::text=a.entity_id
			WHERE a.action=$1 AND (a.reason=$2 OR a.reason LIKE $2 || '#%' OR (a.reason=$3 AND b.session_date=$4::date))
			ORDER BY a.id DESC LIMIT 1
		`, demoAuditAction, fixtureKey, fixture.Code, date.Format("2006-01-02")).Scan(&previousID, &previousStatus)
			if err == nil {
				log.Printf("Preserving DEMO booking %s: id=%d status=%s", fixtureKey, previousID, previousStatus)
				existing++
				if !canRefreshPending(fixture, previousStatus, refreshPending) {
					continue
				}
				var generations int
				if err := tx.QueryRow(ctx, `SELECT count(*) FROM audit_events WHERE action=$1 AND (reason=$2 OR reason LIKE $2 || '#%')`, demoAuditAction, fixtureKey).Scan(&generations); err != nil {
					return created, existing, err
				}
				fixtureKey = fmt.Sprintf("%s#%d", fixtureKey, generations+1)
			}
			if err != nil && !errors.Is(err, pgx.ErrNoRows) {
				return 0, 0, fmt.Errorf("read booking fixture %s: %w", fixture.Code, err)
			}

			sessionID, err := fixtureSession(ctx, tx, rooms[fixture.RoomIndex], date, today, adminID)
			if err != nil {
				return 0, 0, fmt.Errorf("prepare session for %s: %w", fixture.Code, err)
			}
			var bookingID int64
			if fixture.DayOffset >= 0 {
				member := members[fixture.MemberIndex]
				coach := trainers[fixture.TrainerIndex]
				actor := identity.User{ID: member.UserID, Role: identity.RoleMember}
				booked, createErr := bookingRepo.Create(ctx, actor, 0, booking.CreateInput{RoomSessionID: sessionID, PTID: coach.ID}, "demo:"+fixtureKey, demoRequestID)
				if createErr != nil {
					return 0, 0, fmt.Errorf("create live demo booking %s: %w", fixtureKey, createErr)
				}
				bookingID = booked.ID
				if fixture.Status == "CONFIRMED" {
					var currentStatus string
					if err := tx.QueryRow(ctx, `SELECT status FROM bookings WHERE id=$1`, bookingID).Scan(&currentStatus); err != nil {
						return 0, 0, err
					}
					if currentStatus == "PENDING_PT" {
						_, err = bookingRepo.Act(ctx, identity.User{ID: coach.UserID, Role: identity.RolePT}, bookingID, booking.ActionInput{Action: "ACCEPT"}, false, demoRequestID)
						if err != nil {
							return 0, 0, fmt.Errorf("confirm live demo booking %s: %w", fixtureKey, err)
						}
					}
				}
			} else {
				bookingID, err = createFixtureBooking(ctx, tx, fixture, members[fixture.MemberIndex], trainers[fixture.TrainerIndex], sessionID, adminID)
			}
			if err != nil {
				return 0, 0, fmt.Errorf("create booking %s: %w", fixture.Code, err)
			}
			if fixture.MemberAttendance {
				if err := createAttendance(ctx, tx, bookingID, "MEMBER", members[fixture.MemberIndex].UserID, receptionID); err != nil {
					return 0, 0, err
				}
			}
			if fixture.TrainerAttendance {
				if err := createAttendance(ctx, tx, bookingID, "PT", trainers[fixture.TrainerIndex].UserID, receptionID); err != nil {
					return 0, 0, err
				}
			}
			if fixture.CreateReview {
				if _, err := tx.Exec(ctx, `INSERT INTO pt_reviews(booking_id,rating,comment) VALUES($1,5,'Đánh giá DEMO sau buổi hoàn thành.') ON CONFLICT (booking_id) DO NOTHING`, bookingID); err != nil {
					return 0, 0, fmt.Errorf("create review for %s: %w", fixture.Code, err)
				}
			}
			if _, err := tx.Exec(ctx, `
			INSERT INTO audit_events(actor_user_id,action,entity_type,entity_id,reason,request_id)
			VALUES($1,$2,'booking',$3,$4,$5)
		`, adminID, demoAuditAction, fmt.Sprint(bookingID), fixtureKey, demoRequestID); err != nil {
				return 0, 0, fmt.Errorf("audit booking %s: %w", fixture.Code, err)
			}
			created++
		}
		if err := tx.Commit(ctx); err != nil {
			return created, existing, fmt.Errorf("commit booking fixtures: %w", err)
		}
	}
	return created, existing, nil
}

func fixtureSession(ctx context.Context, tx pgx.Tx, room roomFixture, date, today time.Time, adminID int64) (int64, error) {
	var id int64
	if !date.Before(today) {
		err := tx.QueryRow(ctx, `
			SELECT id FROM room_sessions
			WHERE room_id=$1 AND session_date=$2::date AND status='OPEN' AND ends_at>clock_timestamp()
			ORDER BY starts_at,id LIMIT 1
		`, room.ID, date.Format("2006-01-02")).Scan(&id)
		if err != nil {
			return 0, fmt.Errorf("no open future session for %s on %s", room.Name, date.Format("2006-01-02"))
		}
		return id, nil
	}

	var slotID int64
	err := tx.QueryRow(ctx, `
		SELECT time_slot_id FROM room_sessions
		WHERE room_id=$1 AND session_date>=$2::date AND status='OPEN'
		ORDER BY starts_at,id LIMIT 1
	`, room.ID, today.AddDate(0, 0, 1).Format("2006-01-02")).Scan(&slotID)
	if err != nil {
		return 0, fmt.Errorf("find reusable slot for %s: %w", room.Name, err)
	}
	_, err = tx.Exec(ctx, `
		INSERT INTO room_sessions(room_id,time_slot_id,session_date,starts_at,ends_at,capacity,status,created_by_user_id,change_reason,origin)
		SELECT r.id,t.id,$2::date,
		       ($2::date+t.start_time) AT TIME ZONE $3,
		       ($2::date+t.end_time) AT TIME ZONE $3,
		       r.default_capacity,'OPEN',$4,'Phiên lịch sử DEMO','MANUAL'
		FROM rooms r JOIN time_slots t ON t.id=$5
		WHERE r.id=$1
		ON CONFLICT (room_id,session_date,time_slot_id) DO NOTHING
	`, room.ID, date.Format("2006-01-02"), localVietnamTZName, adminID, slotID)
	if err != nil {
		return 0, err
	}
	err = tx.QueryRow(ctx, `SELECT id FROM room_sessions WHERE room_id=$1 AND session_date=$2::date AND time_slot_id=$3`, room.ID, date.Format("2006-01-02"), slotID).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func createFixtureBooking(ctx context.Context, tx pgx.Tx, fixture bookingFixture, member memberFixture, trainer trainerFixture, sessionID, adminID int64) (int64, error) {
	cancelledBy := int64(0)
	if fixture.Status == "CANCELLED" {
		if fixture.CancellationSource == "MEMBER" {
			cancelledBy = member.UserID
		} else {
			cancelledBy = adminID
		}
	}
	var id int64
	err := tx.QueryRow(ctx, `
		INSERT INTO bookings(
			member_id,room_session_id,session_date,time_slot_id,pt_id,status,respond_by,created_by_user_id,
			confirmed_at,cancelled_by_user_id,cancellation_reason,cancelled_at,cancellation_source,cancelled_actor_role,created_at,updated_at
		)
		SELECT $1,s.id,s.session_date,s.time_slot_id,$2,$3::varchar,
			CASE WHEN $3::varchar IN ('PENDING_PT','EXPIRED') THEN s.starts_at-interval '30 minutes' END,
			$4,
			CASE WHEN $3::varchar IN ('CONFIRMED','COMPLETED','NO_SHOW','CANCELLED') THEN LEAST(clock_timestamp(),s.starts_at-interval '30 minutes') END,
			NULLIF($5,0),NULLIF($6,''),
			CASE WHEN $3::varchar='CANCELLED' THEN LEAST(clock_timestamp(),s.starts_at-interval '15 minutes') END,
			NULLIF($7,''),NULLIF($8,''),s.starts_at-interval '1 hour',
			CASE WHEN $3::varchar IN ('COMPLETED','NO_SHOW') THEN s.ends_at+interval '1 minute' ELSE s.starts_at-interval '15 minutes' END
		FROM room_sessions s WHERE s.id=$9
		RETURNING id
	`, member.ID, trainer.ID, fixture.Status, member.UserID, cancelledBy, fixture.CancellationReason, fixture.CancellationSource, fixture.CancellationRole, sessionID).Scan(&id)
	return id, err
}

func createAttendance(ctx context.Context, tx pgx.Tx, bookingID int64, participantType string, participantUserID, receptionID int64) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO booking_attendance(booking_id,participant_type,participant_user_id,checked_in_at,recorded_by_user_id,override_reason)
		SELECT b.id,$2,$3,LEAST(clock_timestamp(),s.starts_at+interval '10 minutes'),$4,'Dữ liệu DEMO'
		FROM bookings b JOIN room_sessions s ON s.id=b.room_session_id
		WHERE b.id=$1
		ON CONFLICT (booking_id,participant_type) DO NOTHING
	`, bookingID, participantType, participantUserID, receptionID)
	return err
}

func openDemoCover(name string) (*os.File, error) {
	candidates := []string{
		filepath.Join("..", "docs", "reports", "assets", "booking-portals", "photos", name),
		filepath.Join("docs", "reports", "assets", "booking-portals", "photos", name),
	}
	for _, candidate := range candidates {
		file, err := os.Open(candidate)
		if err == nil {
			return file, nil
		}
		if !os.IsNotExist(err) {
			return nil, err
		}
	}
	return nil, fmt.Errorf("demo image %s is missing; run from server or repository root", name)
}

func demoRoomDescription(name string) string {
	return demoRoomMarker + " " + name
}

func memberFixtures() []memberFixture {
	items := make([]memberFixture, 0, 12)
	for i := 1; i <= 10; i++ {
		items = append(items, memberFixture{
			Username: fmt.Sprintf("demo.member.%02d", i), FullName: fmt.Sprintf("DEMO · Hội viên %02d", i),
			Email: fmt.Sprintf("demo.member.%02d@demo.local", i), Code: fmt.Sprintf("DEMO-M%02d", i), Status: "ACTIVE",
		})
	}
	return append(items,
		memberFixture{Username: "demo.member.expired", FullName: "DEMO · Hội viên hết hạn", Email: "demo.member.expired@demo.local", Code: "DEMO-M-EXP", Status: "ACTIVE", Expired: true},
		memberFixture{Username: "demo.member.locked", FullName: "DEMO · Hội viên bị khóa", Email: "demo.member.locked@demo.local", Code: "DEMO-M-LOCK", Status: "ACTIVE", Locked: true},
	)
}

func trainerFixtures() []trainerFixture {
	return []trainerFixture{
		{Username: "demo.pt.strength", FullName: "DEMO · HLV Sức mạnh", Email: "demo.pt.strength@demo.local", Specialty: "Strength & Conditioning", Experience: "HLV minh họa luồng đặt PT", Approval: "APPROVED"},
		{Username: "demo.pt.cardio", FullName: "DEMO · HLV Cardio", Email: "demo.pt.cardio@demo.local", Specialty: "Cardio & Endurance", Experience: "HLV minh họa luồng đặt PT", Approval: "APPROVED"},
		{Username: "demo.pt.mobility", FullName: "DEMO · HLV Mobility", Email: "demo.pt.mobility@demo.local", Specialty: "Mobility & Recovery", Experience: "HLV minh họa luồng đặt PT", Approval: "APPROVED"},
		{Username: "demo.pt.pending", FullName: "DEMO · HLV chờ duyệt", Email: "demo.pt.pending@demo.local", Specialty: "Yoga cơ bản", Experience: "Hồ sơ minh họa trạng thái chờ duyệt", Approval: "PENDING"},
	}
}

func roomFixtures() []roomFixture {
	return []roomFixture{
		{Name: "DEMO · Strength Arena", RoomType: "STRENGTH", Capacity: 18, Status: "AVAILABLE", Cover: "strength-demo.png", Alt: "Phòng DEMO tập sức mạnh"},
		{Name: "DEMO · Functional Lab", RoomType: "FUNCTIONAL", Capacity: 14, Status: "AVAILABLE", Cover: "strength-demo.png", Alt: "Phòng DEMO tập chức năng"},
		{Name: "DEMO · Cardio Studio", RoomType: "CARDIO", Capacity: 16, Status: "AVAILABLE", Cover: "cardio-demo.png", Alt: "Phòng DEMO cardio"},
		{Name: "DEMO · Recovery Flow", RoomType: "RECOVERY", Capacity: 10, Status: "AVAILABLE", Cover: "cardio-demo.png", Alt: "Phòng DEMO hồi phục"},
		{Name: "DEMO · PT Studio A", RoomType: "PT", Capacity: 6, Status: "AVAILABLE", Cover: "pt-demo.png", Alt: "Phòng DEMO huấn luyện cá nhân"},
		{Name: "DEMO · PT Studio B", RoomType: "PT", Capacity: 6, Status: "MAINTENANCE", Cover: "pt-demo.png", Alt: "Phòng DEMO đang bảo trì"},
	}
}

func bookingFixtures() []bookingFixture {
	fixtures := make([]bookingFixture, 0, 24)
	for i := 0; i < 6; i++ {
		fixtures = append(fixtures, bookingFixture{Code: fmt.Sprintf("DEMO-PENDING-%02d", i+1), Status: "PENDING_PT", DayOffset: i + 1, RoomIndex: i % 5, MemberIndex: i, TrainerIndex: i % 3})
	}
	for i := 0; i < 8; i++ {
		fixtures = append(fixtures, bookingFixture{Code: fmt.Sprintf("DEMO-CONFIRMED-%02d", i+1), Status: "CONFIRMED", DayOffset: i + 7, RoomIndex: i % 5, MemberIndex: (i + 2) % 10, TrainerIndex: i % 3})
	}
	for i := 0; i < 4; i++ {
		fixtures = append(fixtures, bookingFixture{Code: fmt.Sprintf("DEMO-COMPLETED-%02d", i+1), Status: "COMPLETED", DayOffset: -10 + i, RoomIndex: i % 5, MemberIndex: (i + 4) % 10, TrainerIndex: i % 3, MemberAttendance: true, TrainerAttendance: true, CreateReview: i < 2})
	}
	fixtures = append(fixtures,
		bookingFixture{Code: "DEMO-NO-SHOW-01", Status: "NO_SHOW", DayOffset: -6, RoomIndex: 0, MemberIndex: 8, TrainerIndex: 1, TrainerAttendance: true},
		bookingFixture{Code: "DEMO-NO-SHOW-02", Status: "NO_SHOW", DayOffset: -5, RoomIndex: 1, MemberIndex: 9, TrainerIndex: 2},
		bookingFixture{Code: "DEMO-CANCELLED-01", Status: "CANCELLED", DayOffset: -4, RoomIndex: 2, MemberIndex: 0, TrainerIndex: 0, CancellationSource: "MEMBER", CancellationRole: "MEMBER", CancellationReason: "Dữ liệu DEMO: hội viên hủy lịch."},
		bookingFixture{Code: "DEMO-CANCELLED-02", Status: "CANCELLED", DayOffset: -3, RoomIndex: 3, MemberIndex: 1, TrainerIndex: 1, CancellationSource: "FACILITY", CancellationRole: "ADMIN", CancellationReason: "Dữ liệu DEMO: cơ sở đổi lịch phòng."},
		bookingFixture{Code: "DEMO-REJECTED-01", Status: "REJECTED", DayOffset: -2, RoomIndex: 4, MemberIndex: 2, TrainerIndex: 2},
		bookingFixture{Code: "DEMO-EXPIRED-01", Status: "EXPIRED", DayOffset: -1, RoomIndex: 0, MemberIndex: 3, TrainerIndex: 0},
	)
	sort.Slice(fixtures, func(i, j int) bool { return fixtures[i].Code < fixtures[j].Code })
	return fixtures
}
