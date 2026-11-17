package repository

import (
	"context"
	"fmt"

	"design-profile/backend/internal/model"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

const projectColumns = `id, title, description, category, cover_media_id, sort_order, created_at, updated_at`

// projectScanTargets returns scan destinations matching projectColumns.
func projectScanTargets(p *model.Project) []any {
	return []any{&p.ID, &p.Title, &p.Description, &p.Category, &p.CoverMediaID, &p.SortOrder, &p.CreatedAt, &p.UpdatedAt}
}

type ProjectRepository struct {
	db *pgxpool.Pool
}

func NewProjectRepository(db *pgxpool.Pool) *ProjectRepository {
	return &ProjectRepository{db: db}
}

// List returns all projects without media data.
func (r *ProjectRepository) List(ctx context.Context) ([]model.Project, error) {
	rows, err := r.db.Query(ctx,
		// Residential first, then HoReCa; inside a category, in the order set in the admin panel.
		`SELECT `+projectColumns+` FROM projects
		 ORDER BY CASE category WHEN 'residential' THEN 0 ELSE 1 END, category, sort_order, created_at DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("list projects: %w", err)
	}
	defer rows.Close()

	var projects []model.Project
	for rows.Next() {
		var p model.Project
		if err := rows.Scan(projectScanTargets(&p)...); err != nil {
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
		`SELECT `+projectColumns+` FROM projects WHERE id = $1`,
		id,
	).Scan(projectScanTargets(p)...)
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

// firstInCategory is a sort_order that puts a project before every other one of category $N.
func firstInCategory(param string) string {
	return `(SELECT COALESCE(MIN(sort_order), 0) - 1 FROM projects WHERE category = ` + param + `::varchar)`
}

// Create inserts a new project at the top of its category.
func (r *ProjectRepository) Create(ctx context.Context, title, description, category string) (*model.Project, error) {
	p := &model.Project{}
	err := r.db.QueryRow(ctx,
		`INSERT INTO projects (title, description, category, sort_order)
		 VALUES ($1, $2, $3::varchar, `+firstInCategory("$3")+`)
		 RETURNING `+projectColumns,
		title, description, category,
	).Scan(projectScanTargets(p)...)
	if err != nil {
		return nil, fmt.Errorf("create project: %w", err)
	}
	return p, nil
}

// Update modifies a project's title, description, category and cover.
// A project moved to another category goes to the top of it.
func (r *ProjectRepository) Update(ctx context.Context, id uuid.UUID, title, description, category string, coverMediaID *uuid.UUID) (*model.Project, error) {
	p := &model.Project{}
	err := r.db.QueryRow(ctx,
		`UPDATE projects SET title = $2, description = $3, category = $4::varchar, cover_media_id = $5,
		        sort_order = CASE WHEN category = $4::varchar THEN sort_order ELSE `+firstInCategory("$4")+` END,
		        updated_at = NOW()
		 WHERE id = $1
		 RETURNING `+projectColumns,
		id, title, description, category, coverMediaID,
	).Scan(projectScanTargets(p)...)
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
// A negative sortOrder appends the file after the project's existing media.
func (r *ProjectRepository) AddMedia(ctx context.Context, projectID uuid.UUID, data []byte, contentType, filename string, sortOrder int) (*model.Media, error) {
	m := &model.Media{}
	err := r.db.QueryRow(ctx,
		`INSERT INTO project_media (project_id, data, content_type, filename, sort_order)
		 VALUES ($1, $2, $3, $4,
		         CASE WHEN $5::int >= 0 THEN $5::int
		              ELSE (SELECT COALESCE(MAX(sort_order) + 1, 0) FROM project_media WHERE project_id = $1)
		         END)
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

// ReorderMedia sets sort_order of the project's media to their position in mediaIDs.
func (r *ProjectRepository) ReorderMedia(ctx context.Context, projectID uuid.UUID, mediaIDs []uuid.UUID) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("reorder media: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck // no-op after Commit

	for i, id := range mediaIDs {
		tag, err := tx.Exec(ctx,
			`UPDATE project_media SET sort_order = $1 WHERE id = $2 AND project_id = $3`,
			i, id, projectID,
		)
		if err != nil {
			return fmt.Errorf("reorder media: %w", err)
		}
		if tag.RowsAffected() != 1 {
			return fmt.Errorf("reorder media: media %s does not belong to project %s", id, projectID)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("reorder media: %w", err)
	}
	return nil
}

// ReorderProjects sets sort_order of the category's projects to their position in projectIDs.
func (r *ProjectRepository) ReorderProjects(ctx context.Context, category string, projectIDs []uuid.UUID) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("reorder projects: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck // no-op after Commit

	for i, id := range projectIDs {
		tag, err := tx.Exec(ctx,
			`UPDATE projects SET sort_order = $1 WHERE id = $2 AND category = $3`,
			i, id, category,
		)
		if err != nil {
			return fmt.Errorf("reorder projects: %w", err)
		}
		if tag.RowsAffected() != 1 {
			return fmt.Errorf("reorder projects: project %s is not in category %s", id, category)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("reorder projects: %w", err)
	}
	return nil
}

// DeleteMedia removes a single media file by ID.
func (r *ProjectRepository) DeleteMedia(ctx context.Context, mediaID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM project_media WHERE id = $1`, mediaID)
	if err != nil {
		return fmt.Errorf("delete media: %w", err)
	}
	return nil
}
