// Package email provides email sending via Gmail SMTP using an App Password.
// To use with Gmail:
//  1. Enable 2-Step Verification on the Google account.
//  2. Go to Google Account → Security → App passwords and generate a password.
//  3. Use that password in the config (email.password) and the Gmail address as username/sender.
package email

import (
	"fmt"
	"net/smtp"
	"strings"
)

// Client sends emails via SMTP.
type Client struct {
	host   string
	port   int
	user   string
	pass   string
	sender string
}

// NewClient creates a new SMTP email client.
func NewClient(host string, port int, user, pass, sender string) *Client {
	return &Client{
		host:   host,
		port:   port,
		user:   user,
		pass:   pass,
		sender: sender,
	}
}

// Host returns the configured SMTP host (used for logging).
func (c *Client) Host() string { return c.host }

// SendOTP sends a one-time password to the given recipient.
func (c *Client) SendOTP(to, code string) error {
	subject := "Код входа в панель дизайнера"
	body := fmt.Sprintf(
		"Ваш одноразовый код входа: %s\n\nКод действителен 10 минут.\nЕсли вы не запрашивали код, проигнорируйте это письмо.",
		code,
	)
	return c.send(to, subject, body)
}

func (c *Client) send(to, subject, body string) error {
	addr := fmt.Sprintf("%s:%d", c.host, c.port)
	auth := smtp.PlainAuth("", c.user, c.pass, c.host)

	msg := strings.Join([]string{
		"From: " + c.sender,
		"To: " + to,
		"Subject: " + subject,
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"",
		body,
	}, "\r\n")

	if err := smtp.SendMail(addr, auth, c.sender, []string{to}, []byte(msg)); err != nil {
		return fmt.Errorf("send mail: %w", err)
	}
	return nil
}
