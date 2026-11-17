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
	Category    string `form:"category"`
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
// @Param        category     formData  string  false  "Work direction: residential (default) or commercial"
// @Param        media        formData  file    false  "Media files (images), stored in the order sent"
// @Param        cover_index  formData  int     false  "Index of the media file to use as the cover (0-based)"
// @Success      201  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Router       /admin/projects [post]
func (h *ProjectHandler) CreateProject(c *gin.Context) {
	var body createProjectBody
	if err := c.ShouldBind(&body); err != nil {
		badRequest(c, err.Error())
		return
	}

	project, err := h.svc.Create(c.Request.Context(), body.Title, body.Description, body.Category)
	if err != nil {
		badRequest(c, err.Error())
		return
	}

	// Handle optional media uploads, in the order they were sent.
	var mediaIDs []uuid.UUID
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
			media, err := h.svc.AddMedia(c.Request.Context(), project.ID, data, ct, fh.Filename, i)
			if err != nil {
				badRequest(c, err.Error())
				return
			}
			mediaIDs = append(mediaIDs, media.ID)
		}
	}

	// Optional cover: one of the files just uploaded.
	if raw := c.PostForm("cover_index"); raw != "" {
		idx, err := strconv.Atoi(raw)
		if err != nil || idx < 0 || idx >= len(mediaIDs) {
			badRequest(c, "cover_index must point to one of the uploaded files")
			return
		}
		cover := mediaIDs[idx]
		if _, err := h.svc.Update(c.Request.Context(), project.ID, project.Title, project.Description, project.Category, &cover); err != nil {
			internalError(c, "failed to set the cover")
			return
		}
	}

	// Return the full project.
	full, _ := h.svc.GetByID(c.Request.Context(), project.ID)
	created(c, full)
}

type updateProjectBody struct {
	Title        string  `json:"title"          binding:"required"`
	Description  string  `json:"description"`
	Category     string  `json:"category"       example:"residential"`
	CoverMediaID *string `json:"cover_media_id"`
}

// UpdateProject godoc
// @Summary      Update project (admin)
// @Description  Updates a project's title, description, category (residential by default) and cover photo. All fields are replaced, so send the current cover_media_id to keep the cover.
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

	project, err := h.svc.Update(c.Request.Context(), id, body.Title, body.Description, body.Category, coverMediaID)
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
// @Param        sortOrder  formData  int     false "Position among the project's photos; appended at the end when omitted"
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

	// Without an explicit position the photo goes after the existing ones.
	sortOrder := -1
	if raw := c.PostForm("sortOrder"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 0 {
			badRequest(c, "sortOrder must be a non-negative integer")
			return
		}
		sortOrder = parsed
	}

	media, err := h.svc.AddMedia(c.Request.Context(), projectID, data, ct, fh.Filename, sortOrder)
	if err != nil {
		badRequest(c, err.Error())
		return
	}
	created(c, media)
}

type reorderMediaBody struct {
	MediaIDs []string `json:"media_ids" binding:"required"`
}

// ReorderMedia godoc
// @Summary      Reorder project photos (admin)
// @Description  Sets the order of all photos of a project; media_ids must list every photo exactly once.
// @Tags         admin
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      string            true  "Project UUID"
// @Param        body  body      reorderMediaBody  true  "Photo IDs in the new order"
// @Success      200   {object}  map[string]interface{}
// @Failure      400   {object}  map[string]string
// @Router       /admin/projects/{id}/media/order [put]
func (h *ProjectHandler) ReorderMedia(c *gin.Context) {
	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		badRequest(c, "invalid project id")
		return
	}
	var body reorderMediaBody
	if err := c.ShouldBindJSON(&body); err != nil {
		badRequest(c, err.Error())
		return
	}
	ids := make([]uuid.UUID, 0, len(body.MediaIDs))
	for _, raw := range body.MediaIDs {
		id, err := uuid.Parse(raw)
		if err != nil {
			badRequest(c, "invalid media id: "+raw)
			return
		}
		ids = append(ids, id)
	}
	if err := h.svc.ReorderMedia(c.Request.Context(), projectID, ids); err != nil {
		badRequest(c, err.Error())
		return
	}
	project, err := h.svc.GetByID(c.Request.Context(), projectID)
	if err != nil {
		internalError(c, "failed to load project")
		return
	}
	ok(c, project)
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
