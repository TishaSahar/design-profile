package tasks

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"design-profile/backend/internal/email"
	"design-profile/backend/internal/model"
)

type mailer interface {
	Send(m email.Message) error
}

var contactLabels = map[string]string{"email": "E-mail", "phone": "Телефон", "telegram": "Telegram"}

// NewOrderNotification returns the handler of TaskNewOrderNotification: it emails
// the request to recipients. siteURL (e.g. https://sakhadesign.com), if set, adds
// a link to the request in the admin panel.
func NewOrderNotification(m mailer, recipients []string, siteURL string) Handler {
	return func(_ context.Context, body []byte) error {
		var n model.NewOrderNotification
		if err := json.Unmarshal(body, &n); err != nil {
			return Permanent(fmt.Errorf("decode notification: %w", err))
		}
		if len(recipients) == 0 {
			return Permanent(errors.New("no recipients: set new_order.recipients in notifications.yaml"))
		}
		err := m.Send(email.Message{
			To:      recipients,
			Subject: fmt.Sprintf("Новая заявка на проект: %s %s", n.FirstName, n.LastName),
			Body:    newOrderText(n, siteURL),
			ReplyTo: n.Email, // answering the email writes to the client
		})
		if err != nil && email.IsPermanent(err) {
			return Permanent(err)
		}
		return err
	}
}

func newOrderText(n model.NewOrderNotification, siteURL string) string {
	var b strings.Builder
	b.WriteString("Поступила новая заявка на проект.\n\n")
	fmt.Fprintf(&b, "Имя: %s %s\n", n.FirstName, n.LastName)
	fmt.Fprintf(&b, "E-mail: %s\n", n.Email)
	if n.Phone != "" {
		fmt.Fprintf(&b, "Телефон: %s\n", n.Phone)
	}
	if n.Telegram != "" {
		fmt.Fprintf(&b, "Telegram: %s\n", n.Telegram)
	}
	if label, ok := contactLabels[n.PreferredContact]; ok {
		fmt.Fprintf(&b, "Предпочтительный способ связи: %s\n", label)
	}
	fmt.Fprintf(&b, "Вложений: %d\n", n.Attachments)
	fmt.Fprintf(&b, "Отправлена: %s (UTC)\n", n.SubmittedAt.UTC().Format("02.01.2006 15:04"))
	fmt.Fprintf(&b, "\nОписание проекта:\n%s\n", n.Description)
	if siteURL != "" {
		fmt.Fprintf(&b, "\nОткрыть заявку в панели управления:\n%s/admin.html#/requests/%s\n",
			strings.TrimRight(siteURL, "/"), n.RequestID)
	}
	b.WriteString("\nЧтобы ответить клиенту, просто ответьте на это письмо.\n")
	return b.String()
}
