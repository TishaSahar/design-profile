package model

import (
	"time"

	"github.com/google/uuid"
)

// OTPToken represents a one-time password used for admin authentication.
type OTPToken struct {
	ID        uuid.UUID `db:"id"`
	Email     string    `db:"email"`
	Code      string    `db:"code"`
	ExpiresAt time.Time `db:"expires_at"`
	Used      bool      `db:"used"`
	CreatedAt time.Time `db:"created_at"`
}

// Project represents a designer's portfolio project.
type Project struct {
	ID           uuid.UUID  `json:"id"                     db:"id"`
	Title        string     `json:"title"                  db:"title"`
	Description  string     `json:"description"            db:"description"`
	CoverMediaID *uuid.UUID `json:"cover_media_id,omitempty" db:"cover_media_id"`
	CreatedAt    time.Time  `json:"created_at"             db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"             db:"updated_at"`
	Media        []Media    `json:"media,omitempty"        db:"-"`
}

// Media represents a binary media file attached to a project.
type Media struct {
	ID          uuid.UUID `json:"id"           db:"id"`
	ProjectID   uuid.UUID `json:"project_id"   db:"project_id"`
	Data        []byte    `json:"-"            db:"data"`
	ContentType string    `json:"content_type" db:"content_type"`
	Filename    string    `json:"filename"     db:"filename"`
	SortOrder   int       `json:"sort_order"   db:"sort_order"`
	CreatedAt   time.Time `json:"created_at"   db:"created_at"`
}

// ProjectRequest represents a client's project inquiry.
type ProjectRequest struct {
	ID          uuid.UUID    `json:"id"          db:"id"`
	FirstName   string       `json:"first_name"  db:"first_name"`
	LastName    string       `json:"last_name"   db:"last_name"`
	Contact     string       `json:"contact"     db:"contact"`
	Description string       `json:"description" db:"description"`
	Consented   bool         `json:"consented"   db:"consented"`
	CreatedAt   time.Time    `json:"created_at"  db:"created_at"`
	Attachments []Attachment `json:"attachments,omitempty" db:"-"`
}

// Attachment represents a binary file attached to a project request.
type Attachment struct {
	ID          uuid.UUID `json:"id"           db:"id"`
	RequestID   uuid.UUID `json:"request_id"   db:"request_id"`
	Data        []byte    `json:"-"            db:"data"`
	ContentType string    `json:"content_type" db:"content_type"`
	Filename    string    `json:"filename"     db:"filename"`
	CreatedAt   time.Time `json:"created_at"   db:"created_at"`
}

// Contacts holds the designer's public contact information.
type Contacts struct {
	Telegram  string `json:"telegram"  db:"telegram"`
	Instagram string `json:"instagram" db:"instagram"`
	Email     string `json:"email"     db:"email"`
}
