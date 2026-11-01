package v1handler

import (
	"github.com/gin-gonic/gin"
	"github.com/tlong15306/ttcsn-gym-management/server/internal/domain/identity"
	v1dto "github.com/tlong15306/ttcsn-gym-management/server/internal/dto/v1"
	v1service "github.com/tlong15306/ttcsn-gym-management/server/internal/service/v1"
	"net/http"
	"path/filepath"
	"strconv"
)

func (h *ManagementHandler) Trainers(c *gin.Context) {
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
	if raw := c.Query("approved"); raw != "" {
		flag, e := strconv.ParseBool(raw)
		if e != nil {
			RespondDomainError(c, identity.ErrValidation)
			return
		}
		if flag {
			f.Status = "APPROVED"
		}
	}
	value, total, err := h.trainers.List(c.Request.Context(), a, f)
	respondList(c, value, total, f, err)
}

func (h *ManagementHandler) Trainer(c *gin.Context) {
	a, id, err := actorID(c)
	if err != nil {
		RespondDomainError(c, err)
		return
	}
	value, err := h.trainers.Get(c.Request.Context(), a, id)
	respond(c, value, err, http.StatusOK)
}

func (h *ManagementHandler) OwnTrainer(c *gin.Context) {
	a, _, err := actorID(c)
	if err != nil {
		RespondDomainError(c, err)
		return
	}
	value, err := h.trainers.Own(c.Request.Context(), a)
	respond(c, value, err, http.StatusOK)
}

func (h *ManagementHandler) SaveTrainer(c *gin.Context) {
	a, _, err := actorID(c)
	if err != nil {
		RespondDomainError(c, err)
		return
	}
	in, err := bindBody[v1dto.TrainerProfileRequest](c)
	if err != nil {
		RespondDomainError(c, err)
		return
	}
	value, err := h.trainers.SaveProfile(c.Request.Context(), a, in, RequestID(c))
	respond(c, value, err, http.StatusOK)
}

func (h *ManagementHandler) SubmitTrainer(c *gin.Context) {
	a, _, err := actorID(c)
	if err != nil {
		RespondDomainError(c, err)
		return
	}
	value, err := h.trainers.Submit(c.Request.Context(), a, RequestID(c))
	respond(c, value, err, http.StatusOK)
}

func (h *ManagementHandler) TrainerDecision(c *gin.Context) {
	a, id, err := actorID(c)
	if err != nil {
		RespondDomainError(c, err)
		return
	}
	in, err := bindBody[v1dto.TrainerDecisionRequest](c)
	if err != nil {
		RespondDomainError(c, err)
		return
	}
	value, err := h.trainers.Decision(c.Request.Context(), a, id, in, RequestID(c))
	respond(c, value, err, http.StatusOK)
}

func (h *ManagementHandler) UploadCertificate(c *gin.Context) {
	a, _, err := actorID(c)
	if err != nil {
		RespondDomainError(c, err)
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, v1service.MaxCertificateBytes+65536)
	fileHeader, err := c.FormFile("file")
	if c.Request.MultipartForm != nil {
		defer c.Request.MultipartForm.RemoveAll()
	}
	if err != nil {
		RespondDomainError(c, identity.ErrValidation)
		return
	}
	file, err := fileHeader.Open()
	if err != nil {
		RespondDomainError(c, err)
		return
	}
	defer file.Close()
	value, err := h.trainers.UploadCertificate(c.Request.Context(), a, file, RequestID(c))
	respond(c, value, err, http.StatusOK)
}

func (h *ManagementHandler) DownloadCertificate(c *gin.Context) {
	a, id, err := actorID(c)
	if err != nil {
		RespondDomainError(c, err)
		return
	}
	path, err := h.trainers.Certificate(c.Request.Context(), a, id)
	if err != nil {
		RespondDomainError(c, err)
		return
	}
	c.Header("Cache-Control", "no-store")
	c.Header("X-Content-Type-Options", "nosniff")
	c.FileAttachment(path, "chung-chi"+filepath.Ext(path))
}

func (h *ManagementHandler) Hours(c *gin.Context) {
	a, id, err := actorID(c)
	if err != nil {
		RespondDomainError(c, err)
		return
	}
	value, err := h.trainers.Hours(c.Request.Context(), a, id)
	respond(c, value, err, http.StatusOK)
}

func (h *ManagementHandler) SaveHours(c *gin.Context) {
	a, id, err := actorID(c)
	if err != nil {
		RespondDomainError(c, err)
		return
	}
	in, err := bindBody[v1dto.AvailabilityRequest](c)
	if err != nil {
		RespondDomainError(c, err)
		return
	}
	value, err := h.trainers.SaveHours(c.Request.Context(), a, id, in, RequestID(c))
	respond(c, value, err, http.StatusOK)
}

func (h *ManagementHandler) Absences(c *gin.Context) {
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
	value, total, err := h.trainers.Absences(c.Request.Context(), a, id, f)
	respondList(c, value, total, f, err)
}

func (h *ManagementHandler) SaveAbsence(c *gin.Context) {
	a, id, err := actorID(c)
	if err != nil {
		RespondDomainError(c, err)
		return
	}
	in, err := bindBody[v1dto.UnavailabilityRequest](c)
	if err != nil {
		RespondDomainError(c, err)
		return
	}
	value, err := h.trainers.SaveAbsence(c.Request.Context(), a, id, in, RequestID(c))
	respond(c, value, err, http.StatusCreated)
}

func (h *ManagementHandler) DeleteAbsence(c *gin.Context) {
	a, id, err := actorID(c)
	if err != nil {
		RespondDomainError(c, err)
		return
	}
	absenceID, err := strconv.ParseInt(c.Param("absenceId"), 10, 64)
	if err != nil || absenceID < 1 {
		RespondDomainError(c, identity.ErrValidation)
		return
	}
	if err = h.trainers.DeleteAbsence(c.Request.Context(), a, id, absenceID, RequestID(c)); err != nil {
		RespondDomainError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
