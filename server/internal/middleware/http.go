package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/identity"
	"github.com/tlong15306/ttcsn-gym-management/server/pkg/auth"
)

const actorKey = "actor"

type UserLoader interface {
	FindByID(ctx context.Context, id int64) (identity.User, error)
}

type TokenParser interface {
	Parse(raw string) (auth.Claims, error)
}

type Authenticator struct {
	tokens TokenParser
	users  UserLoader
}

func NewAuthenticator(tokens TokenParser, users UserLoader) *Authenticator {
	return &Authenticator{tokens: tokens, users: users}
}

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if len(requestID) < 8 || len(requestID) > 100 {
			requestID = newRequestID()
		}
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if recovered := recover(); recovered != nil {
				slog.Error("panic recovered", "request_id", requestID(c), "route", c.FullPath())
				apiError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Đã xảy ra lỗi không mong muốn.")
			}
		}()
		c.Next()
	}
}

func AccessLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		started := time.Now()
		c.Next()
		slog.Info("http request", "request_id", requestID(c), "method", c.Request.Method, "route", c.FullPath(), "status", c.Writer.Status(), "latency_ms", time.Since(started).Milliseconds())
	}
}

func CORS(allowedOrigins []string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(allowedOrigins))
	for _, origin := range allowedOrigins {
		allowed[origin] = struct{}{}
	}
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if _, ok := allowed[origin]; ok {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
			c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Request-ID, Idempotency-Key, If-None-Match")
			c.Header("Access-Control-Allow-Methods", "GET, POST, PATCH, PUT, DELETE, OPTIONS")
		}
		if c.Request.Method == http.MethodOptions {
			c.Status(http.StatusNoContent)
			c.Abort()
			return
		}
		c.Next()
	}
}

func (a *Authenticator) Require() gin.HandlerFunc {
	return func(c *gin.Context) {
		const bearer = "Bearer "
		authorization := c.GetHeader("Authorization")
		if !strings.HasPrefix(authorization, bearer) {
			apiError(c, http.StatusUnauthorized, "UNAUTHENTICATED", "Phiên đăng nhập không hợp lệ hoặc đã hết hạn.")
			return
		}
		claims, err := a.tokens.Parse(strings.TrimPrefix(authorization, bearer))
		if err != nil {
			apiError(c, http.StatusUnauthorized, "UNAUTHENTICATED", "Phiên đăng nhập không hợp lệ hoặc đã hết hạn.")
			return
		}
		id, err := strconv.ParseInt(claims.Subject, 10, 64)
		if err != nil || id < 1 {
			apiError(c, http.StatusUnauthorized, "UNAUTHENTICATED", "Phiên đăng nhập không hợp lệ hoặc đã hết hạn.")
			return
		}
		user, err := a.users.FindByID(c.Request.Context(), id)
		if err != nil || user.Status != identity.StatusActive || user.AuthVersion != claims.AuthVersion || user.Role != claims.Role {
			apiError(c, http.StatusUnauthorized, "UNAUTHENTICATED", "Phiên đăng nhập không hợp lệ hoặc đã hết hạn.")
			return
		}
		c.Set(actorKey, user)
		c.Next()
	}
}

func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		actor, ok := Actor(c)
		if !ok || !actor.IsAdmin() {
			apiError(c, http.StatusForbidden, "FORBIDDEN", "Bạn không có quyền thực hiện thao tác này.")
			return
		}
		c.Next()
	}
}

func Actor(c *gin.Context) (identity.User, bool) {
	value, ok := c.Get(actorKey)
	if !ok {
		return identity.User{}, false
	}
	user, ok := value.(identity.User)
	return user, ok
}

func newRequestID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return fmt.Sprintf("req-%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(bytes)
}

func requestID(c *gin.Context) string {
	value, _ := c.Get("request_id")
	requestID, _ := value.(string)
	return requestID
}

func apiError(c *gin.Context, status int, code, message string) {
	c.AbortWithStatusJSON(status, gin.H{
		"error":      gin.H{"code": code, "message": message},
		"request_id": requestID(c),
	})
}
