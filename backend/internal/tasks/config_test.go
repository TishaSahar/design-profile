package tasks

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func dur(d time.Duration) *time.Duration { return &d }

func TestConfigSettings_DefaultsForWhatIsLeftOut(t *testing.T) {
	assert.Equal(t, DefaultSettings, Config{}.Settings())

	s := Config{Delay: dur(0), ErrorDelay: dur(2 * time.Minute)}.Settings()
	assert.Equal(t, time.Duration(0), s.Delay, "0s is kept, not taken for missing")
	assert.Equal(t, 2*time.Minute, s.ErrorDelay)
	assert.Equal(t, DefaultSettings.NoTasksDelay, s.NoTasksDelay)
}

func TestConfigValidate(t *testing.T) {
	assert.NoError(t, Config{}.Validate())
	assert.NoError(t, Config{Delay: dur(0), ErrorDelay: dur(MinDelay), NoTasksDelay: dur(time.Hour)}.Validate())

	for name, c := range map[string]Config{
		"negative delay":       {Delay: dur(-time.Second)},
		"error_delay too low":  {ErrorDelay: dur(100 * time.Millisecond)},
		"no_tasks_delay is 0s": {NoTasksDelay: dur(0)},
	} {
		assert.Error(t, c.Validate(), name)
	}
}
