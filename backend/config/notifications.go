package config

import (
	"errors"
	"fmt"
	"io/fs"
	"net/mail"
	"net/url"
	"os"

	"design-profile/backend/internal/tasks"

	"gopkg.in/yaml.v3"
)

// Notifications is notifications.yaml: who gets emails about events on the site.
type Notifications struct {
	// SiteURL is the public address of the site, for links in the emails.
	SiteURL  string `yaml:"site_url"`
	NewOrder struct {
		// Recipients get an email about every new project request.
		Recipients []string `yaml:"recipients"`
	} `yaml:"new_order"`
	// Outbox paces the worker that sends the notifications (see tasks.Config).
	Outbox tasks.Config `yaml:"outbox"`
}

// LoadNotifications reads notifications.yaml. A missing file means no notifications
// (nil, nil); a file with an invalid address or URL is an error, so a typo does
// not silently stop the emails.
func LoadNotifications(path string) (*Notifications, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read notifications config: %w", err)
	}
	var n Notifications
	if err := yaml.Unmarshal(data, &n); err != nil {
		return nil, fmt.Errorf("parse notifications config: %w", err)
	}
	for _, r := range n.NewOrder.Recipients {
		if a, err := mail.ParseAddress(r); err != nil || a.Address != r {
			return nil, fmt.Errorf("notifications config: %q is not an email address", r)
		}
	}
	if err := n.Outbox.Validate(); err != nil {
		return nil, fmt.Errorf("notifications config: outbox.%w", err)
	}
	if n.SiteURL != "" {
		if u, err := url.Parse(n.SiteURL); err != nil || (u.Scheme != "https" && u.Scheme != "http") || u.Host == "" {
			return nil, fmt.Errorf("notifications config: site_url %q must look like https://example.com", n.SiteURL)
		}
	}
	return &n, nil
}
