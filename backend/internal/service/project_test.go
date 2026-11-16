package service

import (
	"context"
	"strings"
	"testing"
	"time"

	"design-profile/backend/internal/model"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// ── Mock: projectRepo ────────────────────────────────────────────────────────

type mockProjectRepo struct{ mock.Mock }

func (m *mockProjectRepo) List(ctx context.Context) ([]model.Project, error) {
	args := m.Called(ctx)
	return args.Get(0).([]model.Project), args.Error(1)
}

func (m *mockProjectRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.Project, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Project), args.Error(1)
}

func (m *mockProjectRepo) Create(ctx context.Context, title, description, category string) (*model.Project, error) {
	args := m.Called(ctx, title, description, category)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Project), args.Error(1)
}

func (m *mockProjectRepo) Update(ctx context.Context, id uuid.UUID, title, description, category string, coverMediaID *uuid.UUID) (*model.Project, error) {
	args := m.Called(ctx, id, title, description, category, coverMediaID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Project), args.Error(1)
}

func (m *mockProjectRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

func (m *mockProjectRepo) AddMedia(ctx context.Context, projectID uuid.UUID, data []byte, contentType, filename string, sortOrder int) (*model.Media, error) {
	args := m.Called(ctx, projectID, data, contentType, filename, sortOrder)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Media), args.Error(1)
}

func (m *mockProjectRepo) GetMediaData(ctx context.Context, mediaID uuid.UUID) ([]byte, string, error) {
	args := m.Called(ctx, mediaID)
	return args.Get(0).([]byte), args.String(1), args.Error(2)
}

func (m *mockProjectRepo) DeleteMedia(ctx context.Context, mediaID uuid.UUID) error {
	return m.Called(ctx, mediaID).Error(0)
}

// ── Create validation ────────────────────────────────────────────────────────

func TestCreate_EmptyTitle_ReturnsError(t *testing.T) {
	svc := NewProjectService(&mockProjectRepo{})
	_, err := svc.Create(context.Background(), "", "some description", "")
	assert.Error(t, err)
}

func TestCreate_DescriptionTooLong_ReturnsError(t *testing.T) {
	svc := NewProjectService(&mockProjectRepo{})
	_, err := svc.Create(context.Background(), "Title", strings.Repeat("x", 501), "")
	assert.Error(t, err)
}

