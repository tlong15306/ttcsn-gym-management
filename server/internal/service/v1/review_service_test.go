package v1service

import (
	"context"
	"strings"
	"testing"

	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/catalog"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/identity"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/review"
)

type reviewRepoStub struct {
	input review.Input
	page  catalog.Filter
}

func (s *reviewRepoStub) Create(_ context.Context, _ identity.User, _ int64, in review.Input, _ string) (review.Review, error) {
	s.input = in
	return review.Review{Rating: in.Rating}, nil
}
func (*reviewRepoStub) Get(context.Context, identity.User, int64) (*review.Review, error) {
	return nil, nil
}
func (s *reviewRepoStub) ListPT(_ context.Context, _ identity.User, _ int64, f catalog.Filter) ([]review.Review, int, error) {
	s.page = f
	return []review.Review{}, 0, nil
}

func TestReviewServiceValidatesAndNormalizesInput(t *testing.T) {
	repo := &reviewRepoStub{}
	svc := NewReviewService(repo)
	actor := identity.User{ID: 4, Role: identity.RoleMember}
	if _, err := svc.Create(context.Background(), actor, 9, review.Input{Rating: 4, Comment: "  good  "}, "rid"); err != nil {
		t.Fatal(err)
	}
	if repo.input.Comment != "good" {
		t.Fatalf("comment was not trimmed: %q", repo.input.Comment)
	}
	for _, in := range []review.Input{{Rating: 0}, {Rating: 6}, {Rating: 3, Comment: strings.Repeat("x", 1001)}} {
		if _, err := svc.Create(context.Background(), actor, 9, in, "rid"); err != identity.ErrValidation {
			t.Fatalf("invalid review should fail validation: %v", err)
		}
	}
	if _, err := svc.Create(context.Background(), identity.User{ID: 5, Role: identity.RolePT}, 9, review.Input{Rating: 5}, "rid"); err != identity.ErrForbidden {
		t.Fatalf("PT must not create review: %v", err)
	}
}

func TestReviewServiceValidatesPagination(t *testing.T) {
	repo := &reviewRepoStub{}
	svc := NewReviewService(repo)
	if _, _, err := svc.ListPT(context.Background(), identity.User{Role: identity.RoleMember}, 2, catalog.Filter{Page: 1, PageSize: 101}); err != identity.ErrValidation {
		t.Fatalf("oversized page should fail: %v", err)
	}
	_, _, err := svc.ListPT(context.Background(), identity.User{Role: identity.RoleMember}, 2, catalog.Filter{Page: 2, PageSize: 10})
	if err != nil || repo.page.Page != 2 || repo.page.PageSize != 10 {
		t.Fatalf("valid page should be forwarded: filter=%+v err=%v", repo.page, err)
	}
}
