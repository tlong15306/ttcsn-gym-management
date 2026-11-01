package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/identity"
)

func TestTokenServiceIssueAndParse(t *testing.T) {
	service := NewTokenService("01234567890123456789012345678901", time.Minute)
	token, err := service.Issue(identity.User{ID: 8, Role: identity.RoleAdmin, AuthVersion: 3})
	if err != nil {
		t.Fatal(err)
	}
	claims, err := service.Parse(token)
	if err != nil {
		t.Fatal(err)
	}
	if claims.Subject != "8" || claims.Role != identity.RoleAdmin || claims.AuthVersion != 3 {
		t.Fatalf("claims = %#v", claims)
	}
}

func TestTokenServiceRejectsWrongSecret(t *testing.T) {
	issuer := NewTokenService("01234567890123456789012345678901", time.Minute)
	token, err := issuer.Issue(identity.User{ID: 8, Role: identity.RoleAdmin, AuthVersion: 1})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := NewTokenService("abcdefghijabcdefghijabcdefghij12", time.Minute).Parse(token); err == nil {
		t.Fatal("Parse() error = nil, want signature failure")
	}
}

func TestTokenServiceRejectsWrongAlgorithmAndMissingClaims(t *testing.T) {
	secret := "01234567890123456789012345678901"
	service := NewTokenService(secret, time.Minute)
	wrongAlgorithm := jwt.NewWithClaims(jwt.SigningMethodHS384, Claims{Role: identity.RoleAdmin, AuthVersion: 1, RegisteredClaims: jwt.RegisteredClaims{Subject: "1", ID: "id", ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute))}})
	wrongAlgorithmToken, err := wrongAlgorithm.SignedString([]byte(secret))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.Parse(wrongAlgorithmToken); err == nil {
		t.Fatal("Parse() error = nil, want algorithm rejection")
	}

	missingClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{Subject: "1", ID: "id", ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute))})
	missingClaimsToken, err := missingClaims.SignedString([]byte(secret))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.Parse(missingClaimsToken); err == nil {
		t.Fatal("Parse() error = nil, want missing claims rejection")
	}
}

func TestTokenServiceRejectsExpiredToken(t *testing.T) {
	service := NewTokenService("01234567890123456789012345678901", time.Minute)
	service.now = func() time.Time { return time.Unix(0, 0) }
	token, err := service.Issue(identity.User{ID: 8, Role: identity.RoleAdmin, AuthVersion: 1})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.Parse(token); err == nil {
		t.Fatal("Parse() error = nil, want expired token rejection")
	}
}

func TestTokenServiceRejectsMissingExpiration(t *testing.T) {
	secret := "01234567890123456789012345678901"
	claims := Claims{
		Role:        identity.RoleAdmin,
		AuthVersion: 1,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: "1",
			ID:      "test-without-expiration",
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := NewTokenService(secret, time.Minute).Parse(token); err == nil {
		t.Fatal("Parse() error = nil, want missing expiration rejection")
	}
}
