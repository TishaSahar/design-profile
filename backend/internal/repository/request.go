package repository

import (
	"context"
	"fmt"

	"design-profile/backend/internal/model"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RequestRepository struct {
	db *pgxpool.Pool
}

func NewRequestRepository(db *pgxpool.Pool) *RequestRepository {
	return &RequestRepository{db: db}
}

// Create inserts a new project request.
func (r *RequestRepository) Create(ctx context.Context, req *model.ProjectRequest) (*model.ProjectRequest, error) {
	result := &model.ProjectRequest{}
	err := r.db.QueryRow(ctx,
		`INSERT INTO project_requests (first_name, last_name, contact, description, consented)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, first_name, last_name, contact, description, consented, created_at`,
		req.FirstName, req.LastName, req.Contact, req.Description, req.Consented,
	).Scan(&result.ID, &result.FirstName, &result.LastName, &result.Contact,
		&result.Description, &result.Consented, &result.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	return result, nil
}

// AddAttachment inserts a binary attachment for a project request.
func (r *RequestRepository) AddAttachment(ctx context.Context, requestID uuid.UUID, data []byte, contentType, filename string) (*model.Attachment, error) {
	a := &model.Attachment{}
	err := r.db.QueryRow(ctx,
		`INSERT INTO request_attachments (request_id, data, content_type, filename)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, request_id, content_type, filename, created_at`,
		requestID, data, contentType, filename,
	).Scan(&a.ID, &a.RequestID, &a.ContentType, &a.Filename, &a.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("add attachment: %w", err)
	}
	return a, nil
}

// List returns all project requests without binary attachment data.
func (r *RequestRepository) List(ctx context.Context) ([]model.ProjectRequest, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, first_name, last_name, contact, description, consented, created_at
		 FROM project_requests ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("list requests: %w", err)
	}
	defer rows.Close()

	var requests []model.ProjectRequest
	for rows.Next() {
		var req model.ProjectRequest
		if err := rows.Scan(&req.ID, &req.FirstName, &req.LastName, &req.Contact,
			&req.Description, &req.Consented, &req.CreatedAt); err != nil {
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
		`SELECT id, first_name, last_name, contact, description, consented, created_at
		 FROM project_requests WHERE id = $1`,
		id,
	).Scan(&req.ID, &req.FirstName, &req.LastName, &req.Contact,
		&req.Description, &req.Consented, &req.CreatedAt)
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
