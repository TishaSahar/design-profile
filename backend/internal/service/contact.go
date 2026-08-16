package service

import (
	"context"

	"design-profile/backend/internal/model"
)

// ContactService serves designer contact information loaded from a YAML file.
// Contact data is immutable at runtime; edit config/about/contacts.yaml to
// change what is shown on the public site.
type ContactService struct {
	contacts *model.Contacts
}

func NewContactService(contacts *model.Contacts) *ContactService {
	return &ContactService{contacts: contacts}
}

func (s *ContactService) Get(_ context.Context) (*model.Contacts, error) {
	return s.contacts, nil
}
