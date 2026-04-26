package auth

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

const otpLength = 6

// GenerateOTP generates a cryptographically secure 6-digit numeric code.
func GenerateOTP() (string, error) {
	max := big.NewInt(1_000_000) // 000000–999999
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", fmt.Errorf("generate otp: %w", err)
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}
