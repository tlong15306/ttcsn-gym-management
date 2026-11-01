package v1handler

import (
	"github.com/gin-gonic/gin"
	v1dto "github.com/tlong15306/ttcsn-gym-management/server/internal/dto/v1"
	v1service "github.com/tlong15306/ttcsn-gym-management/server/internal/service/v1"
)

type AttendanceHandler struct{ s *v1service.AttendanceService }

func NewAttendanceHandler(s *v1service.AttendanceService) *AttendanceHandler {
	return &AttendanceHandler{s}
}
func (h *AttendanceHandler) Get(c *gin.Context) {
	a, id, err := actorID(c)
	if err != nil {
		RespondDomainError(c, err)
		return
	}
	value, err := h.s.Get(c.Request.Context(), a, id)
	respond(c, value, err, 200)
}
func (h *AttendanceHandler) Record(c *gin.Context) {
	a, id, err := actorID(c)
	if err != nil {
		RespondDomainError(c, err)
		return
	}
	in, err := bindBody[v1dto.AttendanceInput](c)
	if err != nil {
		RespondDomainError(c, err)
		return
	}
	value, err := h.s.Record(c.Request.Context(), a, id, in, RequestID(c))
	status := 200
	if value.Created {
		status = 201
	}
	respond(c, value.Record, err, status)
}
