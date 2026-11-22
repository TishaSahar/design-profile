package model

import (
	"encoding/json"
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
	// Attempts is how many times the code was checked, right or wrong.
	Attempts int `db:"attempts"`
}

// Project represents a designer's portfolio project.
type Project struct {
	ID          uuid.UUID `json:"id"                     db:"id"`
	Title       string    `json:"title"                  db:"title"`
	Description string    `json:"description"            db:"description"`
	// TitleEN and DescriptionEN are the English texts; empty when not translated.
	TitleEN       string     `json:"title_en"               db:"title_en"`
	DescriptionEN string     `json:"description_en"         db:"description_en"`
	Category      string     `json:"category"               db:"category"`
	CoverMediaID  *uuid.UUID `json:"cover_media_id,omitempty" db:"cover_media_id"`
	// SortOrder is the position within the category on the public site (0 = first).
	SortOrder int       `json:"sort_order"             db:"sort_order"`
	CreatedAt time.Time `json:"created_at"             db:"created_at"`
	UpdatedAt time.Time `json:"updated_at"             db:"updated_at"`
	Media     []Media   `json:"media,omitempty"        db:"-"`
}

// ProjectFields are the editable texts and category of a project.
type ProjectFields struct {
	Title         string
	TitleEN       string
	Description   string
	DescriptionEN string
	Category      string
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

// MediaFile is the content of a project photo in one of its sizes.
type MediaFile struct {
	Data        []byte
	ContentType string
	// Final is false while a requested web copy is not made yet and the
	// original is served in its place: such a response must not be cached long.
	Final bool
}

// ProjectRequest represents a client's project inquiry.
type ProjectRequest struct {
	ID        uuid.UUID `json:"id"         db:"id"`
	FirstName string    `json:"first_name" db:"first_name"`
	LastName  string    `json:"last_name"  db:"last_name"`
	// Contact is the free-text contact of requests created before the
	// dedicated fields below existed; it is empty for newer requests.
	Contact          string       `json:"contact,omitempty"     db:"contact"`
	Email            string       `json:"email"                 db:"email"`
	Phone            string       `json:"phone,omitempty"       db:"phone"`
	Telegram         string       `json:"telegram,omitempty"    db:"telegram"`
	PreferredContact string       `json:"preferred_contact"     db:"preferred_contact"`
	Description      string       `json:"description"           db:"description"`
	Consented        bool         `json:"consented"             db:"consented"`
	CreatedAt        time.Time    `json:"created_at"            db:"created_at"`
	Attachments      []Attachment `json:"attachments,omitempty" db:"-"`
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

// LocalizedText is a text in Russian with an optional English translation.
type LocalizedText struct {
	RU string `json:"ru"           yaml:"ru"`
	EN string `json:"en,omitempty" yaml:"en"`
}

// FAQItem is one question and its answer in the FAQ on the public site.
type FAQItem struct {
	Question LocalizedText `json:"question" yaml:"question"`
	Answer   LocalizedText `json:"answer"   yaml:"answer"`
}

// Contacts holds the designer's public contact information.
type Contacts struct {
	Telegram  string `json:"telegram"            db:"telegram"`
	Instagram string `json:"instagram"           db:"instagram"`
	Email     string `json:"email"               db:"email"`
	Bio       string `json:"bio,omitempty"`
	BioEn     string `json:"bio_en,omitempty"`
	PhotoURL  string `json:"photo_url,omitempty"`
}

// Task types: what a background task does (see the tasks table).
const (
	// TaskNewOrderNotification emails the admins about a new project request;
	// its body is a NewOrderNotification.
	TaskNewOrderNotification = "new_order_notification"
)

// Task statuses. A task that succeeded is deleted, so there is no "done".
const (
	TaskStatusNew   = "new"   // waiting for its first run
	TaskStatusError = "error" // failed for a temporary reason, retried at RunAt
	TaskStatusFatal = "fatal" // cannot succeed; kept for a human to look at
)

// Task is a background task from the outbox table.
type Task struct {
	ID        uuid.UUID       `json:"id"`
	Type      string          `json:"type"`
	Status    string          `json:"status"`
	Body      json.RawMessage `json:"body"`
	Attempts  int             `json:"attempts"`
	LastError string          `json:"last_error"`
	RunAt     time.Time       `json:"run_at"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

// NewTask is a task to queue: its type and its body, encoded as JSON.
type NewTask struct {
	Type string
	Body any
}

// NewOrderNotification is the body of a TaskNewOrderNotification: the request
// as it was submitted, so the email does not depend on the request staying in the database.
type NewOrderNotification struct {
	RequestID        uuid.UUID `json:"request_id"`
	FirstName        string    `json:"first_name"`
	LastName         string    `json:"last_name"`
	Email            string    `json:"email"`
	Phone            string    `json:"phone,omitempty"`
	Telegram         string    `json:"telegram,omitempty"`
	PreferredContact string    `json:"preferred_contact"`
	Description      string    `json:"description"`
	Attachments      int       `json:"attachments"`
	SubmittedAt      time.Time `json:"submitted_at"`
}
