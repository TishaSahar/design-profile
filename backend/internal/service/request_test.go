package service

import (
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
