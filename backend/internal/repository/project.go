package repository

import (
	"context"
	"fmt"

	"design-profile/backend/internal/model"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProjectRepository struct {
	db *pgxpool.Pool
}

func NewProjectRepository(db *pgxpool.Pool) *ProjectRepository {
	return &ProjectRepository{db: db}
}

// List returns all projects without media data.
func (r *ProjectRepository) List(ctx context.Context) ([]model.Project, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, title, description, cover_media_id, created_at, updated_at
		 FROM projects ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("list projects: %w", err)
	}
	defer rows.Close()

	var projects []model.Project
	for rows.Next() {
		var p model.Project
		if err := rows.Scan(&p.ID, &p.Title, &p.Description, &p.CoverMediaID, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan project: %w", err)
		}
		projects = append(projects, p)
	}
	return projects, rows.Err()
}

// GetByID returns a project with all its media metadata (no binary data).
func (r *ProjectRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Project, error) {
	p := &model.Project{}
	err := r.db.QueryRow(ctx,
		`SELECT id, title, description, cover_media_id, created_at, updated_at
		 FROM projects WHERE id = $1`,
		id,
	).Scan(&p.ID, &p.Title, &p.Description, &p.CoverMediaID, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get project: %w", err)
	}

	media, err := r.ListMedia(ctx, id)
	if err != nil {
		return nil, err
	}
	p.Media = media
	return p, nil
}

// Create inserts a new project.
func (r *ProjectRepository) Create(ctx context.Context, title, description string) (*model.Project, error) {
	p := &model.Project{}
	err := r.db.QueryRow(ctx,
		`INSERT INTO projects (title, description) VALUES ($1, $2)
		 RETURNING id, title, description, cover_media_id, created_at, updated_at`,
		title, description,
	).Scan(&p.ID, &p.Title, &p.Description, &p.CoverMediaID, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create project: %w", err)
	}
	return p, nil
}

// Update modifies a project's title, description and cover.
func (r *ProjectRepository) Update(ctx context.Context, id uuid.UUID, title, description string, coverMediaID *uuid.UUID) (*model.Project, error) {
	p := &model.Project{}
	err := r.db.QueryRow(ctx,
		`UPDATE projects SET title = $2, description = $3, cover_media_id = $4, updated_at = NOW()
		 WHERE id = $1
		 RETURNING id, title, description, cover_media_id, created_at, updated_at`,
		id, title, description, coverMediaID,
	).Scan(&p.ID, &p.Title, &p.Description, &p.CoverMediaID, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("update project: %w", err)
	}
	return p, nil
}

// Delete removes a project and all its media (cascade).
func (r *ProjectRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM projects WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete project: %w", err)
	}
	return nil
}

// AddMedia inserts a media file for a project and returns metadata (without binary data).
func (r *ProjectRepository) AddMedia(ctx context.Context, projectID uuid.UUID, data []byte, contentType, filename string, sortOrder int) (*model.Media, error) {
	m := &model.Media{}
	err := r.db.QueryRow(ctx,
		`INSERT INTO project_media (project_id, data, content_type, filename, sort_order)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, project_id, content_type, filename, sort_order, created_at`,
		projectID, data, contentType, filename, sortOrder,
	).Scan(&m.ID, &m.ProjectID, &m.ContentType, &m.Filename, &m.SortOrder, &m.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("add media: %w", err)
	}
	return m, nil
}

// GetMediaData returns the binary data and content type for a media file.
func (r *ProjectRepository) GetMediaData(ctx context.Context, mediaID uuid.UUID) ([]byte, string, error) {
	var data []byte
	var contentType string
	err := r.db.QueryRow(ctx,
		`SELECT data, content_type FROM project_media WHERE id = $1`,
		mediaID,
	).Scan(&data, &contentType)
	if err != nil {
		return nil, "", fmt.Errorf("get media data: %w", err)
	}
	return data, contentType, nil
}

// ListMedia returns media metadata for a project (no binary data).
func (r *ProjectRepository) ListMedia(ctx context.Context, projectID uuid.UUID) ([]model.Media, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, project_id, content_type, filename, sort_order, created_at
		 FROM project_media WHERE project_id = $1 ORDER BY sort_order ASC`,
		projectID,
	)
	if err != nil {
		return nil, fmt.Errorf("list media: %w", err)
	}
	defer rows.Close()

	var media []model.Media
	for rows.Next() {
		var m model.Media
		if err := rows.Scan(&m.ID, &m.ProjectID, &m.ContentType, &m.Filename, &m.SortOrder, &m.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan media: %w", err)
		}
		media = append(media, m)
	}
	return media, rows.Err()
}

// DeleteMedia removes a single media file by ID.
func (r *ProjectRepository) DeleteMedia(ctx context.Context, mediaID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM project_media WHERE id = $1`, mediaID)
	if err != nil {
		return fmt.Errorf("delete media: %w", err)
	}
	return nil
}
