package auth

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateOTP(t *testing.T) {
	t.Run("generates a 6-character numeric code", func(t *testing.T) {
		code, err := GenerateOTP()
		require.NoError(t, err)
		assert.Len(t, code, otpLength)
		_, err = strconv.Atoi(code)
		assert.NoError(t, err, "code must be numeric")
	})

	t.Run("generates unique codes", func(t *testing.T) {
		codes := make(map[string]struct{}, 100)
		for i := 0; i < 100; i++ {
			code, err := GenerateOTP()
			require.NoError(t, err)
			codes[code] = struct{}{}
		}
		// With 100 samples from a 1 000 000 space, we expect virtually no collisions.
		assert.Greater(t, len(codes), 90)
	})
}
