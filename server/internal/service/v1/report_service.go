package v1service

import (
	"context"
	"time"

	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/identity"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/report"
)

type ReportRepository interface {
	Operations(context.Context, report.Range, int64) (report.Operations, error)
	Trainer(context.Context, identity.User, int64, report.Range) (report.Trainer, error)
}

type ReportService struct {
	repo ReportRepository
	now  func() time.Time
}

func NewReportService(repo ReportRepository) *ReportService {
	return &ReportService{repo: repo, now: time.Now}
}

func (s *ReportService) Operations(ctx context.Context, actor identity.User, input report.PeriodInput, roomID int64) (report.Operations, error) {
	if actor.Role != identity.RoleAdmin && actor.Role != identity.RoleManager {
		return report.Operations{}, identity.ErrForbidden
	}
	if roomID < 0 {
		return report.Operations{}, identity.ErrValidation
	}
	period, err := report.ResolveRange(input, s.now())
	if err != nil {
		return report.Operations{}, identity.ErrValidation
	}
	return s.repo.Operations(ctx, period, roomID)
}

func (s *ReportService) Trainer(ctx context.Context, actor identity.User, ptID int64, input report.PeriodInput) (report.Trainer, error) {
	if actor.Role != identity.RolePT && actor.Role != identity.RoleAdmin && actor.Role != identity.RoleManager {
		return report.Trainer{}, identity.ErrForbidden
	}
	if ptID < 1 {
		return report.Trainer{}, identity.ErrValidation
	}
	period, err := report.ResolveRange(input, s.now())
	if err != nil {
		return report.Trainer{}, identity.ErrValidation
	}
	return s.repo.Trainer(ctx, actor, ptID, period)
}
