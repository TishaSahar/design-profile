package handler

import (
	"io"

	"design-profile/backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestHandler handles client project request endpoints.
type RequestHandler struct {
	svc *service.RequestService
}

func NewRequestHandler(svc *service.RequestService) *RequestHandler {
	return &RequestHandler{svc: svc}
}

// CreateRequest godoc
// @Summary      Submit project request
// @Description  Submits a new client project request with optional attachments (up to 10 photos or 1 PDF).
// @Tags         requests
// @Accept       multipart/form-data
// @Produce      json
// @Param        first_name   formData  string  true  "First name"
// @Param        last_name    formData  string  true  "Last name"
// @Param        contact      formData  string  true  "Contact info (phone/email/telegram)"
// @Param        description  formData  string  true  "Project description"
// @Param        consented    formData  bool    true  "Consent to personal data processing"
// @Param        attachments  formData  file    false "Attachments (up to 10 images or 1 PDF)"
// @Success      201  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Router       /requests [post]
func (h *RequestHandler) CreateRequest(c *gin.Context) {
	firstName := c.PostForm("first_name")
	lastName := c.PostForm("last_name")
	contact := c.PostForm("contact")
	description := c.PostForm("description")
	consented := c.PostForm("consented") == "true" || c.PostForm("consented") == "1"

	input := service.CreateRequestInput{
		FirstName:   firstName,
		LastName:    lastName,
		Contact:     contact,
		Description: description,
		Consented:   consented,
	}

	form, _ := c.MultipartForm()
	if form != nil {
		for _, fh := range form.File["attachments"] {
			f, err := fh.Open()
			if err != nil {
				internalError(c, "failed to read attachment")
				return
			}
			data, _ := io.ReadAll(f)
			f.Close()
			input.Attachments = append(input.Attachments, service.AttachmentInput{
				Data:        data,
				ContentType: fh.Header.Get("Content-Type"),
				Filename:    fh.Filename,
			})
		}
	}

	req, err := h.svc.Create(c.Request.Context(), input)
	if err != nil {
		badRequest(c, err.Error())
		return
	}
	created(c, req)
}

// ListRequests godoc
// @Summary      List project requests (admin)
// @Description  Returns all client project requests.
// @Tags         admin
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}
// @Router       /admin/requests [get]
func (h *RequestHandler) ListRequests(c *gin.Context) {
	requests, err := h.svc.List(c.Request.Context())
	if err != nil {
		internalError(c, "failed to fetch requests")
		return
	}
	ok(c, requests)
}

// GetRequest godoc
// @Summary      Get project request (admin)
// @Description  Returns a single project request with attachment metadata.
// @Tags         admin
// @Produce      json
// @Security     BearerAuth
// @Param        id  path      string  true  "Request UUID"
// @Success      200 {object}  map[string]interface{}
// @Failure      400 {object}  map[string]string
// @Failure      404 {object}  map[string]string
// @Router       /admin/requests/{id} [get]
func (h *RequestHandler) GetRequest(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		badRequest(c, "invalid request id")
		return
	}
	req, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		notFound(c, "request not found")
		return
	}
	ok(c, req)
}

// ServeAttachment godoc
// @Summary      Serve attachment file (admin)
// @Description  Returns the binary content of a request attachment.
// @Tags         admin
// @Produce      application/octet-stream
// @Security     BearerAuth
// @Param        id            path  string  true  "Request UUID"
// @Param        attachmentId  path  string  true  "Attachment UUID"
// @Success      200
// @Failure      400 {object}  map[string]string
// @Failure      404 {object}  map[string]string
// @Router       /admin/requests/{id}/attachments/{attachmentId} [get]
func (h *RequestHandler) ServeAttachment(c *gin.Context) {
	attachmentID, err := uuid.Parse(c.Param("attachmentId"))
	if err != nil {
		badRequest(c, "invalid attachment id")
		return
	}
	data, contentType, err := h.svc.GetAttachmentData(c.Request.Context(), attachmentID)
	if err != nil {
		notFound(c, "attachment not found")
		return
	}
	c.Data(200, contentType, data)
}
