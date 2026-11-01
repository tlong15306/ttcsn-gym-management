package v1handler

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/catalog"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/identity"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/middleware"
	v1service "github.com/tlong15306/ttcsn-gym-management/server/internal/service/v1"
	"io"
	"net/http"
	"strconv"
)

type ManagementHandler struct {
	members    *v1service.MemberService
	facilities *v1service.FacilityService
	trainers   *v1service.TrainerService
}

func NewManagementHandler(m *v1service.MemberService, f *v1service.FacilityService, t *v1service.TrainerService) *ManagementHandler {
	return &ManagementHandler{m, f, t}
}
func boundFilter(c *gin.Context) (catalog.Filter, error) {
	base, err := listFilter(c)
	if err != nil {
		return catalog.Filter{}, identity.ErrValidation
	}
	roomID := int64(0)
	if raw := c.Query("room_id"); raw != "" {
		roomID, err = strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return catalog.Filter{}, identity.ErrValidation
		}
	}
	return catalog.Filter{Page: base.Page, PageSize: base.PageSize, Query: base.Query, Status: string(base.Status), Date: c.Query("date"), RoomID: roomID, AvailableAt: c.Query("available_at")}, nil
}
func actorID(c *gin.Context) (identity.User, int64, error) {
	a, ok := middleware.Actor(c)
	if !ok {
		return a, 0, identity.ErrUnauthorized
	}
	if c.Param("id") == "" {
		return a, 0, nil
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id < 1 {
		return a, 0, identity.ErrValidation
	}
	return a, id, nil
}
func bindBody[T any](c *gin.Context) (T, error) {
	var value T
	decoder := json.NewDecoder(http.MaxBytesReader(c.Writer, c.Request.Body, 65536))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&value); err != nil {
		return value, identity.ErrValidation
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		return value, identity.ErrValidation
	}
	return value, nil
}
func respond(c *gin.Context, value any, err error, status int) {
	if err != nil {
		RespondDomainError(c, err)
		return
	}
	Success(c, status, value, nil)
}
func respondList(c *gin.Context, value any, total int, f catalog.Filter, err error) {
	if err != nil {
		RespondDomainError(c, err)
		return
	}
	Success(c, http.StatusOK, value, gin.H{"page": f.Page, "page_size": f.PageSize, "total": total})
}
