package repository

import (
	"context"
	"fmt"

	"design-profile/backend/internal/model"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

const requestColumns = `id, first_name, last_name, contact, email, phone, telegram,
	preferred_contact, description, consented, created_at`

// requestScanTargets returns scan destinations matching requestColumns.
func requestScanTargets(r *model.ProjectRequest) []any {
	return []any{&r.ID, &r.FirstName, &r.LastName, &r.Contact, &r.Email, &r.Phone, &r.Telegram,
		&r.PreferredContact, &r.Description, &r.Consented, &r.CreatedAt}
}

type RequestRepository struct {
	db *pgxpool.Pool
}

func NewRequestRepository(db *pgxpool.Pool) *RequestRepository {
	return &RequestRepository{db: db}
}

// Create stores a new project request with its attachments and the tasks it
// queues (such as the admin notification) in one transaction: either all of
// them are saved, or none — a notification never goes out for a request that
// was not saved, and a saved request is never left without its notification.
func (r *RequestRepository) Create(ctx context.Context, req *model.ProjectRequest, attachments []model.Attachment, tasks []model.NewTask) (*model.ProjectRequest, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin request transaction: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck // a no-op after Commit

	result := &model.ProjectRequest{}
	err = tx.QueryRow(ctx,
		`INSERT INTO project_requests
		   (id, first_name, last_name, email, phone, telegram, preferred_contact, description, consented)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		 RETURNING `+requestColumns,
		req.ID, req.FirstName, req.LastName, req.Email, req.Phone, req.Telegram,
		req.PreferredContact, req.Description, req.Consented,
	).Scan(requestScanTargets(result)...)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	for _, a := range attachments {
		att := model.Attachment{}
		err := tx.QueryRow(ctx,
			`INSERT INTO request_attachments (request_id, data, content_type, filename)
			 VALUES ($1, $2, $3, $4)
			 RETURNING id, request_id, content_type, filename, created_at`,
			result.ID, a.Data, a.ContentType, a.Filename,
		).Scan(&att.ID, &att.RequestID, &att.ContentType, &att.Filename, &att.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("add attachment: %w", err)
		}
		result.Attachments = append(result.Attachments, att)
	}

	if err := insertTasks(ctx, tx, tasks); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit request: %w", err)
	}
	return result, nil
}

// List returns all project requests without binary attachment data.
func (r *RequestRepository) List(ctx context.Context) ([]model.ProjectRequest, error) {
	rows, err := r.db.Query(ctx,
		`SELECT `+requestColumns+` FROM project_requests ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("list requests: %w", err)
	}
	defer rows.Close()

	var requests []model.ProjectRequest
	for rows.Next() {
		var req model.ProjectRequest
		if err := rows.Scan(requestScanTargets(&req)...); err != nil {
			return nil, fmt.Errorf("scan request: %w", err)
		}
		requests = append(requests, req)
	}
	return requests, rows.Err()
}

// GetByID returns a single project request with attachment metadata.
func (r *RequestRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.ProjectRequest, error) {
	req := &model.ProjectRequest{}
	err := r.db.QueryRow(ctx,
		`SELECT `+requestColumns+` FROM project_requests WHERE id = $1`,
		id,
	).Scan(requestScanTargets(req)...)
	if err != nil {
		return nil, fmt.Errorf("get request: %w", err)
	}

	rows, err := r.db.Query(ctx,
		`SELECT id, request_id, content_type, filename, created_at
		 FROM request_attachments WHERE request_id = $1 ORDER BY created_at ASC`,
		id,
	)
	if err != nil {
		return nil, fmt.Errorf("list attachments: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var a model.Attachment
		if err := rows.Scan(&a.ID, &a.RequestID, &a.ContentType, &a.Filename, &a.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan attachment: %w", err)
		}
		req.Attachments = append(req.Attachments, a)
	}
	return req, rows.Err()
}

// GetAttachmentData returns binary data for an attachment by ID.
func (r *RequestRepository) GetAttachmentData(ctx context.Context, attachmentID uuid.UUID) ([]byte, string, error) {
	var data []byte
	var contentType string
	err := r.db.QueryRow(ctx,
		`SELECT data, content_type FROM request_attachments WHERE id = $1`,
		attachmentID,
	).Scan(&data, &contentType)
	if err != nil {
		return nil, "", fmt.Errorf("get attachment data: %w", err)
	}
	return data, contentType, nil
}
