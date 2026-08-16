package service

import (
	"context"
	"fmt"

	"design-profile/backend/internal/model"
	"design-profile/backend/internal/repository"

	"github.com/google/uuid"
)

const (
	maxPhotoAttachments = 10
	maxAttachmentSize   = 20 * 1024 * 1024 // 20 MB per file
)

type RequestService struct {
	repo *repository.RequestRepository
}

func NewRequestService(repo *repository.RequestRepository) *RequestService {
	return &RequestService{repo: repo}
}

type CreateRequestInput struct {
	FirstName   string
	LastName    string
	Contact     string
	Description string
	Consented   bool
	Attachments []AttachmentInput
}

type AttachmentInput struct {
	Data        []byte
	ContentType string
	Filename    string
}

// Create validates and stores a new project request with attachments.
func (s *RequestService) Create(ctx context.Context, input CreateRequestInput) (*model.ProjectRequest, error) {
	if !input.Consented {
		return nil, fmt.Errorf("consent to personal data processing is required")
	}
	if input.FirstName == "" || input.LastName == "" {
		return nil, fmt.Errorf("first and last name are required")
	}
	if input.Contact == "" {
		return nil, fmt.Errorf("contact information is required")
	}
	if input.Description == "" {
		return nil, fmt.Errorf("project description is required")
	}

	if err := validateAttachments(input.Attachments); err != nil {
		return nil, err
	}

	req, err := s.repo.Create(ctx, &model.ProjectRequest{
		FirstName:   input.FirstName,
		LastName:    input.LastName,
		Contact:     input.Contact,
		Description: input.Description,
		Consented:   input.Consented,
	})
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	for _, a := range input.Attachments {
		att, err := s.repo.AddAttachment(ctx, req.ID, a.Data, a.ContentType, a.Filename)
		if err != nil {
			return nil, fmt.Errorf("save attachment: %w", err)
		}
		req.Attachments = append(req.Attachments, *att)
	}

	return req, nil
}

func (s *RequestService) List(ctx context.Context) ([]model.ProjectRequest, error) {
	return s.repo.List(ctx)
}

func (s *RequestService) GetByID(ctx context.Context, id uuid.UUID) (*model.ProjectRequest, error) {
	return s.repo.GetByID(ctx, id)
}

// GetAttachmentData returns binary data for a request attachment.
func (s *RequestService) GetAttachmentData(ctx context.Context, attachmentID uuid.UUID) ([]byte, string, error) {
	return s.repo.GetAttachmentData(ctx, attachmentID)
}

func validateAttachments(attachments []AttachmentInput) error {
	hasPDF := false
	photoCount := 0

	for _, a := range attachments {
		if len(a.Data) > maxAttachmentSize {
			return fmt.Errorf("file %s exceeds the 20 MB size limit", a.Filename)
		}
		switch a.ContentType {
		case "application/pdf":
			hasPDF = true
		case "image/jpeg", "image/png", "image/webp", "image/gif":
			photoCount++
		default:
			return fmt.Errorf("unsupported file type: %s (allowed: JPEG, PNG, WEBP, GIF, PDF)", a.ContentType)
		}
	}

	if hasPDF && photoCount > 0 {
		return fmt.Errorf("you may attach either up to 10 photos or a single PDF, not both")
	}
	if hasPDF && len(attachments) > 1 {
		return fmt.Errorf("only one PDF file is allowed")
	}
	if photoCount > maxPhotoAttachments {
		return fmt.Errorf("maximum %d photos allowed", maxPhotoAttachments)
	}
	return nil
}
