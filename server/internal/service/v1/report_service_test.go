package v1service

import (
	"context"
	"testing"
	"time"

	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/identity"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/report"
)

type reportRepoStub struct {
	operations int
	trainer    int
	rangeUsed  report.Range
	roomID     int64
	ptID       int64
}

func (r *reportRepoStub) Operations(_ context.Context, period report.Range, roomID int64) (report.Operations, error) {
	r.operations++
	r.rangeUsed, r.roomID = period, roomID
	return report.Operations{}, nil
}

func (r *reportRepoStub) Trainer(_ context.Context, _ identity.User, ptID int64, period report.Range) (report.Trainer, error) {
	r.trainer++
	r.rangeUsed, r.ptID = period, ptID
	return report.Trainer{}, nil
}

func newReportServiceForTest(repo *reportRepoStub) *ReportService {
	svc := NewReportService(repo)
	svc.now = func() time.Time { return time.Date(2026, 10, 14, 5, 0, 0, 0, time.UTC) }
	return svc
}

func TestReportServiceScopesOperationsAndValidatesRange(t *testing.T) {
	repo := &reportRepoStub{}
	svc := newReportServiceForTest(repo)
	manager := identity.User{Role: identity.RoleManager}
	if _, err := svc.Operations(context.Background(), manager, report.PeriodInput{}, 7); err != nil {
		t.Fatal(err)
	}
	if repo.operations != 1 || repo.roomID != 7 || repo.rangeUsed.From.Format("2006-01-02") != "2026-10-01" {
		t.Fatalf("unexpected repository input: %+v", repo)
	}
	for _, actor := range []identity.User{{Role: identity.RoleMember}, {Role: identity.RolePT}, {Role: identity.RoleReceptionist}} {
		if _, err := svc.Operations(context.Background(), actor, report.PeriodInput{}, 0); err != identity.ErrForbidden {
			t.Errorf("role %s should not read operations report: %v", actor.Role, err)
		}
	}
	if _, err := svc.Operations(context.Background(), manager, report.PeriodInput{From: "2026-10-15", To: "2026-10-14"}, 0); err != identity.ErrValidation {
		t.Fatalf("invalid period should fail validation: %v", err)
	}
	if _, err := svc.Operations(context.Background(), manager, report.PeriodInput{}, -1); err != identity.ErrValidation {
		t.Fatalf("negative room filter should fail validation: %v", err)
	}
}

func TestReportServiceScopesTrainerMetrics(t *testing.T) {
	repo := &reportRepoStub{}
	svc := newReportServiceForTest(repo)
	if _, err := svc.Trainer(context.Background(), identity.User{ID: 8, Role: identity.RolePT}, 12, report.PeriodInput{}); err != nil {
		t.Fatal(err)
	}
	if repo.trainer != 1 || repo.ptID != 12 || repo.rangeUsed.To.Format("2006-01-02") != "2026-10-14" {
		t.Fatalf("unexpected repository input: %+v", repo)
	}
	for _, actor := range []identity.User{{Role: identity.RoleMember}, {Role: identity.RoleReceptionist}} {
		if _, err := svc.Trainer(context.Background(), actor, 12, report.PeriodInput{}); err != identity.ErrForbidden {
			t.Errorf("role %s should not read trainer report: %v", actor.Role, err)
		}
	}
	if _, err := svc.Trainer(context.Background(), identity.User{Role: identity.RoleAdmin}, 0, report.PeriodInput{}); err != identity.ErrValidation {
		t.Fatalf("invalid PT id should fail validation: %v", err)
	}
}
