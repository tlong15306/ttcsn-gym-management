package v1handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/identity"
	v1dto "github.com/tlong15306/ttcsn-gym-management/server/internal/dto/v1"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/middleware"
	v1service "github.com/tlong15306/ttcsn-gym-management/server/internal/service/v1"
)

type IdentityHandler struct {
	service         *v1service.Service
	tokenTTLSeconds int64
}

func NewIdentityHandler(service *v1service.Service, tokenTTLSeconds int64) *IdentityHandler {
	return &IdentityHandler{service: service, tokenTTLSeconds: tokenTTLSeconds}
}

func (h *IdentityHandler) Login(c *gin.Context) {
	var input v1dto.LoginRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		APIError(c, http.StatusBadRequest, "INVALID_REQUEST", "Dữ liệu gửi lên không đúng định dạng.")
		return
	}
	result, err := h.service.Login(c.Request.Context(), input.Username, input.Password, secondsToDuration(h.tokenTTLSeconds))
	if err != nil {
		RespondDomainError(c, err)
		return
	}
	Success(c, http.StatusOK, result, nil)
}

func (h *IdentityHandler) Logout(c *gin.Context) {
	actor, ok := middleware.Actor(c)
	if !ok {
		APIError(c, http.StatusUnauthorized, "UNAUTHENTICATED", "Phiên đăng nhập không hợp lệ hoặc đã hết hạn.")
		return
	}
	if err := h.service.Logout(c.Request.Context(), actor, RequestID(c)); err != nil {
		RespondDomainError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *IdentityHandler) Me(c *gin.Context) {
	actor, ok := middleware.Actor(c)
	if !ok {
		APIError(c, http.StatusUnauthorized, "UNAUTHENTICATED", "Phiên đăng nhập không hợp lệ hoặc đã hết hạn.")
		return
	}
	Success(c, http.StatusOK, actor, nil)
}

func (h *IdentityHandler) UpdateMe(c *gin.Context) {
	actor, ok := middleware.Actor(c)
	if !ok {
		APIError(c, http.StatusUnauthorized, "UNAUTHENTICATED", "Phiên đăng nhập không hợp lệ hoặc đã hết hạn.")
		return
	}
	var input v1dto.UpdateProfileRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		APIError(c, http.StatusBadRequest, "INVALID_REQUEST", "Dữ liệu gửi lên không đúng định dạng.")
		return
	}
	user, err := h.service.UpdateOwnProfile(c.Request.Context(), actor, input.ToServiceInput(), RequestID(c))
	if err != nil {
		RespondDomainError(c, err)
		return
	}
	Success(c, http.StatusOK, user, nil)
}

func (h *IdentityHandler) ListUsers(c *gin.Context) {
	actor, ok := middleware.Actor(c)
	if !ok {
		APIError(c, http.StatusUnauthorized, "UNAUTHENTICATED", "Phiên đăng nhập không hợp lệ hoặc đã hết hạn.")
		return
	}
	filter, err := listFilter(c)
	if err != nil {
		RespondDomainError(c, identity.ErrValidation)
		return
	}
	users, total, err := h.service.ListUsers(c.Request.Context(), actor, filter)
	if err != nil {
		RespondDomainError(c, err)
		return
	}
	Success(c, http.StatusOK, users, gin.H{"page": filter.Page, "page_size": filter.PageSize, "total": total})
}

func (h *IdentityHandler) CreateUser(c *gin.Context) {
	actor, ok := middleware.Actor(c)
	if !ok {
		APIError(c, http.StatusUnauthorized, "UNAUTHENTICATED", "Phiên đăng nhập không hợp lệ hoặc đã hết hạn.")
		return
	}
	var input v1dto.CreateUserRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		APIError(c, http.StatusBadRequest, "INVALID_REQUEST", "Dữ liệu gửi lên không đúng định dạng.")
		return
	}
	user, err := h.service.CreateUser(c.Request.Context(), actor, input.ToServiceInput(), RequestID(c))
	if err != nil {
		RespondDomainError(c, err)
		return
	}
	Success(c, http.StatusCreated, user, nil)
}

func (h *IdentityHandler) UpdateStatus(c *gin.Context) {
	h.updateAccount(c, func(actor identity.User, targetID int64, requestID string) (identity.User, error) {
		var input v1dto.UpdateStatusRequest
		if err := c.ShouldBindJSON(&input); err != nil {
			return identity.User{}, identity.ErrValidation
		}
		return h.service.SetStatus(c.Request.Context(), actor, targetID, input.Status, input.Reason, requestID)
	})
}

func (h *IdentityHandler) GetAccount(c *gin.Context) {
	h.updateAccount(c, func(actor identity.User, id int64, requestID string) (identity.User, error) {
		return h.service.GetAccount(c.Request.Context(), actor, id)
	})
}

func (h *IdentityHandler) UpdateAccount(c *gin.Context) {
	h.updateAccount(c, func(actor identity.User, id int64, requestID string) (identity.User, error) {
		var input v1dto.UpdateProfileRequest
		if err := c.ShouldBindJSON(&input); err != nil {
			return identity.User{}, identity.ErrValidation
		}
		return h.service.UpdateAccount(c.Request.Context(), actor, id, input.ToServiceInput(), requestID)
	})
}

func (h *IdentityHandler) UpdateRole(c *gin.Context) {
	h.updateAccount(c, func(actor identity.User, targetID int64, requestID string) (identity.User, error) {
		var input v1dto.UpdateRoleRequest
		if err := c.ShouldBindJSON(&input); err != nil {
			return identity.User{}, identity.ErrValidation
		}
		return h.service.SetRole(c.Request.Context(), actor, targetID, input.Role, input.Reason, requestID)
	})
}

func (h *IdentityHandler) ResetPassword(c *gin.Context) {
	h.updateAccount(c, func(actor identity.User, targetID int64, requestID string) (identity.User, error) {
		var input v1dto.ResetPasswordRequest
		if err := c.ShouldBindJSON(&input); err != nil {
			return identity.User{}, identity.ErrValidation
		}
		return h.service.ResetPassword(c.Request.Context(), actor, targetID, input.Password, requestID)
	})
}

func (h *IdentityHandler) updateAccount(c *gin.Context, update func(identity.User, int64, string) (identity.User, error)) {
	actor, ok := middleware.Actor(c)
	if !ok {
		APIError(c, http.StatusUnauthorized, "UNAUTHENTICATED", "Phiên đăng nhập không hợp lệ hoặc đã hết hạn.")
		return
	}
	targetID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || targetID < 1 {
		APIError(c, http.StatusBadRequest, "INVALID_REQUEST", "ID tài khoản không hợp lệ.")
		return
	}
	user, err := update(actor, targetID, RequestID(c))
	if err != nil {
		RespondDomainError(c, err)
		return
	}
	Success(c, http.StatusOK, user, nil)
}

func listFilter(c *gin.Context) (identity.ListFilter, error) {
	page, err := parsePositive(c.Query("page"), 1)
	if err != nil {
		return identity.ListFilter{}, err
	}
	pageSize, err := parsePositive(c.Query("page_size"), 20)
	if err != nil {
		return identity.ListFilter{}, err
	}
	return identity.ListFilter{Query: c.Query("q"), Role: identity.Role(c.Query("role")), Status: identity.Status(c.Query("status")), Page: page, PageSize: pageSize}, nil
}

func parsePositive(value string, fallback int) (int, error) {
	if value == "" {
		return fallback, nil
	}
	n, err := strconv.Atoi(value)
	if err != nil || n < 1 || n > 1000000 {
		return 0, identity.ErrValidation
	}
	return n, nil
}
