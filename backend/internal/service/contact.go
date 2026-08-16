package service

import (
	"context"

	"design-profile/backend/internal/model"
)

// ContactService serves designer contact information loaded from a YAML file.
// Contact data is immutable at runtime; edit config/about/contacts.yaml and
// config/about/bio.txt to change what is shown on the public site.
type ContactService struct {
	contacts  *model.Contacts
	photoPath string
}

func NewContactService(contacts *model.Contacts, photoPath string) *ContactService {
	return &ContactService{contacts: contacts, photoPath: photoPath}
}

func (s *ContactService) Get(_ context.Context) (*model.Contacts, error) {
	return s.contacts, nil
}

// PhotoPath returns the filesystem path of the designer's profile photo.
// Returns an empty string if no photo is configured.
func (s *ContactService) PhotoPath() string {
	return s.photoPath
}
