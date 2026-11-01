package v1handler

import (
	"github.com/gin-gonic/gin"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/catalog"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/identity"
	v1dto "github.com/tlong15306/ttcsn-gym-management/server/internal/dto/v1"
	v1service "github.com/tlong15306/ttcsn-gym-management/server/internal/service/v1"
)

type ReviewHandler struct{ s *v1service.ReviewService }

func NewReviewHandler(s *v1service.ReviewService) *ReviewHandler { return &ReviewHandler{s: s} }

func (h *ReviewHandler) Create(c *gin.Context) {
	actor, id, err := actorID(c)
	if err != nil {
		RespondDomainError(c, err)
		return
	}
	in, err := bindBody[v1dto.ReviewInput](c)
	if err != nil {
		RespondDomainError(c, err)
		return
	}
	value, err := h.s.Create(c.Request.Context(), actor, id, in, RequestID(c))
	respond(c, value, err, 201)
}

func (h *ReviewHandler) Get(c *gin.Context) {
	actor, id, err := actorID(c)
	if err != nil {
		RespondDomainError(c, err)
		return
	}
	value, err := h.s.Get(c.Request.Context(), actor, id)
	respond(c, value, err, 200)
}

func (h *ReviewHandler) ListPT(c *gin.Context) {
	actor, id, err := actorID(c)
	if err != nil {
		RespondDomainError(c, err)
		return
	}
	page, err := parsePositive(c.Query("page"), 1)
	if err != nil {
		RespondDomainError(c, identity.ErrValidation)
		return
	}
	pageSize, err := parsePositive(c.Query("page_size"), 20)
	if err != nil || pageSize > 100 {
		RespondDomainError(c, identity.ErrValidation)
		return
	}
	f := catalog.Filter{Page: page, PageSize: pageSize}
	list, total, err := h.s.ListPT(c.Request.Context(), actor, id, f)
	respondList(c, list, total, f, err)
}
