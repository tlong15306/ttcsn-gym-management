package v1service

import (
	"context"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/attendance"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/identity"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/repository"
	"strings"
)

type AttendanceService struct {
	repo repository.AttendanceRepository
}

func NewAttendanceService(r repository.AttendanceRepository) *AttendanceService {
	return &AttendanceService{r}
}
func (s *AttendanceService) Get(ctx context.Context, a identity.User, id int64) (attendance.Summary, error) {
	return s.repo.Get(ctx, a, id)
}
func (s *AttendanceService) Record(ctx context.Context, a identity.User, id int64, in attendance.Input, rid string) (attendance.Result, error) {
	if a.Role != identity.RoleAdmin && a.Role != identity.RoleManager && a.Role != identity.RoleReceptionist {
		return attendance.Result{}, identity.ErrForbidden
	}
	in.OverrideReason = strings.TrimSpace(in.OverrideReason)
	if !oneOf(in.ParticipantType, "MEMBER", "PT") || len([]rune(in.OverrideReason)) > 1000 {
		return attendance.Result{}, identity.ErrValidation
	}
	return s.repo.Record(ctx, a, id, in, rid)
}
