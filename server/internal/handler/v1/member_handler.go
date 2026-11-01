package v1handler

import (
	"github.com/gin-gonic/gin"
	v1dto "github.com/tlong15306/ttcsn-gym-management/server/internal/dto/v1"
	"net/http"
)

func (h *ManagementHandler) Members(c *gin.Context) {
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
	data, total, err := h.members.List(c.Request.Context(), a, f)
	respondList(c, data, total, f, err)
}

func (h *ManagementHandler) EligibleAccounts(c *gin.Context) {
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
	value, total, err := h.members.EligibleAccounts(c.Request.Context(), a, f)
	respondList(c, value, total, f, err)
}

func (h *ManagementHandler) Member(c *gin.Context) {
	a, id, err := actorID(c)
	if err != nil {
		RespondDomainError(c, err)
		return
	}
	value, err := h.members.Get(c.Request.Context(), a, id)
	respond(c, value, err, http.StatusOK)
}

func (h *ManagementHandler) SaveMember(c *gin.Context) {
	a, id, err := actorID(c)
	if err != nil {
		RespondDomainError(c, err)
		return
	}
	in, err := bindBody[v1dto.MemberRequest](c)
	if err != nil {
		RespondDomainError(c, err)
		return
	}
	if id == 0 {
		value, err := h.members.Create(c.Request.Context(), a, in, RequestID(c))
		respond(c, value, err, http.StatusCreated)
	} else {
		value, err := h.members.Update(c.Request.Context(), a, id, in, RequestID(c))
		respond(c, value, err, http.StatusOK)
	}
}

func (h *ManagementHandler) Membership(c *gin.Context) {
	a, id, err := actorID(c)
	if err != nil {
		RespondDomainError(c, err)
		return
	}
	in, err := bindBody[v1dto.MembershipRequest](c)
	if err != nil {
		RespondDomainError(c, err)
		return
	}
	value, err := h.members.Membership(c.Request.Context(), a, id, in, RequestID(c))
	respond(c, value, err, http.StatusOK)
}
