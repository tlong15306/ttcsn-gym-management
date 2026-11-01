package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/identity"
)

type Claims struct {
	Role        identity.Role `json:"role"`
	AuthVersion int           `json:"auth_version"`
	jwt.RegisteredClaims
}

type TokenService struct {
	secret []byte
	ttl    time.Duration
	now    func() time.Time
}

func NewTokenService(secret string, ttl time.Duration) TokenService {
	return TokenService{secret: []byte(secret), ttl: ttl, now: time.Now}
}

func (s TokenService) Issue(user identity.User) (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("generate token id: %w", err)
	}
	now := s.now().UTC()
	claims := Claims{
		Role:        user.Role,
		AuthVersion: user.AuthVersion,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   fmt.Sprintf("%d", user.ID),
			ID:        hex.EncodeToString(bytes),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.ttl)),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.secret)
}

func (s TokenService) Parse(raw string) (Claims, error) {
	var claims Claims
	token, err := jwt.ParseWithClaims(raw, &claims, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing algorithm")
		}
		return s.secret, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}), jwt.WithExpirationRequired())
	if err != nil || !token.Valid || claims.Subject == "" || claims.ID == "" || !claims.Role.Valid() || claims.AuthVersion < 1 {
		return Claims{}, fmt.Errorf("invalid token")
	}
	return claims, nil
}
