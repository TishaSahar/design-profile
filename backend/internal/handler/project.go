package handler

import (
	"io"
	"net/http"
	"strconv"

	"design-profile/backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ProjectHandler handles project-related endpoints.
type ProjectHandler struct {
	svc *service.ProjectService
}

func NewProjectHandler(svc *service.ProjectService) *ProjectHandler {
	return &ProjectHandler{svc: svc}
}

// ListProjects godoc
// @Summary      List projects
// @Description  Returns all portfolio projects with their media metadata.
// @Tags         projects
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Router       /projects [get]
func (h *ProjectHandler) ListProjects(c *gin.Context) {
	projects, err := h.svc.List(c.Request.Context())
	if err != nil {
		internalError(c, "failed to fetch projects")
		return
	}
	ok(c, projects)
}

// GetProject godoc
// @Summary      Get project
// @Description  Returns a single project with all media metadata.
// @Tags         projects
// @Produce      json
// @Param        id  path      string  true  "Project UUID"
// @Success      200 {object}  map[string]interface{}
// @Failure      400 {object}  map[string]string
// @Failure      404 {object}  map[string]string
// @Router       /projects/{id} [get]
func (h *ProjectHandler) GetProject(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		badRequest(c, "invalid project id")
		return
	}
	project, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		notFound(c, "project not found")
		return
	}
	ok(c, project)
}

// ServeMedia godoc
// @Summary      Serve media file
// @Description  Returns the binary content of a project media file.
// @Tags         projects
// @Produce      application/octet-stream
// @Param        id       path  string  true  "Project UUID"
// @Param        mediaId  path  string  true  "Media UUID"
// @Success      200
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /projects/{id}/media/{mediaId} [get]
func (h *ProjectHandler) ServeMedia(c *gin.Context) {
	mediaID, err := uuid.Parse(c.Param("mediaId"))
	if err != nil {
		badRequest(c, "invalid media id")
		return
	}
	data, contentType, err := h.svc.GetMediaData(c.Request.Context(), mediaID)
	if err != nil {
		notFound(c, "media not found")
		return
	}
	c.Data(http.StatusOK, contentType, data)
}

type createProjectBody struct {
	Title       string `form:"title"       binding:"required"`
	Description string `form:"description"`
}

// CreateProject godoc
// @Summary      Create project (admin)
// @Description  Creates a new portfolio project with optional media files.
// @Tags         admin
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        title        formData  string  true   "Project title"
// @Param        description  formData  string  false  "Project description (max 500 chars)"
// @Param        media        formData  file    false  "Media files (images)"
// @Success      201  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Router       /admin/projects [post]
func (h *ProjectHandler) CreateProject(c *gin.Context) {
	var body createProjectBody
	if err := c.ShouldBind(&body); err != nil {
		badRequest(c, err.Error())
		return
	}

	project, err := h.svc.Create(c.Request.Context(), body.Title, body.Description)
	if err != nil {
		badRequest(c, err.Error())
		return
	}

	// Handle optional media uploads.
	form, _ := c.MultipartForm()
	if form != nil {
		files := form.File["media"]
		for i, fh := range files {
			f, err := fh.Open()
			if err != nil {
				internalError(c, "failed to read uploaded file")
				return
			}
			data, _ := io.ReadAll(f)
			f.Close()

			ct := fh.Header.Get("Content-Type")
			if _, err := h.svc.AddMedia(c.Request.Context(), project.ID, data, ct, fh.Filename, i); err != nil {
				badRequest(c, err.Error())
				return
			}
		}
	}

	// Return the full project.
	full, _ := h.svc.GetByID(c.Request.Context(), project.ID)
	created(c, full)
}

type updateProjectBody struct {
	Title        string  `json:"title"          binding:"required"`
	Description  string  `json:"description"`
	CoverMediaID *string `json:"cover_media_id"`
}

