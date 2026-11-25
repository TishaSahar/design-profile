package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeNotifications(t *testing.T, text string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "notifications.yaml")
	require.NoError(t, os.WriteFile(path, []byte(text), 0o600))
	return path
}

func TestLoadNotifications_MissingFileMeansOff(t *testing.T) {
	n, err := LoadNotifications(filepath.Join(t.TempDir(), "notifications.yaml"))
	require.NoError(t, err)
	assert.Nil(t, n)
}

func TestLoadNotifications_Valid(t *testing.T) {
	n, err := LoadNotifications(writeNotifications(t, "site_url: https://sakhadesign.com\nnew_order:\n  recipients:\n    - a@example.com\n    - b@example.com\n"))
	require.NoError(t, err)
	assert.Equal(t, []string{"a@example.com", "b@example.com"}, n.NewOrder.Recipients)
	assert.Equal(t, "https://sakhadesign.com", n.SiteURL)
}

func TestLoadNotifications_RejectsTypos(t *testing.T) {
	for _, text := range []string{
		"new_order:\n  recipients:\n    - not-an-email\n",
		"new_order:\n  recipients:\n    - Olga <a@example.com>\n",
		"site_url: sakhadesign.com\n",
		"new_order: [broken\n",
	} {
		_, err := LoadNotifications(writeNotifications(t, text))
		assert.Error(t, err, text)
	}
}

func TestNotificationsExample_IsValid(t *testing.T) {
	n, err := LoadNotifications("notifications.yaml.example")
	require.NoError(t, err)
	assert.NotEmpty(t, n.NewOrder.Recipients)
	require.NotNil(t, n.Outbox.ErrorDelay)
	assert.Equal(t, time.Minute, *n.Outbox.ErrorDelay)
}

// The outbox section itself is tested with tasks.Config; here, that it is read and checked.
func TestLoadNotifications_Outbox(t *testing.T) {
	n, err := LoadNotifications(writeNotifications(t, "outbox:\n  delay: 0s\n  error_delay: 2m\n"))
	require.NoError(t, err)
	assert.Equal(t, 2*time.Minute, n.Outbox.Settings().ErrorDelay)

	for _, text := range []string{"outbox:\n  error_delay: 100ms\n", "outbox:\n  delay: soon\n"} {
		_, err := LoadNotifications(writeNotifications(t, text))
		assert.Error(t, err, text)
	}
}
