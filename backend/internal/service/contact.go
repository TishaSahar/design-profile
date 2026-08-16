package service

import (
	"context"

	"design-profile/backend/internal/model"
)

type contactRepo interface {
	Get(ctx context.Context) (*model.Contacts, error)
	Update(ctx context.Context, c *model.Contacts) (*model.Contacts, error)
}

type ContactService struct {
	repo contactRepo
}

func NewContactService(repo contactRepo) *ContactService {
	return &ContactService{repo: repo}
}

func (s *ContactService) Get(ctx context.Context) (*model.Contacts, error) {
	return s.repo.Get(ctx)
}

func (s *ContactService) Update(ctx context.Context, c *model.Contacts) (*model.Contacts, error) {
	return s.repo.Update(ctx, c)
}
