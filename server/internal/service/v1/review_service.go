package v1service

import (
	"context"
	"strings"
	"unicode/utf8"

	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/catalog"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/identity"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/review"
)

type ReviewRepository interface {
	Create(context.Context, identity.User, int64, review.Input, string) (review.Review, error)
	Get(context.Context, identity.User, int64) (*review.Review, error)
	ListPT(context.Context, identity.User, int64, catalog.Filter) ([]review.Review, int, error)
}

type ReviewService struct{ repo ReviewRepository }

func NewReviewService(repo ReviewRepository) *ReviewService { return &ReviewService{repo: repo} }

func (s *ReviewService) Create(ctx context.Context, actor identity.User, bookingID int64, in review.Input, requestID string) (review.Review, error) {
	if actor.Role != identity.RoleMember {
		return review.Review{}, identity.ErrForbidden
	}
	if bookingID < 1 {
		return review.Review{}, identity.ErrValidation
	}
	in.Comment = strings.TrimSpace(in.Comment)
	if in.Rating < 1 || in.Rating > 5 || utf8.RuneCountInString(in.Comment) > 1000 {
		return review.Review{}, identity.ErrValidation
	}
	return s.repo.Create(ctx, actor, bookingID, in, requestID)
}

func (s *ReviewService) Get(ctx context.Context, actor identity.User, bookingID int64) (*review.Review, error) {
	if bookingID < 1 {
		return nil, identity.ErrValidation
	}
	return s.repo.Get(ctx, actor, bookingID)
}

func (s *ReviewService) ListPT(ctx context.Context, actor identity.User, ptID int64, f catalog.Filter) ([]review.Review, int, error) {
	if ptID < 1 {
		return nil, 0, identity.ErrValidation
	}
	f, err := normalizeFilter(f)
	if err != nil {
		return nil, 0, err
	}
	return s.repo.ListPT(ctx, actor, ptID, f)
}
