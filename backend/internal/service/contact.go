package service

import (
	"context"

	"design-profile/backend/internal/model"
	"design-profile/backend/internal/repository"
)

type ContactService struct {
	repo *repository.ContactRepository
}

func NewContactService(repo *repository.ContactRepository) *ContactService {
	return &ContactService{repo: repo}
}

func (s *ContactService) Get(ctx context.Context) (*model.Contacts, error) {
	return s.repo.Get(ctx)
}

func (s *ContactService) Update(ctx context.Context, c *model.Contacts) (*model.Contacts, error) {
	return s.repo.Update(ctx, c)
}
