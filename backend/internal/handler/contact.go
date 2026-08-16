package handler

import (
	"net/http"
	"os"

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
// @Description  Returns the designer's public contact information (sourced from config/about/contacts.yaml and bio.txt).
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

// ServePhoto godoc
// @Summary      Get designer profile photo
// @Description  Returns the designer's profile photo from config/about/.
// @Tags         contacts
// @Produce      image/jpeg
// @Success      200
// @Failure      404
// @Router       /about/photo [get]
func (h *ContactHandler) ServePhoto(c *gin.Context) {
	path := h.svc.PhotoPath()
	if path == "" {
		c.Status(http.StatusNotFound)
		return
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		c.Status(http.StatusNotFound)
		return
	}
	c.File(path)
}
