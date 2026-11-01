package v1handler

import (
	"github.com/gin-gonic/gin"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/facility"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/identity"
	v1dto "github.com/tlong15306/ttcsn-gym-management/server/internal/dto/v1"
	v1service "github.com/tlong15306/ttcsn-gym-management/server/internal/service/v1"
	"net/http"
	"strconv"
)

type CalendarHandler struct{ service *v1service.CalendarService }

func NewCalendarHandler(s *v1service.CalendarService) *CalendarHandler { return &CalendarHandler{s} }
func (h *CalendarHandler) Weekly(c *gin.Context) {
	a, id, err := actorID(c)
	if err != nil {
		RespondDomainError(c, err)
		return
	}
	s, err := h.service.Snapshot(c.Request.Context(), a, id)
	respond(c, s, err, 200)
}
func (h *CalendarHandler) Calendar(c *gin.Context) {
	_, id, err := actorID(c)
	if err != nil {
		RespondDomainError(c, err)
		return
	}
	days, err := h.service.Days(c.Request.Context(), id, c.Query("from"), c.Query("to"))
	respond(c, days, err, 200)
}
func (h *CalendarHandler) History(c *gin.Context) {
	a, id, err := actorID(c)
	if err != nil {
		RespondDomainError(c, err)
		return
	}
	f, err := boundFilter(c)
	if err != nil {
		RespondDomainError(c, err)
		return
	}
	s, err := h.service.Snapshot(c.Request.Context(), a, id)
	if err != nil {
		RespondDomainError(c, err)
		return
	}
	type event struct {
		Kind      string                      `json:"kind"`
		Version   *facility.ScheduleVersion   `json:"version,omitempty"`
		Exception *facility.CalendarException `json:"exception,omitempty"`
	}
	events := []event{}
	for i := len(s.Versions) - 1; i >= 0; i-- {
		events = append(events, event{Kind: "WEEKLY", Version: &s.Versions[i]})
	}
	for i := len(s.Exceptions) - 1; i >= 0; i-- {
		events = append(events, event{Kind: "EXCEPTION", Exception: &s.Exceptions[i]})
	}
	start := (f.Page - 1) * f.PageSize
	if start > len(events) {
		start = len(events)
	}
	end := start + f.PageSize
	if end > len(events) {
		end = len(events)
	}
	respondList(c, events[start:end], len(events), f, nil)
}
func (h *CalendarHandler) Preview(c *gin.Context) {
	a, id, err := actorID(c)
	if err != nil {
		RespondDomainError(c, err)
		return
	}
	in, err := bindBody[v1dto.CalendarMutationRequest](c)
	if err != nil {
		RespondDomainError(c, err)
		return
	}
	p, err := h.service.Preview(c.Request.Context(), a, id, in)
	respond(c, p, err, 200)
}
func (h *CalendarHandler) confirm(c *gin.Context, kind string) {
	a, id, err := actorID(c)
	if err != nil {
		RespondDomainError(c, err)
		return
	}
	in, err := bindBody[v1dto.CalendarConfirmRequest](c)
	if err != nil {
		RespondDomainError(c, err)
		return
	}
	if in.Kind != kind {
		RespondDomainError(c, identity.ErrValidation)
		return
	}
	if raw := c.Param("exceptionId"); raw != "" {
		exception, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || exception < 1 || exception != in.ExceptionID {
			RespondDomainError(c, identity.ErrValidation)
			return
		}
	} else if kind == "EXCEPTION" && in.ExceptionID != 0 {
		RespondDomainError(c, identity.ErrValidation)
		return
	}
	result, err := h.service.Confirm(c.Request.Context(), a, id, in, c.GetHeader("Idempotency-Key"), RequestID(c))
	respond(c, result, err, http.StatusOK)
}
func (h *CalendarHandler) Publish(c *gin.Context)     { h.confirm(c, "WEEKLY") }
func (h *CalendarHandler) Exception(c *gin.Context)   { h.confirm(c, "EXCEPTION") }
func (h *CalendarHandler) Withdraw(c *gin.Context)    { h.confirm(c, "WITHDRAW") }
func (h *CalendarHandler) Materialize(c *gin.Context) { h.confirm(c, "EXTEND") }
