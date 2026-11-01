package v1service

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/facility"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/identity"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/repository"
	"strings"
	"time"
)

type CalendarService struct {
	repo repository.CalendarRepository
	key  [32]byte
	now  func() time.Time
}

func NewCalendarService(repo repository.CalendarRepository) *CalendarService {
	s := &CalendarService{repo: repo, now: time.Now}
	if _, err := rand.Read(s.key[:]); err != nil {
		panic("calendar signing key unavailable")
	}
	return s
}

type previewClaims struct {
	Actor, Room       int64
	Hash, Fingerprint string
	Expires           time.Time
}

func mutationHash(m facility.CalendarMutation) string {
	b, _ := json.Marshal(m)
	return fmt.Sprintf("%x", sha256.Sum256(b))
}
func (s *CalendarService) normalize(in facility.CalendarMutation, replay ...bool) (facility.CalendarMutation, error) {
	in.Reason = strings.TrimSpace(in.Reason)
	if in.BaseRevision < 0 || len(in.Reason) == 0 || len(in.Reason) > 1000 || in.Days < 0 || in.Days > 56 {
		return in, identity.ErrValidation
	}
	if in.Days == 0 {
		in.Days = 28
	}
	today := facility.Today(s.now())
	upper := s.now().In(facility.Vietnam).AddDate(0, 0, 56).Format("2006-01-02")
	validDate := func(date string) bool {
		return dateValid(date) && ((len(replay) > 0 && replay[0]) || (date >= today && date <= upper))
	}
	var err error
	switch in.Kind {
	case "WEEKLY":
		if !validDate(in.EffectiveFrom) || in.Date != "" || in.ExceptionID != 0 || in.Mode != "" || in.ClosureType != "" {
			return in, identity.ErrValidation
		}
		in.Hours, err = facility.NormalizeHours(in.Hours, true)
	case "EXCEPTION":
		if !validDate(in.Date) || in.EffectiveFrom != "" || in.ExceptionID < 0 || !oneOf(in.Mode, "CLOSED", "REPLACE") || !oneOf(in.ClosureType, "HOLIDAY", "MAINTENANCE", "OTHER") || (in.Mode == "CLOSED" && len(in.Hours) > 0) || (in.Mode == "REPLACE" && len(in.Hours) == 0) {
			return in, identity.ErrValidation
		}
		in.Hours, err = facility.NormalizeHours(in.Hours, false)
	case "WITHDRAW":
		if in.ExceptionID < 1 || in.EffectiveFrom != "" || in.Date != "" || in.Mode != "" || in.ClosureType != "" || len(in.Hours) > 0 {
			return in, identity.ErrValidation
		}
		in.Hours = []facility.Hours{}
	case "EXTEND":
		if in.ExceptionID != 0 || in.EffectiveFrom != "" || in.Date != "" || in.Mode != "" || in.ClosureType != "" || len(in.Hours) > 0 {
			return in, identity.ErrValidation
		}
		in.Hours = []facility.Hours{}
	default:
		return in, identity.ErrValidation
	}
	return in, err
}
func (s *CalendarService) Snapshot(ctx context.Context, a identity.User, id int64) (facility.CalendarSnapshot, error) {
	if !managesFacilities(a) {
		return facility.CalendarSnapshot{}, identity.ErrForbidden
	}
	return s.repo.Snapshot(ctx, id)
}
func (s *CalendarService) Preview(ctx context.Context, a identity.User, id int64, in facility.CalendarMutation) (facility.CalendarPreview, error) {
	if !managesFacilities(a) {
		return facility.CalendarPreview{}, identity.ErrForbidden
	}
	in, err := s.normalize(in)
	if err != nil {
		return facility.CalendarPreview{}, err
	}
	p, finger, err := s.repo.Preview(ctx, id, in)
	if err != nil {
		return p, err
	}
	p.ExpiresAt = s.now().Add(5 * time.Minute)
	b, _ := json.Marshal(previewClaims{a.ID, id, mutationHash(in), finger, p.ExpiresAt})
	mac := hmac.New(sha256.New, s.key[:])
	mac.Write(b)
	p.ID = base64.RawURLEncoding.EncodeToString(b) + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return p, nil
}
func (s *CalendarService) Confirm(ctx context.Context, a identity.User, id int64, in facility.CalendarConfirm, key, rid string) (facility.CalendarResult, error) {
	empty := facility.CalendarResult{}
	if !managesFacilities(a) {
		return empty, identity.ErrForbidden
	}
	if len(key) < 8 || len(key) > 100 {
		return empty, identity.ErrValidation
	}
	m, err := s.normalize(in.CalendarMutation, true)
	if err != nil {
		return empty, err
	}
	if result, found, err := s.repo.Replay(ctx, id, a.ID, m, key); found || err != nil {
		return result, err
	}
	m, err = s.normalize(in.CalendarMutation)
	if err != nil {
		return empty, err
	}
	parts := strings.Split(in.PreviewID, ".")
	if len(parts) != 2 {
		return empty, identity.Conflict("Preview không hợp lệ. Hãy xem trước lại.")
	}
	b, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return empty, identity.ErrValidation
	}
	sig, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return empty, identity.ErrValidation
	}
	mac := hmac.New(sha256.New, s.key[:])
	mac.Write(b)
	if !hmac.Equal(sig, mac.Sum(nil)) {
		return empty, identity.Conflict("Preview không hợp lệ hoặc API vừa khởi động lại.")
	}
	var claims previewClaims
	if err = json.Unmarshal(b, &claims); err != nil {
		return empty, identity.ErrValidation
	}
	if claims.Actor != a.ID || claims.Room != id || claims.Hash != mutationHash(m) {
		return empty, identity.Conflict("Preview không khớp người dùng, phòng hoặc nội dung xác nhận.")
	}
	return s.repo.Commit(ctx, id, a.ID, m, claims.Fingerprint, claims.Expires, key, rid)
}
func (s *CalendarService) Days(ctx context.Context, id int64, from, to string) ([]facility.CalendarDay, error) {
	a, err := time.Parse("2006-01-02", from)
	if err != nil {
		return nil, identity.ErrValidation
	}
	b, err := time.Parse("2006-01-02", to)
	if err != nil || b.Before(a) || b.Sub(a) > 55*24*time.Hour {
		return nil, identity.ErrValidation
	}
	snapshot, err := s.repo.Snapshot(ctx, id)
	if err != nil {
		return nil, err
	}
	days := []facility.CalendarDay{}
	for !a.After(b) {
		days = append(days, snapshot.Day(a.Format("2006-01-02")))
		a = a.AddDate(0, 0, 1)
	}
	return days, nil
}
