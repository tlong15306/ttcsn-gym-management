package v1service

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/catalog"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/identity"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/trainer"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/repository"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const MaxCertificateBytes = 5 * 1024 * 1024

type TrainerService struct {
	repo      repository.TrainerRepository
	uploadDir string
}

func NewTrainerService(repo repository.TrainerRepository, uploadDir string) *TrainerService {
	return &TrainerService{repo, uploadDir}
}
func (s *TrainerService) List(ctx context.Context, a identity.User, f catalog.Filter) ([]trainer.Profile, int, error) {
	f, err := normalizeFilter(f)
	if err != nil || !oneOf(f.Status, "", "DRAFT", "PENDING", "APPROVED", "REJECTED", "SUSPENDED") {
		return nil, 0, identity.ErrValidation
	}
	if f.AvailableAt != "" {
		if _, err := time.Parse(time.RFC3339, f.AvailableAt); err != nil {
			return nil, 0, identity.ErrValidation
		}
	}
	profiles, total, err := s.repo.List(ctx, a, f)
	for i := range profiles {
		profiles[i] = publicTrainerProfile(a, profiles[i])
	}
	return profiles, total, err
}
func publicTrainerProfile(a identity.User, p trainer.Profile) trainer.Profile {
	if !managesFacilities(a) && p.UserID != a.ID {
		p.DecisionReason = ""
		p.DecidedAt = nil
	}
	return p
}
func (s *TrainerService) Get(ctx context.Context, a identity.User, id int64) (trainer.Profile, error) {
	p, err := s.repo.Get(ctx, id)
	if err != nil {
		return p, err
	}
	if !managesFacilities(a) && p.UserID != a.ID && (p.ApprovalStatus != "APPROVED" || p.User.Role != identity.RolePT || p.User.Status != identity.StatusActive) {
		return trainer.Profile{}, identity.ErrNotFound
	}
	return publicTrainerProfile(a, p), nil
}
func (s *TrainerService) Own(ctx context.Context, a identity.User) (trainer.Profile, error) {
	if a.Role != identity.RolePT {
		return trainer.Profile{}, identity.ErrForbidden
	}
	return s.repo.Own(ctx, a.ID)
}
func (s *TrainerService) SaveProfile(ctx context.Context, a identity.User, in trainer.ProfileInput, rid string) (trainer.Profile, error) {
	if a.Role != identity.RolePT {
		return trainer.Profile{}, identity.ErrForbidden
	}
	in.Specialty = strings.TrimSpace(in.Specialty)
	in.ExperienceText = strings.TrimSpace(in.ExperienceText)
	if in.Specialty == "" || len([]rune(in.Specialty)) > 200 || in.ExperienceText == "" || len(in.ExperienceText) > 10000 {
		return trainer.Profile{}, identity.ErrValidation
	}
	return s.repo.SaveProfile(ctx, a.ID, in, rid)
}
func (s *TrainerService) Submit(ctx context.Context, a identity.User, rid string) (trainer.Profile, error) {
	if a.Role != identity.RolePT {
		return trainer.Profile{}, identity.ErrForbidden
	}
	return s.repo.Submit(ctx, a.ID, rid)
}
func (s *TrainerService) Decision(ctx context.Context, a identity.User, id int64, in trainer.DecisionInput, rid string) (trainer.Profile, error) {
	if !managesFacilities(a) {
		return trainer.Profile{}, identity.ErrForbidden
	}
	if !oneOf(in.Action, "APPROVE", "REJECT", "SUSPEND") || strings.TrimSpace(in.Reason) == "" {
		return trainer.Profile{}, identity.ErrValidation
	}
	return s.repo.Decision(ctx, id, in, a.ID, rid)
}
func (s *TrainerService) scheduleProfile(ctx context.Context, a identity.User, id int64) (trainer.Profile, error) {
	var p trainer.Profile
	var err error
	if id == 0 {
		p, err = s.repo.Own(ctx, a.ID)
	} else {
		p, err = s.repo.Get(ctx, id)
	}
	if err != nil {
		return p, err
	}
	if !managesFacilities(a) && (a.Role != identity.RolePT || p.UserID != a.ID) {
		return trainer.Profile{}, identity.ErrForbidden
	}
	return p, nil
}
func (s *TrainerService) Hours(ctx context.Context, a identity.User, id int64) ([]trainer.WorkingHour, error) {
	p, err := s.scheduleProfile(ctx, a, id)
	if err != nil {
		return nil, err
	}
	return s.repo.Hours(ctx, p.ID)
}
func validHours(hours []trainer.WorkingHour) bool {
	if len(hours) > 50 {
		return false
	}
	for i, h := range hours {
		if h.DayOfWeek < 1 || h.DayOfWeek > 7 || !validTime(h.StartTime) || !(validTime(h.EndTime) || h.EndTime == "24:00") || h.StartTime >= h.EndTime {
			return false
		}
		for j := 0; j < i; j++ {
			other := hours[j]
			if h.DayOfWeek == other.DayOfWeek && h.StartTime < other.EndTime && other.StartTime < h.EndTime {
				return false
			}
		}
	}
	return true
}
func (s *TrainerService) SaveHours(ctx context.Context, a identity.User, id int64, in trainer.AvailabilityInput, rid string) ([]trainer.WorkingHour, error) {
	p, err := s.scheduleProfile(ctx, a, id)
	if err != nil {
		return nil, err
	}
	if !validHours(in.Hours) {
		return nil, identity.ErrValidation
	}
	return s.repo.SaveHours(ctx, p.ID, in.Hours, a.ID, rid)
}
func (s *TrainerService) Absences(ctx context.Context, a identity.User, id int64, f catalog.Filter) ([]trainer.Unavailability, int, error) {
	f, err := normalizeFilter(f)
	if err != nil {
		return nil, 0, err
	}
	p, err := s.scheduleProfile(ctx, a, id)
	if err != nil {
		return nil, 0, err
	}
	return s.repo.Absences(ctx, p.ID, f)
}
func (s *TrainerService) SaveAbsence(ctx context.Context, a identity.User, id int64, in trainer.Unavailability, rid string) (trainer.Unavailability, error) {
	p, err := s.scheduleProfile(ctx, a, id)
	if err != nil {
		return in, err
	}
	start, e1 := time.Parse(time.RFC3339, in.StartsAt)
	end, e2 := time.Parse(time.RFC3339, in.EndsAt)
	if e1 != nil || e2 != nil || !start.Before(end) || strings.TrimSpace(in.Reason) == "" {
		return in, identity.ErrValidation
	}
	return s.repo.SaveAbsence(ctx, p.ID, in, a.ID, rid)
}
func (s *TrainerService) DeleteAbsence(ctx context.Context, a identity.User, id, absenceID int64, rid string) error {
	p, err := s.scheduleProfile(ctx, a, id)
	if err != nil {
		return err
	}
	return s.repo.DeleteAbsence(ctx, p.ID, absenceID, a.ID, rid)
}
func certificateExtension(content []byte) (string, error) {
	mime := http.DetectContentType(content)
	switch mime {
	case "application/pdf":
		if !bytes.HasPrefix(content, []byte("%PDF-")) || !bytes.Contains(content[max(0, len(content)-1024):], []byte("%%EOF")) {
			return "", identity.ErrValidation
		}
		return ".pdf", nil
	case "image/png":
		if !validCertificateImage(content) {
			return "", identity.ErrValidation
		}
		return ".png", nil
	case "image/jpeg":
		if !validCertificateImage(content) {
			return "", identity.ErrValidation
		}
		return ".jpg", nil
	default:
		return "", identity.ErrValidation
	}
}

