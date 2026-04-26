package auth

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testSecret = "test-secret-key-for-unit-testing-32chars"

func TestGenerateAndValidateToken(t *testing.T) {
	t.Run("valid token round-trip", func(t *testing.T) {
		email := "admin@example.com"
		token, err := GenerateToken(email, testSecret, 24)
		require.NoError(t, err)
		assert.NotEmpty(t, token)

		claims, err := ValidateToken(token, testSecret)
		require.NoError(t, err)
		assert.Equal(t, email, claims.Email)
	})

	t.Run("rejects wrong secret", func(t *testing.T) {
		token, err := GenerateToken("admin@example.com", testSecret, 24)
		require.NoError(t, err)

		_, err = ValidateToken(token, "wrong-secret")
		assert.Error(t, err)
	})

	t.Run("rejects expired token", func(t *testing.T) {
		token, err := GenerateToken("admin@example.com", testSecret, -1) // already expired
		require.NoError(t, err)

		// Give the expiry a moment to pass.
		time.Sleep(10 * time.Millisecond)
		_, err = ValidateToken(token, testSecret)
		assert.Error(t, err)
	})

	t.Run("rejects malformed token", func(t *testing.T) {
		_, err := ValidateToken("not.a.jwt", testSecret)
		assert.Error(t, err)
	})
}
