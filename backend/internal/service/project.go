package service

import (
	"context"
	"fmt"
	"mime"
	"path/filepath"

	"design-profile/backend/internal/model"
	"design-profile/backend/internal/repository"

	"github.com/google/uuid"
)

type ProjectService struct {
	repo *repository.ProjectRepository
}

func NewProjectService(repo *repository.ProjectRepository) *ProjectService {
	return &ProjectService{repo: repo}
}

func (s *ProjectService) List(ctx context.Context) ([]model.Project, error) {
	return s.repo.List(ctx)
}

func (s *ProjectService) GetByID(ctx context.Context, id uuid.UUID) (*model.Project, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *ProjectService) Create(ctx context.Context, title, description string) (*model.Project, error) {
	if title == "" {
		return nil, fmt.Errorf("title is required")
	}
	if len(description) > 500 {
		return nil, fmt.Errorf("description must not exceed 500 characters")
	}
	return s.repo.Create(ctx, title, description)
}

func (s *ProjectService) Update(ctx context.Context, id uuid.UUID, title, description string, coverMediaID *uuid.UUID) (*model.Project, error) {
	if title == "" {
		return nil, fmt.Errorf("title is required")
	}
	if len(description) > 500 {
		return nil, fmt.Errorf("description must not exceed 500 characters")
	}
	return s.repo.Update(ctx, id, title, description, coverMediaID)
}

func (s *ProjectService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

// AddMedia stores a media file and returns its metadata.
func (s *ProjectService) AddMedia(ctx context.Context, projectID uuid.UUID, data []byte, contentType, filename string, sortOrder int) (*model.Media, error) {
	if !isAllowedMediaType(contentType) {
		return nil, fmt.Errorf("unsupported media type: %s", contentType)
	}
	if contentType == "" {
		contentType = mime.TypeByExtension(filepath.Ext(filename))
	}
	return s.repo.AddMedia(ctx, projectID, data, contentType, filename, sortOrder)
}

// GetMediaData returns the raw bytes and content type for a media file.
func (s *ProjectService) GetMediaData(ctx context.Context, mediaID uuid.UUID) ([]byte, string, error) {
	return s.repo.GetMediaData(ctx, mediaID)
}

// DeleteMedia removes a media file.
func (s *ProjectService) DeleteMedia(ctx context.Context, mediaID uuid.UUID) error {
	return s.repo.DeleteMedia(ctx, mediaID)
}

func isAllowedMediaType(ct string) bool {
	allowed := []string{
		"image/jpeg", "image/png", "image/gif", "image/webp", "image/svg+xml",
	}
	for _, a := range allowed {
		if ct == a {
			return true
		}
	}
	return false
}
