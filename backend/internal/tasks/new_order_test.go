package tasks

import (
	"context"
	"encoding/json"
	"errors"
	"net/textproto"
	"testing"
	"time"

	"design-profile/backend/internal/email"
	"design-profile/backend/internal/model"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeMailer struct {
	sent []email.Message
	err  error
}

func (m *fakeMailer) Send(msg email.Message) error {
	m.sent = append(m.sent, msg)
	return m.err
}

func orderBody(t *testing.T) ([]byte, model.NewOrderNotification) {
	t.Helper()
	n := model.NewOrderNotification{
		RequestID: uuid.New(), FirstName: "Анна", LastName: "Иванова", Email: "anna@example.com",
		Telegram: "@anna", PreferredContact: "telegram", Description: "Квартира 60 м²",
		Attachments: 2, SubmittedAt: time.Date(2026, 11, 22, 9, 30, 0, 0, time.UTC),
	}
	body, err := json.Marshal(n)
	require.NoError(t, err)
	return body, n
}

func TestNewOrderNotification_SendsRequestToAllRecipients(t *testing.T) {
	m := &fakeMailer{}
	body, n := orderBody(t)
	h := NewOrderNotification(m, []string{"olga@example.com", "team@example.com"}, "https://sakhadesign.com/")

	require.NoError(t, h(context.Background(), body))

	require.Len(t, m.sent, 1)
	msg := m.sent[0]
	assert.Equal(t, []string{"olga@example.com", "team@example.com"}, msg.To)
	assert.Equal(t, "anna@example.com", msg.ReplyTo)
	assert.Contains(t, msg.Subject, "Анна Иванова")
	for _, want := range []string{"anna@example.com", "@anna", "Предпочтительный способ связи: Telegram",
		"Вложений: 2", "Квартира 60 м²", "https://sakhadesign.com/admin.html#/requests/" + n.RequestID.String()} {
		assert.Contains(t, msg.Body, want)
	}
	assert.NotContains(t, msg.Body, "Телефон", "an empty contact is left out")
}

func TestNewOrderNotification_FailuresByKind(t *testing.T) {
	body, _ := orderBody(t)
	tests := []struct {
		name      string
		err       error
		permanent bool
	}{
		{"server unreachable", errors.New("dial tcp: connection refused"), false},
		{"try again later (421)", &textproto.Error{Code: 421, Msg: "busy"}, false},
		{"rate limited (450)", &textproto.Error{Code: 450, Msg: "too many messages"}, false},
		{"login rejected (535)", &textproto.Error{Code: 535, Msg: "bad credentials"}, false},
		{"no such mailbox (550)", &textproto.Error{Code: 550, Msg: "no such user"}, true},
		{"bad address (553)", &textproto.Error{Code: 553, Msg: "bad address"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewOrderNotification(&fakeMailer{err: tt.err}, []string{"olga@example.com"}, "")
			err := h(context.Background(), body)
			require.Error(t, err)
			var p *PermanentError
			assert.Equal(t, tt.permanent, errors.As(err, &p))
		})
	}
}

func TestNewOrderNotification_UnusableTaskIsPermanent(t *testing.T) {
	var p *PermanentError
	body, _ := orderBody(t)

	err := NewOrderNotification(&fakeMailer{}, []string{"olga@example.com"}, "")(context.Background(), []byte(`{broken`))
	assert.True(t, errors.As(err, &p), "a body that cannot be read")

	m := &fakeMailer{}
	err = NewOrderNotification(m, nil, "")(context.Background(), body)
	assert.True(t, errors.As(err, &p), "no recipients configured")
	assert.Empty(t, m.sent)
}
