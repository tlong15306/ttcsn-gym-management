package v1service

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/identity"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/repository"
)

type PasswordHasher interface {
	Hash(password string) (string, error)
	Compare(hash, password string) error
}

type TokenIssuer interface {
	Issue(user identity.User) (string, error)
}

type Service struct {
	repository repository.UserRepository
	hasher     PasswordHasher
	tokens     TokenIssuer
}

func NewService(userRepository repository.UserRepository, hasher PasswordHasher, tokens TokenIssuer) *Service {
	return &Service{repository: userRepository, hasher: hasher, tokens: tokens}
}

type LoginResult struct {
	AccessToken string        `json:"access_token"`
	TokenType   string        `json:"token_type"`
	ExpiresIn   int64         `json:"expires_in"`
	User        identity.User `json:"user"`
}

func (s *Service) Login(ctx context.Context, username, password string, ttl time.Duration) (LoginResult, error) {
	user, err := s.repository.FindByUsername(ctx, identity.NormalizeIdentifier(username))
	if err != nil || user.Status != identity.StatusActive || s.hasher.Compare(user.PasswordHash, password) != nil {
		return LoginResult{}, identity.ErrInvalidCredentials
	}

	token, err := s.tokens.Issue(user)
	if err != nil {
		return LoginResult{}, fmt.Errorf("issue token: %w", err)
	}

	return LoginResult{AccessToken: token, TokenType: "Bearer", ExpiresIn: int64(ttl.Seconds()), User: user}, nil
}

func (s *Service) UpdateProfile(ctx context.Context, userID int64, input identity.ProfileInput) (identity.User, error) {
	if !validContact(input.FullName, input.Email, input.Phone) {
		return identity.User{}, identity.ErrValidation
	}
	input.FullName = strings.TrimSpace(input.FullName)
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))
	input.Phone = strings.TrimSpace(input.Phone)
	return s.repository.UpdateProfile(ctx, userID, input)
}

func (s *Service) UpdateOwnProfile(ctx context.Context, actor identity.User, input identity.ProfileInput, rid string) (identity.User, error) {
	if !validContact(input.FullName, input.Email, input.Phone) {
		return identity.User{}, identity.ErrValidation
	}
	input.FullName = strings.TrimSpace(input.FullName)
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))
	input.Phone = strings.TrimSpace(input.Phone)
	return s.repository.UpdateAccount(ctx, actor.ID, input, actor.ID, rid)
}

func validContact(name, email, phone string) bool {
	address, err := mail.ParseAddress(strings.TrimSpace(email))
	return err == nil && address.Address == strings.TrimSpace(email) && len(strings.TrimSpace(name)) > 0 && len([]rune(name)) <= 150 && len(email) <= 254 && len(phone) <= 20
}

func (s *Service) GetAccount(ctx context.Context, actor identity.User, id int64) (identity.User, error) {
	if !actor.IsAdmin() {
		return identity.User{}, identity.ErrForbidden
	}
	return s.repository.FindByID(ctx, id)
}

func (s *Service) UpdateAccount(ctx context.Context, actor identity.User, id int64, input identity.ProfileInput, requestID string) (identity.User, error) {
	if !actor.IsAdmin() {
		return identity.User{}, identity.ErrForbidden
	}
	if !validContact(input.FullName, input.Email, input.Phone) {
		return identity.User{}, identity.ErrValidation
	}
	input.FullName = strings.TrimSpace(input.FullName)
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))
	input.Phone = strings.TrimSpace(input.Phone)
	return s.repository.UpdateAccount(ctx, id, input, actor.ID, requestID)
}

func (s *Service) CreateUser(ctx context.Context, actor identity.User, input identity.CreateInput, requestID string) (identity.User, error) {
	if !actor.IsAdmin() {
		return identity.User{}, identity.ErrForbidden
	}
	if err := validateCreate(input); err != nil {
		return identity.User{}, err
	}

	hash, err := s.hasher.Hash(input.Password)
	if err != nil {
		return identity.User{}, fmt.Errorf("hash password: %w", err)
	}
	input.Username = identity.NormalizeIdentifier(input.Username)
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))
	input.Phone = strings.TrimSpace(input.Phone)
	input.FullName = strings.TrimSpace(input.FullName)
	return s.repository.Create(ctx, input, hash, actor.ID, requestID)
}

func validateCreate(input identity.CreateInput) error {
	if len(identity.NormalizeIdentifier(input.Username)) < 3 || len(identity.NormalizeIdentifier(input.Username)) > 50 {
		return identity.ErrValidation
	}
	if !validPassword(input.Password) {
		return identity.ErrValidation
	}
	if !validContact(input.FullName, input.Email, input.Phone) || !input.Role.Valid() {
		return identity.ErrValidation
	}
	return nil
}

func (s *Service) SetStatus(ctx context.Context, actor identity.User, targetID int64, status identity.Status, reason, requestID string) (identity.User, error) {
	if !actor.IsAdmin() {
		return identity.User{}, identity.ErrForbidden
	}
	if !status.Valid() || strings.TrimSpace(reason) == "" {
		return identity.User{}, identity.ErrValidation
	}
	return s.repository.SetStatus(ctx, targetID, status, strings.TrimSpace(reason), actor.ID, requestID)
}

func (s *Service) SetRole(ctx context.Context, actor identity.User, targetID int64, role identity.Role, reason, requestID string) (identity.User, error) {
	if !actor.IsAdmin() {
		return identity.User{}, identity.ErrForbidden
	}
	if !role.Valid() || strings.TrimSpace(reason) == "" {
		return identity.User{}, identity.ErrValidation
	}
	return s.repository.SetRole(ctx, targetID, role, strings.TrimSpace(reason), actor.ID, requestID)
}

func (s *Service) ResetPassword(ctx context.Context, actor identity.User, targetID int64, password, requestID string) (identity.User, error) {
	if !actor.IsAdmin() {
		return identity.User{}, identity.ErrForbidden
	}
	if !validPassword(password) {
		return identity.User{}, identity.ErrValidation
	}
	hash, err := s.hasher.Hash(password)
	if err != nil {
		return identity.User{}, fmt.Errorf("hash password: %w", err)
	}
	return s.repository.ResetPassword(ctx, targetID, hash, actor.ID, requestID)
}

func (s *Service) Logout(ctx context.Context, actor identity.User, requestID string) error {
	if actor.ID == 0 {
		return identity.ErrUnauthorized
	}
	return s.repository.BumpAuthVersion(ctx, actor.ID, actor.ID, requestID)
}

func validPassword(password string) bool {
	return utf8.RuneCountInString(password) >= 8 && len(password) <= 72
}

func (s *Service) GetUser(ctx context.Context, id int64) (identity.User, error) {
	return s.repository.FindByID(ctx, id)
}

func (s *Service) ListUsers(ctx context.Context, actor identity.User, filter identity.ListFilter) ([]identity.User, int, error) {
	if !actor.IsAdmin() {
		return nil, 0, identity.ErrForbidden
	}
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 {
		filter.PageSize = 20
	}
	if filter.PageSize > 100 || (filter.Role != "" && !filter.Role.Valid()) || (filter.Status != "" && !filter.Status.Valid()) {
		return nil, 0, identity.ErrValidation
	}
	filter.Query = strings.TrimSpace(filter.Query)
	return s.repository.List(ctx, filter)
}

func Is(err, target error) bool {
	return errors.Is(err, target)
}
