package tasks

import (
	"fmt"
	"time"
)

// Settings pace the worker.
type Settings struct {
	// Delay is the pause after a task that is done (or failed for good) before
	// the next one, e.g. to stay within the sending limits of the mail server.
	Delay time.Duration
	// ErrorDelay is the pause after a temporary failure; the failed task itself
	// is retried no sooner than that either.
	ErrorDelay time.Duration
	// NoTasksDelay is how often the table is checked while nothing is due.
	// A newly queued task wakes the worker at once (see Wake).
	NoTasksDelay time.Duration
}

// DefaultSettings is used for what Config leaves out.
var DefaultSettings = Settings{
	Delay:        time.Second,
	ErrorDelay:   time.Minute,
	NoTasksDelay: 15 * time.Second,
}

// MinDelay is the least ErrorDelay and NoTasksDelay: shorter ones would hammer
// the mail server or the database for nothing.
const MinDelay = time.Second

// Config is how the worker is set up in a YAML file (the outbox section of
// notifications.yaml). Durations are written like 500ms, 30s, 2m or 1h; a value
// left out (nil) keeps its default from DefaultSettings.
type Config struct {
	Delay        *time.Duration `yaml:"delay"`
	ErrorDelay   *time.Duration `yaml:"error_delay"`
	NoTasksDelay *time.Duration `yaml:"no_tasks_delay"`
}

// Validate reports a value the worker cannot run with.
func (c Config) Validate() error {
	if c.Delay != nil && *c.Delay < 0 {
		return fmt.Errorf("delay must not be negative")
	}
	for _, d := range []struct {
		name  string
		value *time.Duration
	}{{"error_delay", c.ErrorDelay}, {"no_tasks_delay", c.NoTasksDelay}} {
		if d.value != nil && *d.value < MinDelay {
			return fmt.Errorf("%s must be at least %s, like 30s or 1m", d.name, MinDelay)
		}
	}
	return nil
}

// Settings returns the settings to run the worker with: the values set in the
// config, the defaults for the rest.
func (c Config) Settings() Settings {
	s := DefaultSettings
	if c.Delay != nil {
		s.Delay = *c.Delay
	}
	if c.ErrorDelay != nil {
		s.ErrorDelay = *c.ErrorDelay
	}
	if c.NoTasksDelay != nil {
		s.NoTasksDelay = *c.NoTasksDelay
	}
	return s
}
