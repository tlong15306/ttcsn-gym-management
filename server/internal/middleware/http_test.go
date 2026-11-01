package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/identity"
	"github.com/tlong15306/ttcsn-gym-management/server/pkg/auth"
)

type fakeTokens struct {
	claims auth.Claims
	err    error
}

func (f fakeTokens) Parse(string) (auth.Claims, error) { return f.claims, f.err }

type fakeUsers struct {
	user identity.User
	err  error
}

func (f fakeUsers) FindByID(_ context.Context, _ int64) (identity.User, error) { return f.user, f.err }

func TestRequireAdminRejectsMember(t *testing.T) {
	gin.SetMode(gin.TestMode)
	member := identity.User{ID: 1, Role: identity.RoleMember, Status: identity.StatusActive, AuthVersion: 1}
	auth := NewAuthenticator(fakeTokens{claims: claimsFor(member)}, fakeUsers{user: member})
	r := gin.New()
	r.Use(RequestID())
	r.GET("/users", auth.Require(), RequireAdmin(), func(c *gin.Context) { c.Status(http.StatusOK) })

	request := httptest.NewRequest(http.MethodGet, "/users", nil)
	request.Header.Set("Authorization", "Bearer valid")
	response := httptest.NewRecorder()
	r.ServeHTTP(response, request)
	if response.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusForbidden)
	}
	if response.Header().Get("X-Request-ID") == "" {
		t.Fatal("missing request id")
	}
}

func TestRequireRejectsStaleToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	user := identity.User{ID: 1, Role: identity.RoleAdmin, Status: identity.StatusActive, AuthVersion: 2}
	stale := user
	stale.AuthVersion = 1
	auth := NewAuthenticator(fakeTokens{claims: claimsFor(stale)}, fakeUsers{user: user})
	r := gin.New()
	r.Use(RequestID())
	r.GET("/me", auth.Require(), func(c *gin.Context) { c.Status(http.StatusOK) })

	request := httptest.NewRequest(http.MethodGet, "/me", nil)
	request.Header.Set("Authorization", "Bearer stale")
	response := httptest.NewRecorder()
	r.ServeHTTP(response, request)
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusUnauthorized)
	}
}

func TestCORSOnlyReturnsConfiguredOrigin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	const allowed = "http://127.0.0.1:5173"
	r := gin.New()
	r.Use(CORS([]string{allowed}))
	r.Any("/rooms", func(c *gin.Context) { c.Status(http.StatusOK) })

	preflight := httptest.NewRequest(http.MethodOptions, "/rooms", nil)
	preflight.Header.Set("Origin", allowed)
	preflightResponse := httptest.NewRecorder()
	r.ServeHTTP(preflightResponse, preflight)
	if preflightResponse.Code != http.StatusNoContent || preflightResponse.Header().Get("Access-Control-Allow-Origin") != allowed {
		t.Fatalf("allowed preflight status=%d origin=%q", preflightResponse.Code, preflightResponse.Header().Get("Access-Control-Allow-Origin"))
	}
	if preflightResponse.Header().Get("Vary") != "Origin" {
		t.Fatal("allowed preflight must vary by Origin")
	}

	denied := httptest.NewRequest(http.MethodOptions, "/rooms", nil)
	denied.Header.Set("Origin", "https://untrusted.example")
	deniedResponse := httptest.NewRecorder()
	r.ServeHTTP(deniedResponse, denied)
	if deniedResponse.Code != http.StatusNoContent || deniedResponse.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Fatalf("untrusted preflight status=%d origin=%q", deniedResponse.Code, deniedResponse.Header().Get("Access-Control-Allow-Origin"))
	}
}

func claimsFor(user identity.User) auth.Claims {
	return auth.Claims{Role: user.Role, AuthVersion: user.AuthVersion, RegisteredClaims: jwt.RegisteredClaims{Subject: "1", ID: "token"}}
}
