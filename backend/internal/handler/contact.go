package handler

import (
	"design-profile/backend/internal/service"

	"github.com/gin-gonic/gin"
)

// ContactHandler handles contact information endpoints.
type ContactHandler struct {
	svc *service.ContactService
}

func NewContactHandler(svc *service.ContactService) *ContactHandler {
	return &ContactHandler{svc: svc}
}

// GetContacts godoc
// @Summary      Get designer contacts
// @Description  Returns the designer's public contact information (sourced from config/about/contacts.yaml).
// @Tags         contacts
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Router       /contacts [get]
func (h *ContactHandler) GetContacts(c *gin.Context) {
	contacts, err := h.svc.Get(c.Request.Context())
	if err != nil {
		internalError(c, "failed to fetch contacts")
		return
	}
	ok(c, contacts)
}
