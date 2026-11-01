package v1service

import (
	"context"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/booking"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/catalog"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/identity"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/repository"
	"strings"
	"time"
)

type BookingService struct{ repo repository.BookingRepository }

const maxBatchBookingSessions = 24

func NewBookingService(repo repository.BookingRepository) *BookingService {
	return &BookingService{repo}
}
func canCreate(a identity.User) bool {
	return a.Role == identity.RoleMember || a.Role == identity.RoleAdmin || a.Role == identity.RoleReceptionist
}
func (s *BookingService) Create(ctx context.Context, a identity.User, member int64, in booking.CreateInput, key, rid string) (booking.Booking, error) {
	if member == 0 && a.Role != identity.RoleMember || member > 0 && (a.Role != identity.RoleAdmin && a.Role != identity.RoleReceptionist) {
		return booking.Booking{}, identity.ErrForbidden
	}
	if in.RoomSessionID < 1 || in.PTID < 0 || member < 0 || len(key) < 8 || len(key) > 100 || strings.TrimSpace(key) != key {
		return booking.Booking{}, identity.ErrValidation
	}
	return s.repo.Create(ctx, a, member, in, key, rid)
}
func (s *BookingService) CreateBatch(ctx context.Context, a identity.User, member int64, in booking.BatchCreateInput, key, rid string) ([]booking.Booking, error) {
	if member == 0 && a.Role != identity.RoleMember || member > 0 && (a.Role != identity.RoleAdmin && a.Role != identity.RoleReceptionist) {
		return nil, identity.ErrForbidden
	}
	if len(in.RoomSessionIDs) < 1 || len(in.RoomSessionIDs) > maxBatchBookingSessions || in.PTID < 0 || member < 0 || len(key) < 8 || len(key) > 100 || strings.TrimSpace(key) != key {
		return nil, identity.ErrValidation
	}
	seen := make(map[int64]struct{}, len(in.RoomSessionIDs))
	for _, sessionID := range in.RoomSessionIDs {
		if sessionID < 1 {
			return nil, identity.ErrValidation
		}
		if _, duplicate := seen[sessionID]; duplicate {
			return nil, identity.ErrValidation
		}
		seen[sessionID] = struct{}{}
	}
	return s.repo.CreateBatch(ctx, a, member, in, key, rid)
}
func (s *BookingService) Get(ctx context.Context, a identity.User, id int64) (booking.Booking, error) {
	return s.repo.Get(ctx, a, id)
}
func (s *BookingService) List(ctx context.Context, a identity.User, f booking.Filter) ([]booking.Booking, int, error) {
	var err error
	f.Filter, err = normalizeFilter(f.Filter)
	for _, date := range []string{f.From, f.To} {
		if date != "" {
			if _, e := time.Parse("2006-01-02", date); e != nil {
				return nil, 0, identity.ErrValidation
			}
		}
	}
	if err != nil || f.PTID < 0 || f.RoomID < 0 || !oneOf(f.Status, "", "PENDING_PT", "CONFIRMED", "REJECTED", "EXPIRED", "CANCELLED", "COMPLETED", "NO_SHOW") || (f.From != "" && f.To != "" && f.From > f.To) {
		return nil, 0, identity.ErrValidation
	}
	return s.repo.List(ctx, a, f)
}
func (s *BookingService) Act(ctx context.Context, a identity.User, id int64, in booking.ActionInput, cancel bool, rid string) (booking.Booking, error) {
	if !cancel && a.Role != identity.RolePT {
		return booking.Booking{}, identity.ErrForbidden
	}
	in.Reason = strings.TrimSpace(in.Reason)
	if len([]rune(in.Reason)) > 1000 || (!cancel && !oneOf(in.Action, "ACCEPT", "REJECT")) || (cancel && in.Action != "") {
		return booking.Booking{}, identity.ErrValidation
	}
	if cancel && a.Role == identity.RolePT {
		b, err := s.repo.Get(ctx, a, id)
		if err != nil {
			return b, err
		}
		if b.Status == "PENDING_PT" {
			return b, booking.Fail("BOOKING_INVALID_STATE", "Hãy từ chối yêu cầu đang chờ thay vì hủy lịch.")
		}
	}
	return s.repo.Act(ctx, a, id, in, cancel, rid)
}
func (s *BookingService) Sessions(ctx context.Context, a identity.User, f catalog.Filter) ([]booking.SessionOption, int, error) {
	if !canCreate(a) {
		return nil, 0, identity.ErrForbidden
	}
	_, err := time.Parse("2006-01-02", f.Date)
	if err != nil || f.RoomID < 1 {
		return nil, 0, identity.ErrValidation
	}
	f, err = normalizeFilter(f)
	if err != nil {
		return nil, 0, err
	}
	return s.repo.Sessions(ctx, a, f)
}
func (s *BookingService) PTs(ctx context.Context, a identity.User, id int64, f catalog.Filter) ([]booking.PTOption, int, error) {
	if !canCreate(a) {
		return nil, 0, identity.ErrForbidden
	}
	if id < 1 {
		return nil, 0, identity.ErrValidation
	}
	f, err := normalizeFilter(f)
	if err != nil {
		return nil, 0, err
	}
	return s.repo.PTs(ctx, id, f)
}
