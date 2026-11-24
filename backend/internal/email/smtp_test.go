package email

import (
	"errors"
	"fmt"
	"net/textproto"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsPermanent(t *testing.T) {
	wrap := func(code int) error { return fmt.Errorf("send mail: %w", &textproto.Error{Code: code}) }
	assert.True(t, IsPermanent(wrap(550)))
	assert.True(t, IsPermanent(wrap(554)))
	assert.False(t, IsPermanent(wrap(421)))
	assert.False(t, IsPermanent(wrap(451)))
	assert.False(t, IsPermanent(wrap(535)), "a rejected login is fixed in the config")
	assert.False(t, IsPermanent(errors.New("i/o timeout")))
}

func TestHeaderSafe_DropsLineBreaks(t *testing.T) {
	assert.Equal(t, "Anna Bcc: x@evil.com", headerSafe("Anna\r\nBcc: x@evil.com"))
}
