package v1service

import (
	"context"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/catalog"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/facility"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/identity"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/repository"
	"strings"
	"time"
)

type FacilityService struct{ repo repository.FacilityRepository }

func (s *FacilityService) Slot(ctx context.Context, id int64) (facility.TimeSlot, error) {
	return s.repo.Slot(ctx, id)
}
func (s *FacilityService) UpdateSlot(ctx context.Context, a identity.User, id int64, in facility.SlotInput, rid string) (facility.TimeSlot, error) {
	if !managesFacilities(a) {
		return facility.TimeSlot{}, identity.ErrForbidden
	}
	if strings.TrimSpace(in.Name) == "" || len(in.Name) > 80 || !validTime(in.StartTime) || !(validTime(in.EndTime) || in.EndTime == "24:00") || in.StartTime >= in.EndTime || strings.TrimSpace(in.Reason) == "" {
		return facility.TimeSlot{}, identity.ErrValidation
	}
	return s.repo.UpdateSlot(ctx, id, in, a.ID, rid)
}

func NewFacilityService(repo repository.FacilityRepository) *FacilityService {
	return &FacilityService{repo}
}
func managesFacilities(a identity.User) bool {
	return a.Role == identity.RoleAdmin || a.Role == identity.RoleManager
}
func oneOf(s string, values ...string) bool {
	for _, v := range values {
		if s == v {
			return true
		}
	}
	return false
}
func (s *FacilityService) Rooms(ctx context.Context, f catalog.Filter) ([]facility.Room, int, error) {
	f, err := normalizeFilter(f)
	if err != nil || !oneOf(f.Status, "", "AVAILABLE", "MAINTENANCE", "INACTIVE") {
		return nil, 0, identity.ErrValidation
	}
	return s.repo.Rooms(ctx, f)
}
func (s *FacilityService) Room(ctx context.Context, id int64) (facility.Room, error) {
	return s.repo.Room(ctx, id)
}
func (s *FacilityService) RoomImpact(ctx context.Context, a identity.User, id int64) (facility.Impact, error) {
	if !managesFacilities(a) {
		return facility.Impact{}, identity.ErrForbidden
	}
	return s.repo.RoomImpact(ctx, id)
}
func (s *FacilityService) SaveRoom(ctx context.Context, a identity.User, id int64, in facility.RoomInput, rid string) (facility.Room, error) {
	if !managesFacilities(a) {
		return facility.Room{}, identity.ErrForbidden
	}
	if in.Status == "" && id == 0 {
		in.Status = "AVAILABLE"
	}
	in.Name = strings.TrimSpace(in.Name)
	in.RoomType = strings.TrimSpace(in.RoomType)
	if in.Name == "" || len([]rune(in.Name)) > 100 || in.RoomType == "" || len(in.RoomType) > 50 || in.DefaultCapacity < 1 || in.DefaultCapacity > 10000 || len(in.Description) > 4000 || !oneOf(in.Status, "AVAILABLE", "MAINTENANCE", "INACTIVE") || (id > 0 && strings.TrimSpace(in.Reason) == "") {
		return facility.Room{}, identity.ErrValidation
	}
	return s.repo.SaveRoom(ctx, id, in, a.ID, rid)
}
func (s *FacilityService) Slots(ctx context.Context, f catalog.Filter) ([]facility.TimeSlot, int, error) {
	f, err := normalizeFilter(f)
	if err != nil || !oneOf(f.Status, "", "ACTIVE", "INACTIVE") {
		return nil, 0, identity.ErrValidation
	}
	return s.repo.Slots(ctx, f)
}
func validTime(s string) bool { _, err := time.Parse("15:04", s); return err == nil }
func (s *FacilityService) CreateSlot(ctx context.Context, a identity.User, in facility.SlotInput, rid string) (facility.TimeSlot, error) {
	if !managesFacilities(a) {
		return facility.TimeSlot{}, identity.ErrForbidden
	}
	in.Name = strings.TrimSpace(in.Name)
	if in.Name == "" || len(in.Name) > 80 || !validTime(in.StartTime) || !(validTime(in.EndTime) || in.EndTime == "24:00") || in.StartTime >= in.EndTime {
		return facility.TimeSlot{}, identity.ErrValidation
	}
	return s.repo.CreateSlot(ctx, in, a.ID, rid)
}
func (s *FacilityService) SlotStatus(ctx context.Context, a identity.User, id int64, in facility.SlotStatusInput, rid string) (facility.TimeSlot, error) {
	if !managesFacilities(a) {
		return facility.TimeSlot{}, identity.ErrForbidden
	}
	if strings.TrimSpace(in.Reason) == "" {
		return facility.TimeSlot{}, identity.ErrValidation
	}
	return s.repo.SlotStatus(ctx, id, in, a.ID, rid)
}
func (s *FacilityService) Sessions(ctx context.Context, f catalog.Filter) ([]facility.Session, int, error) {
	f, err := normalizeFilter(f)
	if err != nil || !oneOf(f.Status, "", "OPEN", "CLOSED", "MAINTENANCE") || (f.Date != "" && !dateValid(f.Date)) || f.RoomID < 0 {
		return nil, 0, identity.ErrValidation
	}
	return s.repo.Sessions(ctx, f)
}
func (s *FacilityService) Session(ctx context.Context, id int64) (facility.Session, error) {
	return s.repo.Session(ctx, id)
}
func (s *FacilityService) SaveSession(ctx context.Context, a identity.User, id int64, in facility.SessionInput, rid string) (facility.Session, error) {
	if !managesFacilities(a) {
		return facility.Session{}, identity.ErrForbidden
	}
	if id == 0 {
		if in.RoomID < 1 || in.TimeSlotID < 1 || !dateValid(in.SessionDate) || in.Capacity < 0 || in.Capacity > 10000 {
			return facility.Session{}, identity.ErrValidation
		}
	} else {
		if in.Capacity < 1 || in.Capacity > 10000 || !oneOf(in.Status, "OPEN", "CLOSED", "MAINTENANCE") || strings.TrimSpace(in.Reason) == "" {
			return facility.Session{}, identity.ErrValidation
		}
	}
	return s.repo.SaveSession(ctx, id, in, a.ID, rid)
}
