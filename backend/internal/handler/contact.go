package handler

import (
	"design-profile/backend/internal/model"
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
// @Description  Returns the designer's public contact information.
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

type updateContactsBody struct {
	Telegram  string `json:"telegram"`
	Instagram string `json:"instagram"`
	Email     string `json:"email"`
}

// UpdateContacts godoc
// @Summary      Update contacts (admin)
// @Description  Updates the designer's contact information.
// @Tags         admin
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      updateContactsBody  true  "Contact information"
// @Success      200   {object}  map[string]interface{}
// @Failure      400   {object}  map[string]string
// @Router       /admin/contacts [put]
func (h *ContactHandler) UpdateContacts(c *gin.Context) {
	var body updateContactsBody
	if err := c.ShouldBindJSON(&body); err != nil {
		badRequest(c, err.Error())
		return
	}
	contacts, err := h.svc.Update(c.Request.Context(), &model.Contacts{
		Telegram:  body.Telegram,
		Instagram: body.Instagram,
		Email:     body.Email,
	})
	if err != nil {
		internalError(c, "failed to update contacts")
		return
	}
	ok(c, contacts)
}