func validCertificateImage(content []byte) bool {
	cfg, _, err := image.DecodeConfig(bytes.NewReader(content))
	if err != nil || cfg.Width < 1 || cfg.Height < 1 || int64(cfg.Width)*int64(cfg.Height) > 20000000 {
		return false
	}
	_, _, err = image.Decode(bytes.NewReader(content))
	return err == nil
}
func (s *TrainerService) UploadCertificate(ctx context.Context, a identity.User, reader io.Reader, rid string) (trainer.Profile, error) {
	if a.Role != identity.RolePT {
		return trainer.Profile{}, identity.ErrForbidden
	}
	content, err := io.ReadAll(io.LimitReader(reader, MaxCertificateBytes+1))
	if err != nil {
		return trainer.Profile{}, err
	}
	if len(content) == 0 || len(content) > MaxCertificateBytes {
		return trainer.Profile{}, identity.ErrValidation
	}
	ext, err := certificateExtension(content)
	if err != nil {
		return trainer.Profile{}, err
	}
	if err = os.MkdirAll(s.uploadDir, 0700); err != nil {
		return trainer.Profile{}, err
	}
	var random [24]byte
	if _, err = rand.Read(random[:]); err != nil {
		return trainer.Profile{}, err
	}
	name := hex.EncodeToString(random[:]) + ext
	path := filepath.Join(s.uploadDir, name)
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return trainer.Profile{}, err
	}
	_, writeErr := file.Write(content)
	closeErr := file.Close()
	if writeErr != nil || closeErr != nil {
		os.Remove(path)
		return trainer.Profile{}, errors.Join(writeErr, closeErr)
	}
	profile, err := s.repo.SetCertificate(ctx, a.ID, name, rid)
	if err != nil {
		os.Remove(path)
	}
	return profile, err
}
func (s *TrainerService) Certificate(ctx context.Context, a identity.User, id int64) (string, error) {
	p, err := s.repo.Get(ctx, id)
	if err != nil {
		return "", err
	}
	if !managesFacilities(a) && (a.Role != identity.RolePT || p.UserID != a.ID) {
		return "", identity.ErrForbidden
	}
	name, err := s.repo.Certificate(ctx, id)
	if err != nil {
		return "", err
	}
	if name != filepath.Base(name) || strings.ContainsAny(name, "/\\") || !oneOf(filepath.Ext(name), ".pdf", ".jpg", ".png") {
		return "", fmt.Errorf("invalid certificate storage name")
	}
	path := filepath.Join(s.uploadDir, name)
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return "", identity.ErrNotFound
		}
		return "", err
	}
	return path, nil
}
