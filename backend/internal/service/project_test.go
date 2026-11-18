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

func (m *mockProjectRepo) Create(ctx context.Context, f model.ProjectFields) (*model.Project, error) {
	args := m.Called(ctx, f)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Project), args.Error(1)
}

func (m *mockProjectRepo) Update(ctx context.Context, id uuid.UUID, f model.ProjectFields, coverMediaID *uuid.UUID) (*model.Project, error) {
	args := m.Called(ctx, id, f, coverMediaID)
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

func (m *mockProjectRepo) ReorderProjects(ctx context.Context, category string, projectIDs []uuid.UUID) error {
	return m.Called(ctx, category, projectIDs).Error(0)
}

func (m *mockProjectRepo) ReorderMedia(ctx context.Context, projectID uuid.UUID, mediaIDs []uuid.UUID) error {
	return m.Called(ctx, projectID, mediaIDs).Error(0)
}

// fields builds ProjectFields with Russian texts only.
func fields(title, description, category string) model.ProjectFields {
	return model.ProjectFields{Title: title, Description: description, Category: category}
}

// ── Create validation ────────────────────────────────────────────────────────

func TestCreate_EmptyTitle_ReturnsError(t *testing.T) {
	svc := NewProjectService(&mockProjectRepo{})
	_, err := svc.Create(context.Background(), fields("", "some description", ""))
	assert.Error(t, err)
}

func TestCreate_DescriptionTooLong_ReturnsError(t *testing.T) {
	svc := NewProjectService(&mockProjectRepo{})
	_, err := svc.Create(context.Background(), fields("Title", strings.Repeat("x", 501), ""))
	assert.Error(t, err)
}

func TestCreate_DescriptionExactly500Chars_Allowed(t *testing.T) {
	repo := &mockProjectRepo{}
	svc := NewProjectService(repo)
	ctx := context.Background()
	desc := strings.Repeat("a", 500)

	repo.On("Create", ctx, fields("Title", desc, "residential")).Return(&model.Project{ID: uuid.New()}, nil)

	_, err := svc.Create(ctx, fields("Title", desc, ""))
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
	repo.On("Create", ctx, fields("My project", "Nice desc", "commercial")).Return(expected, nil)

	got, err := svc.Create(ctx, fields("My project", "Nice desc", "commercial"))

	require.NoError(t, err)
	assert.Equal(t, expected, got)
	repo.AssertExpectations(t)
}

// ── Update validation ────────────────────────────────────────────────────────

func TestUpdate_EmptyTitle_ReturnsError(t *testing.T) {
	svc := NewProjectService(&mockProjectRepo{})
	_, err := svc.Update(context.Background(), uuid.New(), fields("", "desc", ""), nil)
	assert.Error(t, err)
}

func TestUpdate_DescriptionTooLong_ReturnsError(t *testing.T) {
	svc := NewProjectService(&mockProjectRepo{})
	_, err := svc.Update(context.Background(), uuid.New(), fields("Title", strings.Repeat("y", 501), ""), nil)
	assert.Error(t, err)
}

func TestUpdate_ValidInput_CallsRepo(t *testing.T) {
	repo := &mockProjectRepo{}
	svc := NewProjectService(repo)
	ctx := context.Background()
	id := uuid.New()

	expected := &model.Project{ID: id, Title: "Updated"}
	repo.On("Update", ctx, id, fields("Updated", "desc", "commercial"), (*uuid.UUID)(nil)).Return(expected, nil)

	got, err := svc.Update(ctx, id, fields("Updated", "desc", "commercial"), nil)

	require.NoError(t, err)
	assert.Equal(t, expected, got)
	repo.AssertExpectations(t)
}

func TestCreate_Cyrillic500Chars_Allowed(t *testing.T) {
	repo := &mockProjectRepo{}
	svc := NewProjectService(repo)
	ctx := context.Background()
	desc := strings.Repeat("ж", 500) // 500 characters, 1000 bytes

	repo.On("Create", ctx, fields("Title", desc, "residential")).Return(&model.Project{ID: uuid.New()}, nil)

	_, err := svc.Create(ctx, fields("Title", desc, ""))
	require.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestCreate_Cyrillic501Chars_ReturnsError(t *testing.T) {
	svc := NewProjectService(&mockProjectRepo{})
	_, err := svc.Create(context.Background(), fields("Title", strings.Repeat("ж", 501), ""))
	assert.Error(t, err)
}

// ── Category ─────────────────────────────────────────────────────────────────

func TestCreate_UnknownCategory_ReturnsError(t *testing.T) {
	svc := NewProjectService(&mockProjectRepo{})
	_, err := svc.Create(context.Background(), fields("Title", "desc", "office"))
	assert.Error(t, err)
}

func TestUpdate_UnknownCategory_ReturnsError(t *testing.T) {
	svc := NewProjectService(&mockProjectRepo{})
	_, err := svc.Update(context.Background(), uuid.New(), fields("Title", "desc", "Residential"), nil)
	assert.Error(t, err, "categories are case-sensitive")
}

func TestUpdate_EmptyCategory_DefaultsToResidential(t *testing.T) {
	repo := &mockProjectRepo{}
	svc := NewProjectService(repo)
	ctx := context.Background()
	id := uuid.New()

	repo.On("Update", ctx, id, fields("Title", "desc", "residential"), (*uuid.UUID)(nil)).Return(&model.Project{ID: id}, nil)

	_, err := svc.Update(ctx, id, fields("Title", "desc", ""), nil)
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

// ── ReorderMedia ─────────────────────────────────────────────────────────────

func projectWithMedia(n int) (*model.Project, []uuid.UUID) {
	p := &model.Project{ID: uuid.New()}
	ids := make([]uuid.UUID, n)
	for i := range ids {
		ids[i] = uuid.New()
		p.Media = append(p.Media, model.Media{ID: ids[i], ProjectID: p.ID, SortOrder: i})
	}
	return p, ids
}

func TestReorderMedia_ValidOrder_CallsRepo(t *testing.T) {
	repo := &mockProjectRepo{}
	svc := NewProjectService(repo)
	ctx := context.Background()
	p, ids := projectWithMedia(3)
	newOrder := []uuid.UUID{ids[2], ids[0], ids[1]}

	repo.On("GetByID", ctx, p.ID).Return(p, nil)
	repo.On("ReorderMedia", ctx, p.ID, newOrder).Return(nil)

	require.NoError(t, svc.ReorderMedia(ctx, p.ID, newOrder))
	repo.AssertExpectations(t)
}

func TestReorderMedia_InvalidLists_ReturnError(t *testing.T) {
	ctx := context.Background()
	p, ids := projectWithMedia(3)
	for name, order := range map[string][]uuid.UUID{
		"missing photo": {ids[0], ids[1]},
		"foreign photo": {ids[0], ids[1], uuid.New()},
		"duplicate":     {ids[0], ids[0], ids[1]},
	} {
		t.Run(name, func(t *testing.T) {
			repo := &mockProjectRepo{}
			repo.On("GetByID", ctx, p.ID).Return(p, nil)
			assert.Error(t, NewProjectService(repo).ReorderMedia(ctx, p.ID, order))
			repo.AssertNotCalled(t, "ReorderMedia", mock.Anything, mock.Anything, mock.Anything)
		})
	}
}

// Two residential projects and one commercial one.
func projectsByCategory() ([]model.Project, []uuid.UUID) {
	ids := []uuid.UUID{uuid.New(), uuid.New(), uuid.New()}
	return []model.Project{
		{ID: ids[0], Category: CategoryResidential},
		{ID: ids[1], Category: CategoryCommercial},
		{ID: ids[2], Category: CategoryResidential},
	}, ids
}

func TestReorderProjects_ValidOrder_CallsRepo(t *testing.T) {
	repo := &mockProjectRepo{}
	ctx := context.Background()
	projects, ids := projectsByCategory()
	newOrder := []uuid.UUID{ids[2], ids[0]}

	repo.On("List", ctx).Return(projects, nil)
	repo.On("ReorderProjects", ctx, CategoryResidential, newOrder).Return(nil)

	require.NoError(t, NewProjectService(repo).ReorderProjects(ctx, CategoryResidential, newOrder))
	repo.AssertExpectations(t)
}

func TestReorderProjects_InvalidInput_ReturnError(t *testing.T) {
	ctx := context.Background()
	projects, ids := projectsByCategory()
	for name, tc := range map[string]struct {
		category string
		order    []uuid.UUID
	}{
		"missing project":           {CategoryResidential, []uuid.UUID{ids[0]}},
		"project of other category": {CategoryResidential, []uuid.UUID{ids[0], ids[1]}},
		"unknown project":           {CategoryResidential, []uuid.UUID{ids[0], uuid.New()}},
		"duplicate":                 {CategoryResidential, []uuid.UUID{ids[0], ids[0]}},
		"empty category":            {"", []uuid.UUID{ids[1]}},
		"unknown category":          {"office", []uuid.UUID{ids[1]}},
	} {
		t.Run(name, func(t *testing.T) {
			repo := &mockProjectRepo{}
			repo.On("List", ctx).Return(projects, nil)
			assert.Error(t, NewProjectService(repo).ReorderProjects(ctx, tc.category, tc.order))
			repo.AssertNotCalled(t, "ReorderProjects", mock.Anything, mock.Anything, mock.Anything)
		})
	}
}

func TestCreate_EnglishTexts_PassedToRepo(t *testing.T) {
	repo := &mockProjectRepo{}
	svc := NewProjectService(repo)
	ctx := context.Background()
	in := model.ProjectFields{Title: "Дом", TitleEN: "House", Description: "Описание", DescriptionEN: "Description"}
	want := in
	want.Category = "residential"
	repo.On("Create", ctx, want).Return(&model.Project{ID: uuid.New()}, nil)

	_, err := svc.Create(ctx, in)
	require.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestCreate_EnglishDescriptionTooLong_ReturnsError(t *testing.T) {
	svc := NewProjectService(&mockProjectRepo{})
	_, err := svc.Create(context.Background(), model.ProjectFields{Title: "Title", DescriptionEN: strings.Repeat("x", 501)})
	assert.Error(t, err)
}
