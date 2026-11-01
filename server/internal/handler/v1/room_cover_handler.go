package v1handler

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/identity"
	v1service "github.com/tlong15306/ttcsn-gym-management/server/internal/service/v1"
	"net/http"
)

type RoomCoverHandler struct{ s *v1service.RoomCoverService }

func NewRoomCoverHandler(s *v1service.RoomCoverService) *RoomCoverHandler {
	return &RoomCoverHandler{s}
}
func (h *RoomCoverHandler) Upload(c *gin.Context) {
	a, id, err := actorID(c)
	if err != nil {
		RespondDomainError(c, err)
		return
	}
	if a.Role != identity.RoleAdmin && a.Role != identity.RoleManager {
		RespondDomainError(c, identity.ErrForbidden)
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 5*1024*1024+8192)
	if err = c.Request.ParseMultipartForm(5 * 1024 * 1024); err != nil {
		RespondDomainError(c, identity.ErrValidation)
		return
	}
	if c.Request.MultipartForm != nil {
		defer c.Request.MultipartForm.RemoveAll()
	}
	f, _, err := c.Request.FormFile("file")
	if err != nil {
		RespondDomainError(c, identity.ErrValidation)
		return
	}
	defer f.Close()
	value, err := h.s.Upload(c.Request.Context(), a, id, f, c.PostForm("alt_text"), c.PostForm("is_demo") == "true", RequestID(c))
	respond(c, value, err, 200)
}
func (h *RoomCoverHandler) Delete(c *gin.Context) {
	a, id, err := actorID(c)
	if err != nil {
		RespondDomainError(c, err)
		return
	}
	value, err := h.s.Delete(c.Request.Context(), a, id, RequestID(c))
	respond(c, value, err, 200)
}
func (h *RoomCoverHandler) Read(c *gin.Context) {
	_, id, err := actorID(c)
	if err != nil {
		RespondDomainError(c, err)
		return
	}
	size := c.DefaultQuery("size", "detail")
	if size != "card" && size != "detail" {
		RespondDomainError(c, identity.ErrValidation)
		return
	}
	data, version, err := h.s.Read(c.Request.Context(), id, size == "card")
	if err != nil {
		RespondDomainError(c, err)
		return
	}
	tag := fmt.Sprintf(`"room-%d-%d-%s"`, id, version, size)
	c.Header("ETag", tag)
	c.Header("Cache-Control", "private, no-cache")
	c.Header("X-Content-Type-Options", "nosniff")
	if c.GetHeader("If-None-Match") == tag {
		c.Status(304)
		return
	}
	c.Data(200, "image/jpeg", data)
}
