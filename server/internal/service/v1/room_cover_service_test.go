package v1service

import (
	"bytes"
	"context"
	"errors"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/facility"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/identity"
	"image"
	"image/png"
	"testing"
)

type coverStub struct {
	calls         int
	detail, thumb []byte
}

func (s *coverStub) Set(_ context.Context, _ int64, detail, thumb []byte, _ string, _ bool, _ int64, _ string) (facility.Room, error) {
	s.calls++
	s.detail = detail
	s.thumb = thumb
	return facility.Room{}, nil
}
func (s *coverStub) Read(context.Context, int64, bool) ([]byte, int64, error) { return nil, 0, nil }
func TestRoomCoverValidationAndBounds(t *testing.T) {
	repo := &coverStub{}
	service := NewRoomCoverService(repo)
	ctx := context.Background()
	admin := identity.User{ID: 1, Role: identity.RoleAdmin}
	var raw bytes.Buffer
	png.Encode(&raw, image.NewRGBA(image.Rect(0, 0, 2000, 1000)))
	if _, err := service.Upload(ctx, admin, 1, bytes.NewReader(raw.Bytes()), "Ảnh phòng", false, "test"); err != nil {
		t.Fatal(err)
	}
	for _, out := range []struct {
		raw   []byte
		width int
	}{{repo.detail, 1600}, {repo.thumb, 480}} {
		cfg, format, err := image.DecodeConfig(bytes.NewReader(out.raw))
		if err != nil || format != "jpeg" || cfg.Width != out.width {
			t.Fatal("image not bounded", cfg, err)
		}
	}
	for _, input := range [][]byte{[]byte("fake jpeg"), bytes.Repeat([]byte("x"), 5*1024*1024+1)} {
		if _, err := service.Upload(ctx, admin, 1, bytes.NewReader(input), "Room", false, "test"); !errors.Is(err, identity.ErrValidation) {
			t.Fatal("invalid accepted", err)
		}
	}
	for _, role := range []identity.Role{identity.RoleMember, identity.RolePT, identity.RoleReceptionist} {
		if _, err := service.Upload(ctx, identity.User{Role: role}, 1, bytes.NewReader(raw.Bytes()), "Room", false, "test"); !errors.Is(err, identity.ErrForbidden) {
			t.Fatal("role upload permitted", role)
		}
	}
	if repo.calls != 1 {
		t.Fatal("invalid upload wrote DB")
	}
}
