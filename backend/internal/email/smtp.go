// Package email provides email sending via Gmail SMTP using an App Password.
// To use with Gmail:
//  1. Enable 2-Step Verification on the Google account.
//  2. Go to Google Account → Security → App passwords and generate a password.
//  3. Use that password in the config (email.password) and the Gmail address as username/sender.
package email

import (
	"errors"
	"fmt"
	"mime"
	"net/smtp"
	"net/textproto"
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
	return c.Send(Message{To: []string{to}, Subject: subject, Body: body})
}

// Message is a plain-text email.
type Message struct {
	To      []string
	Subject string
	Body    string
	// ReplyTo, if set, is where a reply goes instead of the sender (e.g. the client).
	ReplyTo string
}

// Send sends a plain-text email to all of m.To in one SMTP transaction.
// Use IsPermanent to tell whether a failure is worth retrying.
func (c *Client) Send(m Message) error {
	addr := fmt.Sprintf("%s:%d", c.host, c.port)
	auth := smtp.PlainAuth("", c.user, c.pass, c.host)

	headers := []string{
		"From: " + c.sender,
		"To: " + strings.Join(m.To, ", "),
		// Encoded, so Cyrillic shows up right and a line break in it cannot add headers.
		"Subject: " + mime.QEncoding.Encode("utf-8", headerSafe(m.Subject)),
	}
	if m.ReplyTo != "" {
		headers = append(headers, "Reply-To: "+headerSafe(m.ReplyTo))
	}
	msg := strings.Join(append(headers,
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"Content-Transfer-Encoding: 8bit",
		"",
		strings.ReplaceAll(strings.ReplaceAll(m.Body, "\r\n", "\n"), "\n", "\r\n"),
	), "\r\n")

	if err := smtp.SendMail(addr, auth, c.sender, m.To, []byte(msg)); err != nil {
		return fmt.Errorf("send mail: %w", err)
	}
	return nil
}

// headerSafe drops line breaks, which would let a value start a header of its own.
func headerSafe(s string) string {
	return strings.Join(strings.Fields(strings.NewReplacer("\r", " ", "\n", " ").Replace(s)), " ")
}

// IsPermanent reports whether a send failed for good: the SMTP server rejected the
// message or a recipient with a 5xx reply, so sending it again cannot help.
// Everything else is temporary: the server could not be reached, replied 4xx
// ("try again later", rate limits), or did not accept the login (530, 534, 535) —
// the last one is fixed in the config, after which the same message goes through.
func IsPermanent(err error) bool {
	var reply *textproto.Error
	if !errors.As(err, &reply) {
		return false
	}
	switch reply.Code {
	case 530, 534, 535:
		return false
	}
	return reply.Code >= 500 && reply.Code < 600
}
