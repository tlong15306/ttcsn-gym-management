package v1handler

import (
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/catalog"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/identity"
	v1service "github.com/tlong15306/ttcsn-gym-management/server/internal/service/v1"
)

type DiscoveryHandler struct{ s *v1service.DiscoveryService }

func NewDiscoveryHandler(s *v1service.DiscoveryService) *DiscoveryHandler {
	return &DiscoveryHandler{s}
}
func publicFilter(c *gin.Context) (catalog.Filter, error) {
	f, err := listFilter(c)
	return catalog.Filter{Page: f.Page, PageSize: f.PageSize, Query: f.Query, Status: c.Query("room_type")}, err
}
func publicID(c *gin.Context) (int64, error) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id < 1 {
		return 0, identity.ErrValidation
	}
	return id, nil
}
func (h *DiscoveryHandler) Rooms(c *gin.Context) {
	f, err := publicFilter(c)
	if err != nil {
		RespondDomainError(c, err)
		return
	}
	data, total, err := h.s.Rooms(c.Request.Context(), f)
	respondList(c, data, total, f, err)
}
func (h *DiscoveryHandler) Room(c *gin.Context) {
	id, err := publicID(c)
	if err != nil {
		RespondDomainError(c, err)
		return
	}
	data, err := h.s.Room(c.Request.Context(), id)
	respond(c, data, err, 200)
}
func (h *DiscoveryHandler) Trainers(c *gin.Context) {
	f, err := publicFilter(c)
	if err != nil {
		RespondDomainError(c, err)
		return
	}
	data, total, err := h.s.Trainers(c.Request.Context(), f)
	respondList(c, data, total, f, err)
}
func (h *DiscoveryHandler) Sessions(c *gin.Context) {
	id, err := publicID(c)
	if err != nil {
		RespondDomainError(c, err)
		return
	}
	f, err := publicFilter(c)
	if err != nil {
		RespondDomainError(c, err)
		return
	}
	data, total, err := h.s.Sessions(c.Request.Context(), id, c.Query("from"), c.Query("to"), f)
	c.Header("Cache-Control", "no-store")
	respondList(c, data, total, f, err)
}
func (h *DiscoveryHandler) Cover(c *gin.Context) {
	id, err := publicID(c)
	if err != nil {
		RespondDomainError(c, err)
		return
	}
	size := c.DefaultQuery("size", "card")
	data, version, err := h.s.Cover(c.Request.Context(), id, size)
	if err != nil {
		RespondDomainError(c, err)
		return
	}
	tag := fmt.Sprintf(`"public-room-%d-%d-%s"`, id, version, size)
	c.Header("ETag", tag)
	c.Header("Cache-Control", "public, no-cache")
	c.Header("X-Content-Type-Options", "nosniff")
	if c.GetHeader("If-None-Match") == tag {
		c.Status(304)
		return
	}
	c.Data(200, "image/jpeg", data)
}
