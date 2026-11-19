package service

import (
	"context"
	"fmt"
	"mime"
	"path/filepath"
	"unicode/utf8"

	"design-profile/backend/internal/model"

	"github.com/google/uuid"
)

type projectRepo interface {
	List(ctx context.Context) ([]model.Project, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.Project, error)
	Create(ctx context.Context, f model.ProjectFields) (*model.Project, error)
	Update(ctx context.Context, id uuid.UUID, f model.ProjectFields, coverMediaID *uuid.UUID) (*model.Project, error)
	Delete(ctx context.Context, id uuid.UUID) error
	AddMedia(ctx context.Context, projectID uuid.UUID, data []byte, contentType, filename string, sortOrder int) (*model.Media, error)
	GetMediaData(ctx context.Context, mediaID uuid.UUID) ([]byte, string, error)
	DeleteMedia(ctx context.Context, mediaID uuid.UUID) error
	ReorderMedia(ctx context.Context, projectID uuid.UUID, mediaIDs []uuid.UUID) error
	ReorderProjects(ctx context.Context, category string, projectIDs []uuid.UUID) error
}

// Project categories: the two work directions shown on the public site.
const (
	CategoryResidential = "residential"
	CategoryCommercial  = "commercial"
)

// normalizeCategory defaults an empty category to residential and rejects unknown values.
func normalizeCategory(category string) (string, error) {
	switch category {
	case "":
		return CategoryResidential, nil
	case CategoryResidential, CategoryCommercial:
		return category, nil
	default:
		return "", fmt.Errorf("unknown project category: %s (allowed: %s, %s)",
			category, CategoryResidential, CategoryCommercial)
	}
}

// maxDescription is the length limit of a project description, in characters.
const maxDescription = 1000

// prepareProject validates the fields and defaults an empty category.
func prepareProject(f model.ProjectFields) (model.ProjectFields, error) {
	if f.Title == "" {
		return f, fmt.Errorf("title is required")
	}
	// Characters, not bytes: Cyrillic letters take two bytes each, and the
	// columns are VARCHAR(255) and VARCHAR(1000) characters.
	if utf8.RuneCountInString(f.TitleEN) > 255 {
		return f, fmt.Errorf("English title must not exceed 255 characters")
	}
	if utf8.RuneCountInString(f.Description) > maxDescription {
		return f, fmt.Errorf("description must not exceed %d characters", maxDescription)
	}
	if utf8.RuneCountInString(f.DescriptionEN) > maxDescription {
		return f, fmt.Errorf("English description must not exceed %d characters", maxDescription)
	}
	category, err := normalizeCategory(f.Category)
	if err != nil {
		return f, err
	}
	f.Category = category
	return f, nil
}

type ProjectService struct {
	repo projectRepo
}

func NewProjectService(repo projectRepo) *ProjectService {
	return &ProjectService{repo: repo}
}

func (s *ProjectService) List(ctx context.Context) ([]model.Project, error) {
	return s.repo.List(ctx)
}

func (s *ProjectService) GetByID(ctx context.Context, id uuid.UUID) (*model.Project, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *ProjectService) Create(ctx context.Context, f model.ProjectFields) (*model.Project, error) {
	f, err := prepareProject(f)
	if err != nil {
		return nil, err
	}
	return s.repo.Create(ctx, f)
}

func (s *ProjectService) Update(ctx context.Context, id uuid.UUID, f model.ProjectFields, coverMediaID *uuid.UUID) (*model.Project, error) {
	f, err := prepareProject(f)
	if err != nil {
		return nil, err
	}
	return s.repo.Update(ctx, id, f, coverMediaID)
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

// ReorderProjects sets the order of one category's projects on the public site.
// The list must contain every project of the category exactly once.
func (s *ProjectService) ReorderProjects(ctx context.Context, category string, projectIDs []uuid.UUID) error {
	if category == "" {
		return fmt.Errorf("category is required")
	}
	category, err := normalizeCategory(category)
	if err != nil {
		return err
	}
	projects, err := s.repo.List(ctx)
	if err != nil {
		return err
	}
	inCategory := make(map[uuid.UUID]bool)
	for _, p := range projects {
		if p.Category == category {
			inCategory[p.ID] = true
		}
	}
	if len(projectIDs) != len(inCategory) {
		return fmt.Errorf("the order must list all %d projects of the category", len(inCategory))
	}
	seen := make(map[uuid.UUID]bool, len(projectIDs))
	for _, id := range projectIDs {
		if !inCategory[id] {
			return fmt.Errorf("project %s is not in category %s", id, category)
		}
		if seen[id] {
			return fmt.Errorf("project %s is listed twice", id)
		}
		seen[id] = true
	}
	return s.repo.ReorderProjects(ctx, category, projectIDs)
}

// ReorderMedia puts the project's photos in the given order. The list must
// contain every photo of the project exactly once.
func (s *ProjectService) ReorderMedia(ctx context.Context, projectID uuid.UUID, mediaIDs []uuid.UUID) error {
	project, err := s.repo.GetByID(ctx, projectID)
	if err != nil {
		return fmt.Errorf("project not found: %w", err)
	}
	if len(mediaIDs) != len(project.Media) {
		return fmt.Errorf("the order must list all %d photos of the project", len(project.Media))
	}
	owned := make(map[uuid.UUID]bool, len(project.Media))
	for _, m := range project.Media {
		owned[m.ID] = true
	}
	seen := make(map[uuid.UUID]bool, len(mediaIDs))
	for _, id := range mediaIDs {
		if !owned[id] {
			return fmt.Errorf("photo %s does not belong to this project", id)
		}
		if seen[id] {
			return fmt.Errorf("photo %s is listed twice", id)
		}
		seen[id] = true
	}
	return s.repo.ReorderMedia(ctx, projectID, mediaIDs)
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
