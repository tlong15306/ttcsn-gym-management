package v1handler

import (
	"errors"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/attendance"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/booking"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/identity"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/review"
)

type errorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func Success(c *gin.Context, status int, data any, meta any) {
	payload := gin.H{"data": data, "request_id": RequestID(c)}
	if meta != nil {
		payload["meta"] = meta
	}
	c.JSON(status, payload)
}

func APIError(c *gin.Context, status int, code, message string) {
	c.AbortWithStatusJSON(status, gin.H{
		"error":      errorBody{Code: code, Message: message},
		"request_id": RequestID(c),
	})
}

func RequestID(c *gin.Context) string {
	if requestID, ok := c.Get("request_id"); ok {
		if value, ok := requestID.(string); ok {
			return value
		}
	}
	return ""
}

func RespondDomainError(c *gin.Context, err error) {
	var attendanceError *attendance.Error
	if errors.As(err, &attendanceError) {
		APIError(c, 422, attendanceError.Code, attendanceError.Message)
		return
	}
	var bookingError *booking.Error
	if errors.As(err, &bookingError) {
		if bookingError.Code == "BATCH_SESSION_INVALID" || bookingError.Code == "BATCH_SESSIONS_NOT_CONSECUTIVE" {
			APIError(c, http.StatusUnprocessableEntity, bookingError.Code, bookingError.Message)
			return
		}
		APIError(c, 409, bookingError.Code, bookingError.Message)
		return
	}
	var reviewError *review.Error
	if errors.As(err, &reviewError) {
		status := http.StatusConflict
		if reviewError.Code == "REVIEW_NOT_ELIGIBLE" {
			status = http.StatusUnprocessableEntity
		}
		APIError(c, status, reviewError.Code, reviewError.Message)
		return
	}
	var public *identity.PublicError
	if errors.As(err, &public) {
		APIError(c, http.StatusConflict, "INVALID_STATE", public.Message)
		return
	}
	switch {
	case errors.Is(err, identity.ErrInvalidCredentials):
		APIError(c, http.StatusUnauthorized, "UNAUTHENTICATED", "Tên đăng nhập hoặc mật khẩu không đúng.")
	case errors.Is(err, identity.ErrUnauthorized):
		APIError(c, http.StatusUnauthorized, "UNAUTHENTICATED", "Phiên đăng nhập không hợp lệ hoặc đã hết hạn.")
	case errors.Is(err, identity.ErrForbidden):
		APIError(c, http.StatusForbidden, "FORBIDDEN", "Bạn không có quyền thực hiện thao tác này.")
	case errors.Is(err, identity.ErrNotFound):
		APIError(c, http.StatusNotFound, "NOT_FOUND", "Không tìm thấy dữ liệu yêu cầu.")
	case errors.Is(err, identity.ErrConflict):
		APIError(c, http.StatusConflict, "INVALID_STATE", "Dữ liệu đã tồn tại hoặc trạng thái không cho phép thao tác.")
	case errors.Is(err, identity.ErrLastActiveAdmin):
		APIError(c, http.StatusConflict, "INVALID_STATE", "Không thể khóa hoặc hạ quyền Quản trị viên đang hoạt động cuối cùng.")
	case errors.Is(err, identity.ErrProfileRequirement):
		APIError(c, http.StatusConflict, "INVALID_STATE", "Tài khoản cần có hồ sơ nghiệp vụ trước khi gán vai trò này.")
	case errors.Is(err, identity.ErrValidation):
		APIError(c, http.StatusUnprocessableEntity, "VALIDATION_FAILED", "Dữ liệu không hợp lệ.")
	default:
		APIError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Đã xảy ra lỗi không mong muốn.")
	}
}
