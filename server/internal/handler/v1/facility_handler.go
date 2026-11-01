package v1handler

import (
	"github.com/gin-gonic/gin"
	v1dto "github.com/tlong15306/ttcsn-gym-management/server/internal/dto/v1"
	"net/http"
)

func (h *ManagementHandler) Rooms(c *gin.Context) {
	f, err := boundFilter(c)
	if err != nil {
		RespondDomainError(c, err)
		return
	}
	data, total, err := h.facilities.Rooms(c.Request.Context(), f)
	respondList(c, data, total, f, err)
}

func (h *ManagementHandler) Room(c *gin.Context) {
	_, id, err := actorID(c)
	if err != nil {
		RespondDomainError(c, err)
		return
	}
	value, err := h.facilities.Room(c.Request.Context(), id)
	respond(c, value, err, http.StatusOK)
}

func (h *ManagementHandler) RoomImpact(c *gin.Context) {
	a, id, err := actorID(c)
	if err != nil {
		RespondDomainError(c, err)
		return
	}
	value, err := h.facilities.RoomImpact(c.Request.Context(), a, id)
	respond(c, value, err, http.StatusOK)
}

func (h *ManagementHandler) SaveRoom(c *gin.Context) {
	a, id, err := actorID(c)
	if err != nil {
		RespondDomainError(c, err)
		return
	}
	in, err := bindBody[v1dto.RoomRequest](c)
	if err != nil {
		RespondDomainError(c, err)
		return
	}
	value, err := h.facilities.SaveRoom(c.Request.Context(), a, id, in, RequestID(c))
	status := http.StatusOK
	if id == 0 {
		status = http.StatusCreated
	}
	respond(c, value, err, status)
}

func (h *ManagementHandler) Slots(c *gin.Context) {
	f, err := boundFilter(c)
	if err != nil {
		RespondDomainError(c, err)
		return
	}
	data, total, err := h.facilities.Slots(c.Request.Context(), f)
	respondList(c, data, total, f, err)
}

func (h *ManagementHandler) Slot(c *gin.Context) {
	_, id, err := actorID(c)
	if err != nil {
		RespondDomainError(c, err)
		return
	}
	value, err := h.facilities.Slot(c.Request.Context(), id)
	respond(c, value, err, http.StatusOK)
}

func (h *ManagementHandler) UpdateSlot(c *gin.Context) {
	a, id, err := actorID(c)
	if err != nil {
		RespondDomainError(c, err)
		return
	}
	in, err := bindBody[v1dto.SlotRequest](c)
	if err != nil {
		RespondDomainError(c, err)
		return
	}
	value, err := h.facilities.UpdateSlot(c.Request.Context(), a, id, in, RequestID(c))
	respond(c, value, err, http.StatusOK)
}

func (h *ManagementHandler) CreateSlot(c *gin.Context) {
	a, _, err := actorID(c)
	if err != nil {
		RespondDomainError(c, err)
		return
	}
	in, err := bindBody[v1dto.SlotRequest](c)
	if err != nil {
		RespondDomainError(c, err)
		return
	}
	value, err := h.facilities.CreateSlot(c.Request.Context(), a, in, RequestID(c))
	respond(c, value, err, http.StatusCreated)
}

func (h *ManagementHandler) SlotStatus(c *gin.Context) {
	a, id, err := actorID(c)
	if err != nil {
		RespondDomainError(c, err)
		return
	}
	in, err := bindBody[v1dto.SlotStatusRequest](c)
	if err != nil {
		RespondDomainError(c, err)
		return
	}
	value, err := h.facilities.SlotStatus(c.Request.Context(), a, id, in, RequestID(c))
	respond(c, value, err, http.StatusOK)
}

func (h *ManagementHandler) Sessions(c *gin.Context) {
	f, err := boundFilter(c)
	if err != nil {
		RespondDomainError(c, err)
		return
	}
	data, total, err := h.facilities.Sessions(c.Request.Context(), f)
	respondList(c, data, total, f, err)
}

func (h *ManagementHandler) Session(c *gin.Context) {
	_, id, err := actorID(c)
	if err != nil {
		RespondDomainError(c, err)
		return
	}
	value, err := h.facilities.Session(c.Request.Context(), id)
	respond(c, value, err, http.StatusOK)
}

func (h *ManagementHandler) SaveSession(c *gin.Context) {
	a, id, err := actorID(c)
	if err != nil {
		RespondDomainError(c, err)
		return
	}
	in, err := bindBody[v1dto.SessionRequest](c)
	if err != nil {
		RespondDomainError(c, err)
		return
	}
	value, err := h.facilities.SaveSession(c.Request.Context(), a, id, in, RequestID(c))
	status := http.StatusOK
	if id == 0 {
		status = http.StatusCreated
	}
	respond(c, value, err, status)
}
