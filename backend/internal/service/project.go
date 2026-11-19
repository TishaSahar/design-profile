package service

import (
	"context"
	"fmt"
	"log/slog"
	"mime"
	"path/filepath"
	"unicode/utf8"

	"design-profile/backend/internal/imageproc"
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
	GetMediaData(ctx context.Context, mediaID uuid.UUID, size string) (*model.MediaFile, error)
	SetMediaCopies(ctx context.Context, mediaID uuid.UUID, small, large []byte) error
	MediaWithoutCopies(ctx context.Context) ([]uuid.UUID, error)
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
	m, err := s.repo.AddMedia(ctx, projectID, data, contentType, filename, sortOrder)
	if err != nil {
		return nil, err
	}
	// The photo is saved; without copies the site shows the original until the
	// next startup makes them, so a failure here does not fail the upload.
	s.makeCopies(ctx, m.ID, data)
	return m, nil
}

// Photo sizes for GetMediaData: the uploaded original or one of its web copies.
const (
	MediaOriginal = ""
	MediaSmall    = "small"
	MediaLarge    = "large"
)

// ValidMediaSize reports whether size is one of the photo sizes above.
func ValidMediaSize(size string) bool {
	return size == MediaOriginal || size == MediaSmall || size == MediaLarge
}

// GetMediaData returns a photo in the given size (see MediaOriginal and the others).
func (s *ProjectService) GetMediaData(ctx context.Context, mediaID uuid.UUID, size string) (*model.MediaFile, error) {
	if !ValidMediaSize(size) {
		return nil, fmt.Errorf("unknown photo size: %s", size)
	}
	return s.repo.GetMediaData(ctx, mediaID, size)
}

// makeCopies stores the web copies of a photo. A photo they cannot be made of
// (GIF, SVG, a broken file) is marked as having none, so it is not retried.
func (s *ProjectService) makeCopies(ctx context.Context, mediaID uuid.UUID, data []byte) {
	small, large, err := imageproc.Copies(data)
	if err != nil {
		slog.Info("photo is served without web copies", "media_id", mediaID, "reason", err)
		small, large = nil, nil
	}
	if err := s.repo.SetMediaCopies(ctx, mediaID, small, large); err != nil {
		slog.Error("store photo copies", "media_id", mediaID, "err", err)
	}
}

// MakeMissingCopies makes the web copies of photos uploaded before copies existed,
// or whose copies failed to save. It runs once on startup, one photo at a time to
// keep memory use low; the site serves the originals of photos still waiting.
func (s *ProjectService) MakeMissingCopies(ctx context.Context) {
	ids, err := s.repo.MediaWithoutCopies(ctx)
	if err != nil {
		slog.Error("find photos without web copies", "err", err)
		return
	}
	if len(ids) == 0 {
		return
	}
	slog.Info("making web copies of photos", "count", len(ids))
	for _, id := range ids {
		if ctx.Err() != nil {
			return
		}
		f, err := s.repo.GetMediaData(ctx, id, MediaOriginal)
		if err != nil {
			slog.Error("read photo for web copies", "media_id", id, "err", err)
			continue
		}
		s.makeCopies(ctx, id, f.Data)
	}
	slog.Info("web copies of photos are ready", "count", len(ids))
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
