package v1service

import (
	"context"
	"strings"
	"time"

	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/catalog"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/discovery"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/identity"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/repository"
)

type DiscoveryService struct {
	repo repository.DiscoveryRepository
}

func NewDiscoveryService(repo repository.DiscoveryRepository) *DiscoveryService {
	return &DiscoveryService{repo}
}
func discoveryFilter(f catalog.Filter) (catalog.Filter, error) {
	f, err := normalizeFilter(f)
	if len([]rune(f.Query)) > 120 || len(f.Status) > 50 {
		return f, identity.ErrValidation
	}
	f.Status = strings.TrimSpace(f.Status)
	return f, err
}
func (s *DiscoveryService) Rooms(ctx context.Context, f catalog.Filter) ([]discovery.Room, int, error) {
	f, err := discoveryFilter(f)
	if err != nil {
		return nil, 0, err
	}
	return s.repo.Rooms(ctx, f)
}
func (s *DiscoveryService) Room(ctx context.Context, id int64) (discovery.Room, error) {
	if id < 1 {
		return discovery.Room{}, identity.ErrValidation
	}
	return s.repo.Room(ctx, id)
}
func (s *DiscoveryService) Trainers(ctx context.Context, f catalog.Filter) ([]discovery.Trainer, int, error) {
	f, err := discoveryFilter(f)
	if err != nil {
		return nil, 0, err
	}
	return s.repo.Trainers(ctx, f)
}
func (s *DiscoveryService) Sessions(ctx context.Context, id int64, from, to string, f catalog.Filter) ([]discovery.Session, int, error) {
	f, err := discoveryFilter(f)
	if err != nil || id < 1 {
		return nil, 0, identity.ErrValidation
	}
	start, e1 := time.Parse("2006-01-02", from)
	end, e2 := time.Parse("2006-01-02", to)
	if e1 != nil || e2 != nil || end.Before(start) || end.Sub(start) > 6*24*time.Hour {
		return nil, 0, identity.ErrValidation
	}
	return s.repo.Sessions(ctx, id, from, to, f)
}
func (s *DiscoveryService) Cover(ctx context.Context, id int64, size string) ([]byte, int64, error) {
	if id < 1 || (size != "card" && size != "detail") {
		return nil, 0, identity.ErrValidation
	}
	return s.repo.Cover(ctx, id, size == "card")
}
