package v1service

import (
	"bytes"
	"context"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/facility"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/identity"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/repository"
	"image"
	"image/color"
	"image/jpeg"
	_ "image/png"
	"io"
	"strings"
)

type RoomCoverService struct {
	repo repository.RoomCoverRepository
}

func NewRoomCoverService(repo repository.RoomCoverRepository) *RoomCoverService {
	return &RoomCoverService{repo}
}
func resizeCover(src image.Image, max int) []byte {
	b := src.Bounds()
	w, h := b.Dx(), b.Dy()
	if w > max || h > max {
		if w >= h {
			h = h * max / w
			w = max
		} else {
			w = w * max / h
			h = max
		}
	}
	if w < 1 {
		w = 1
	}
	if h < 1 {
		h = 1
	}
	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			r, g, blue, a := src.At(b.Min.X+x*b.Dx()/w, b.Min.Y+y*b.Dy()/h).RGBA()
			dst.SetRGBA(x, y, color.RGBA{uint8((r + 65535 - a) / 257), uint8((g + 65535 - a) / 257), uint8((blue + 65535 - a) / 257), 255})
		}
	}
	var out bytes.Buffer
	_ = jpeg.Encode(&out, dst, &jpeg.Options{Quality: 88})
	return out.Bytes()
}
func (s *RoomCoverService) Upload(ctx context.Context, a identity.User, id int64, input io.Reader, alt string, demo bool, rid string) (facility.Room, error) {
	if !managesFacilities(a) {
		return facility.Room{}, identity.ErrForbidden
	}
	alt = strings.TrimSpace(alt)
	if alt == "" || len([]rune(alt)) > 300 {
		return facility.Room{}, identity.ErrValidation
	}
	raw, err := io.ReadAll(io.LimitReader(input, 5*1024*1024+1))
	if err != nil || len(raw) > 5*1024*1024 {
		return facility.Room{}, identity.ErrValidation
	}
	cfg, format, err := image.DecodeConfig(bytes.NewReader(raw))
	if err != nil || (format != "jpeg" && format != "png") || cfg.Width < 1 || cfg.Height < 1 || cfg.Width > 4096 || cfg.Height > 4096 || int64(cfg.Width)*int64(cfg.Height) > 16_000_000 {
		return facility.Room{}, identity.ErrValidation
	}
	img, _, err := image.Decode(bytes.NewReader(raw))
	if err != nil {
		return facility.Room{}, identity.ErrValidation
	}
	return s.repo.Set(ctx, id, resizeCover(img, 1600), resizeCover(img, 480), alt, demo, a.ID, rid)
}
func (s *RoomCoverService) Delete(ctx context.Context, a identity.User, id int64, rid string) (facility.Room, error) {
	if !managesFacilities(a) {
		return facility.Room{}, identity.ErrForbidden
	}
	return s.repo.Set(ctx, id, nil, nil, "", false, a.ID, rid)
}
func (s *RoomCoverService) Read(ctx context.Context, id int64, card bool) ([]byte, int64, error) {
	return s.repo.Read(ctx, id, card)
}