// UpdateProject godoc
// @Summary      Update project (admin)
// @Description  Updates a project's title, description, and cover photo.
// @Tags         admin
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      string             true  "Project UUID"
// @Param        body  body      updateProjectBody  true  "Update payload"
// @Success      200   {object}  map[string]interface{}
// @Failure      400   {object}  map[string]string
// @Router       /admin/projects/{id} [put]
func (h *ProjectHandler) UpdateProject(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		badRequest(c, "invalid project id")
		return
	}

	var body updateProjectBody
	if err := c.ShouldBindJSON(&body); err != nil {
		badRequest(c, err.Error())
		return
	}

	var coverMediaID *uuid.UUID
	if body.CoverMediaID != nil && *body.CoverMediaID != "" {
		parsed, err := uuid.Parse(*body.CoverMediaID)
		if err != nil {
			badRequest(c, "invalid cover_media_id")
			return
		}
		coverMediaID = &parsed
	}

	project, err := h.svc.Update(c.Request.Context(), id, body.Title, body.Description, coverMediaID)
	if err != nil {
		badRequest(c, err.Error())
		return
	}
	ok(c, project)
}

// DeleteProject godoc
// @Summary      Delete project (admin)
// @Description  Deletes a project and all its associated media.
// @Tags         admin
// @Produce      json
// @Security     BearerAuth
// @Param        id  path      string  true  "Project UUID"
// @Success      200 {object}  map[string]string
// @Failure      400 {object}  map[string]string
// @Router       /admin/projects/{id} [delete]
func (h *ProjectHandler) DeleteProject(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		badRequest(c, "invalid project id")
		return
	}
	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		internalError(c, "failed to delete project")
		return
	}
	ok(c, gin.H{"message": "project deleted"})
}

// AddMedia godoc
// @Summary      Add media to project (admin)
// @Description  Uploads additional image files to an existing project.
// @Tags         admin
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        id         path      string  true  "Project UUID"
// @Param        media      formData  file    true  "Image file"
// @Param        sortOrder  formData  int     false "Sort order"
// @Success      201 {object}  map[string]interface{}
// @Failure      400 {object}  map[string]string
// @Router       /admin/projects/{id}/media [post]
func (h *ProjectHandler) AddMedia(c *gin.Context) {
	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		badRequest(c, "invalid project id")
		return
	}

	fh, err := c.FormFile("media")
	if err != nil {
		badRequest(c, "media file is required")
		return
	}

	f, err := fh.Open()
	if err != nil {
		internalError(c, "failed to open file")
		return
	}
	defer f.Close()

	data, _ := io.ReadAll(f)
	ct := fh.Header.Get("Content-Type")

	sortOrderStr := c.DefaultPostForm("sortOrder", "0")
	sortOrder, _ := strconv.Atoi(sortOrderStr)

	media, err := h.svc.AddMedia(c.Request.Context(), projectID, data, ct, fh.Filename, sortOrder)
	if err != nil {
		badRequest(c, err.Error())
		return
	}
	created(c, media)
}

// DeleteMedia godoc
// @Summary      Delete media file (admin)
// @Description  Removes a single media file from a project.
// @Tags         admin
// @Produce      json
// @Security     BearerAuth
// @Param        id       path  string  true  "Project UUID"
// @Param        mediaId  path  string  true  "Media UUID"
// @Success      200 {object}  map[string]string
// @Failure      400 {object}  map[string]string
// @Router       /admin/projects/{id}/media/{mediaId} [delete]
func (h *ProjectHandler) DeleteMedia(c *gin.Context) {
	mediaID, err := uuid.Parse(c.Param("mediaId"))
	if err != nil {
		badRequest(c, "invalid media id")
		return
	}
	if err := h.svc.DeleteMedia(c.Request.Context(), mediaID); err != nil {
		internalError(c, "failed to delete media")
		return
	}
	ok(c, gin.H{"message": "media deleted"})
}
