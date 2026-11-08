package service

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateAttachments(t *testing.T) {
	photo := AttachmentInput{Data: []byte("img"), ContentType: "image/jpeg", Filename: "a.jpg"}
	pdf := AttachmentInput{Data: []byte("pdf"), ContentType: "application/pdf", Filename: "a.pdf"}

	t.Run("accepts up to 10 photos", func(t *testing.T) {
		photos := make([]AttachmentInput, 10)
		for i := range photos {
			photos[i] = photo
		}
		assert.NoError(t, validateAttachments(photos))
	})

	t.Run("rejects 11 photos", func(t *testing.T) {
		photos := make([]AttachmentInput, 11)
		for i := range photos {
			photos[i] = photo
		}
		assert.Error(t, validateAttachments(photos))
	})

	t.Run("accepts single PDF", func(t *testing.T) {
		assert.NoError(t, validateAttachments([]AttachmentInput{pdf}))
	})

	t.Run("rejects multiple PDFs", func(t *testing.T) {
		assert.Error(t, validateAttachments([]AttachmentInput{pdf, pdf}))
	})

	t.Run("rejects PDF mixed with photos", func(t *testing.T) {
		assert.Error(t, validateAttachments([]AttachmentInput{pdf, photo}))
	})

	t.Run("rejects unsupported file type", func(t *testing.T) {
		bad := AttachmentInput{Data: []byte("x"), ContentType: "text/plain", Filename: "doc.txt"}
		assert.Error(t, validateAttachments([]AttachmentInput{bad}))
	})

	t.Run("rejects file exceeding size limit", func(t *testing.T) {
		big := AttachmentInput{
			Data:        make([]byte, maxAttachmentSize+1),
			ContentType: "image/jpeg",
			Filename:    "huge.jpg",
		}
		assert.Error(t, validateAttachments([]AttachmentInput{big}))
	})

	t.Run("accepts empty attachments", func(t *testing.T) {
		assert.NoError(t, validateAttachments(nil))
	})
}

func TestNormalizeContacts(t *testing.T) {
	tests := []struct {
		name          string
		input         CreateRequestInput
		wantErr       bool
		wantPreferred string
	}{
		{name: "email only defaults preference to email", input: CreateRequestInput{Email: "a@b.co"}, wantPreferred: "email"},
		{name: "keeps filled preferred method", input: CreateRequestInput{Email: "a@b.co", Telegram: "@olga", PreferredContact: "telegram"}, wantPreferred: "telegram"},
		{name: "rejects missing email even with other contacts", input: CreateRequestInput{Phone: "+7 900 000-00-00"}, wantErr: true},
		{name: "rejects blank email", input: CreateRequestInput{Email: "   "}, wantErr: true},
		{name: "rejects malformed email", input: CreateRequestInput{Email: "not-an-email"}, wantErr: true},
		{name: "rejects display-name email", input: CreateRequestInput{Email: "Olga <a@b.co>"}, wantErr: true},
		{name: "rejects preference for empty contact", input: CreateRequestInput{Email: "a@b.co", PreferredContact: "phone"}, wantErr: true},
		{name: "rejects whitespace-only preferred contact value", input: CreateRequestInput{Email: "a@b.co", Telegram: "  ", PreferredContact: "telegram"}, wantErr: true},
		{name: "rejects unknown preference", input: CreateRequestInput{Email: "a@b.co", PreferredContact: "fax"}, wantErr: true},
		{name: "rejects overlong contact", input: CreateRequestInput{Email: "a@b.co", Phone: strings.Repeat("1", maxContactLength+1)}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := tt.input
			err := normalizeContacts(&input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.wantPreferred, input.PreferredContact)
		})
	}

	t.Run("trims surrounding whitespace", func(t *testing.T) {
		input := CreateRequestInput{Email: " a@b.co ", Telegram: " @olga "}
		assert.NoError(t, normalizeContacts(&input))
		assert.Equal(t, "a@b.co", input.Email)
		assert.Equal(t, "@olga", input.Telegram)
	})
}