func TestCreate_DescriptionExactly500Chars_Allowed(t *testing.T) {
	repo := &mockProjectRepo{}
	svc := NewProjectService(repo)
	ctx := context.Background()
	desc := strings.Repeat("a", 500)

	repo.On("Create", ctx, "Title", desc, "residential").Return(&model.Project{ID: uuid.New()}, nil)

	_, err := svc.Create(ctx, "Title", desc, "")
	require.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestCreate_ValidInput_CallsRepo(t *testing.T) {
	repo := &mockProjectRepo{}
	svc := NewProjectService(repo)
	ctx := context.Background()

	expected := &model.Project{
		ID:          uuid.New(),
		Title:       "My project",
		Description: "Nice desc",
		CreatedAt:   time.Now(),
	}
	repo.On("Create", ctx, "My project", "Nice desc", "commercial").Return(expected, nil)

	got, err := svc.Create(ctx, "My project", "Nice desc", "commercial")

	require.NoError(t, err)
	assert.Equal(t, expected, got)
	repo.AssertExpectations(t)
}

// ── Update validation ────────────────────────────────────────────────────────

func TestUpdate_EmptyTitle_ReturnsError(t *testing.T) {
	svc := NewProjectService(&mockProjectRepo{})
	_, err := svc.Update(context.Background(), uuid.New(), "", "desc", "", nil)
	assert.Error(t, err)
}

func TestUpdate_DescriptionTooLong_ReturnsError(t *testing.T) {
	svc := NewProjectService(&mockProjectRepo{})
	_, err := svc.Update(context.Background(), uuid.New(), "Title", strings.Repeat("y", 501), "", nil)
	assert.Error(t, err)
}

func TestUpdate_ValidInput_CallsRepo(t *testing.T) {
	repo := &mockProjectRepo{}
	svc := NewProjectService(repo)
	ctx := context.Background()
	id := uuid.New()

	expected := &model.Project{ID: id, Title: "Updated"}
	repo.On("Update", ctx, id, "Updated", "desc", "commercial", (*uuid.UUID)(nil)).Return(expected, nil)

	got, err := svc.Update(ctx, id, "Updated", "desc", "commercial", nil)

	require.NoError(t, err)
	assert.Equal(t, expected, got)
	repo.AssertExpectations(t)
}

func TestCreate_Cyrillic500Chars_Allowed(t *testing.T) {
	repo := &mockProjectRepo{}
	svc := NewProjectService(repo)
	ctx := context.Background()
	desc := strings.Repeat("ж", 500) // 500 characters, 1000 bytes

	repo.On("Create", ctx, "Title", desc, "residential").Return(&model.Project{ID: uuid.New()}, nil)

	_, err := svc.Create(ctx, "Title", desc, "")
	require.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestCreate_Cyrillic501Chars_ReturnsError(t *testing.T) {
	svc := NewProjectService(&mockProjectRepo{})
	_, err := svc.Create(context.Background(), "Title", strings.Repeat("ж", 501), "")
	assert.Error(t, err)
}

// ── Category ─────────────────────────────────────────────────────────────────

func TestCreate_UnknownCategory_ReturnsError(t *testing.T) {
	svc := NewProjectService(&mockProjectRepo{})
	_, err := svc.Create(context.Background(), "Title", "desc", "office")
	assert.Error(t, err)
}

func TestUpdate_UnknownCategory_ReturnsError(t *testing.T) {
	svc := NewProjectService(&mockProjectRepo{})
	_, err := svc.Update(context.Background(), uuid.New(), "Title", "desc", "Residential", nil)
	assert.Error(t, err, "categories are case-sensitive")
}

func TestUpdate_EmptyCategory_DefaultsToResidential(t *testing.T) {
	repo := &mockProjectRepo{}
	svc := NewProjectService(repo)
	ctx := context.Background()
	id := uuid.New()

	repo.On("Update", ctx, id, "Title", "desc", "residential", (*uuid.UUID)(nil)).Return(&model.Project{ID: id}, nil)

	_, err := svc.Update(ctx, id, "Title", "desc", "", nil)
	require.NoError(t, err)
	repo.AssertExpectations(t)
}

// ── AddMedia validation ──────────────────────────────────────────────────────

func TestAddMedia_UnsupportedType_ReturnsError(t *testing.T) {
	svc := NewProjectService(&mockProjectRepo{})
	ctx := context.Background()

	for _, ct := range []string{"application/pdf", "text/plain", "video/mp4", "application/octet-stream"} {
		t.Run(ct, func(t *testing.T) {
			_, err := svc.AddMedia(ctx, uuid.New(), []byte("data"), ct, "file", 0)
			assert.Error(t, err, "expected error for content type: %s", ct)
		})
	}
}

func TestAddMedia_AllowedTypes_PassValidation(t *testing.T) {
	repo := &mockProjectRepo{}
	svc := NewProjectService(repo)
	ctx := context.Background()
	projectID := uuid.New()
	fakeMedia := &model.Media{ID: uuid.New()}

	allowed := []string{"image/jpeg", "image/png", "image/gif", "image/webp", "image/svg+xml"}
	for _, ct := range allowed {
		ct := ct
		t.Run(ct, func(t *testing.T) {
			repo.On("AddMedia", ctx, projectID, []byte("img"), ct, "file.jpg", 0).Return(fakeMedia, nil)

			got, err := svc.AddMedia(ctx, projectID, []byte("img"), ct, "file.jpg", 0)
			require.NoError(t, err)
			assert.Equal(t, fakeMedia, got)
		})
	}
	repo.AssertExpectations(t)
}
