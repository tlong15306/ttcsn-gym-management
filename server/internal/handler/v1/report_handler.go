package v1handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/identity"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/report"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/middleware"
	v1service "github.com/tlong15306/ttcsn-gym-management/server/internal/service/v1"
)

type ReportHandler struct{ s *v1service.ReportService }

func NewReportHandler(s *v1service.ReportService) *ReportHandler { return &ReportHandler{s: s} }

func reportPeriod(c *gin.Context) report.PeriodInput {
	return report.PeriodInput{From: c.Query("from"), To: c.Query("to")}
}

func (h *ReportHandler) Operations(c *gin.Context) {
	actor, ok := middleware.Actor(c)
	if !ok {
		RespondDomainError(c, identity.ErrUnauthorized)
		return
	}
	roomID := int64(0)
	if raw := c.Query("room_id"); raw != "" {
		parsed, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || parsed < 1 {
			RespondDomainError(c, identity.ErrValidation)
			return
		}
		roomID = parsed
	}
	value, err := h.s.Operations(c.Request.Context(), actor, reportPeriod(c), roomID)
	respond(c, value, err, 200)
}

func (h *ReportHandler) Trainer(c *gin.Context) {
	actor, ptID, err := actorID(c)
	if err != nil {
		RespondDomainError(c, err)
		return
	}
	value, err := h.s.Trainer(c.Request.Context(), actor, ptID, reportPeriod(c))
	respond(c, value, err, 200)
}
