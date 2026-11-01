package v1handler

import (
	"github.com/gin-gonic/gin"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/booking"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/identity"
	v1dto "github.com/tlong15306/ttcsn-gym-management/server/internal/dto/v1"
	v1service "github.com/tlong15306/ttcsn-gym-management/server/internal/service/v1"
	"strconv"
)

type BookingHandler struct{ s *v1service.BookingService }

func NewBookingHandler(s *v1service.BookingService) *BookingHandler { return &BookingHandler{s} }
func (h *BookingHandler) Create(c *gin.Context) {
	a, member, err := actorID(c)
	if err != nil {
		RespondDomainError(c, err)
		return
	}
	in, err := bindBody[v1dto.CreateBooking](c)
	if err != nil {
		RespondDomainError(c, err)
		return
	}
	b, err := h.s.Create(c.Request.Context(), a, member, in, c.GetHeader("Idempotency-Key"), RequestID(c))
	respond(c, b, err, 201)
}
func (h *BookingHandler) CreateBatch(c *gin.Context) {
	a, member, err := actorID(c)
	if err != nil {
		RespondDomainError(c, err)
		return
	}
	in, err := bindBody[v1dto.CreateBookingBatch](c)
	if err != nil {
		RespondDomainError(c, err)
		return
	}
	bookings, err := h.s.CreateBatch(c.Request.Context(), a, member, in, c.GetHeader("Idempotency-Key"), RequestID(c))
	respond(c, bookings, err, 201)
}
func (h *BookingHandler) Get(c *gin.Context) {
	a, id, err := actorID(c)
	if err != nil {
		RespondDomainError(c, err)
		return
	}
	b, err := h.s.Get(c.Request.Context(), a, id)
	respond(c, b, err, 200)
}
func (h *BookingHandler) List(c *gin.Context) {
	a, _, err := actorID(c)
	if err != nil {
		RespondDomainError(c, err)
		return
	}
	f, err := boundFilter(c)
	if err != nil {
		RespondDomainError(c, err)
		return
	}
	pt := int64(0)
	if raw := c.Query("pt"); raw != "" {
		pt, err = strconv.ParseInt(raw, 10, 64)
		if err != nil {
			RespondDomainError(c, identity.ErrValidation)
			return
		}
	}
	if raw := c.Query("history"); raw != "" && raw != "true" && raw != "false" {
		RespondDomainError(c, identity.ErrValidation)
		return
	}
	list, total, err := h.s.List(c.Request.Context(), a, booking.Filter{Filter: f, From: c.Query("from"), To: c.Query("to"), PTID: pt, History: c.Query("history") == "true"})
	respondList(c, list, total, f, err)
}
func (h *BookingHandler) Act(c *gin.Context) {
	a, id, err := actorID(c)
	if err != nil {
		RespondDomainError(c, err)
		return
	}
	in, err := bindBody[v1dto.BookingAction](c)
	if err != nil {
		RespondDomainError(c, err)
		return
	}
	b, err := h.s.Act(c.Request.Context(), a, id, in, c.FullPath() == "/api/v1/bookings/:id/cancel", RequestID(c))
	respond(c, b, err, 200)
}
func (h *BookingHandler) Sessions(c *gin.Context) {
	a, _, err := actorID(c)
	if err != nil {
		RespondDomainError(c, err)
		return
	}
	f, err := boundFilter(c)
	if err != nil {
		RespondDomainError(c, err)
		return
	}
	list, total, err := h.s.Sessions(c.Request.Context(), a, f)
	respondList(c, list, total, f, err)
}
func (h *BookingHandler) PTs(c *gin.Context) {
	a, _, err := actorID(c)
	if err != nil {
		RespondDomainError(c, err)
		return
	}
	f, err := boundFilter(c)
	if err != nil {
		RespondDomainError(c, err)
		return
	}
	id, err := strconv.ParseInt(c.Query("room_session_id"), 10, 64)
	if err != nil {
		RespondDomainError(c, identity.ErrValidation)
		return
	}
	list, total, err := h.s.PTs(c.Request.Context(), a, id, f)
	respondList(c, list, total, f, err)
}
