package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"design-profile/backend/internal/auth"
	"design-profile/backend/internal/email"
	"design-profile/backend/internal/model"
	"design-profile/backend/internal/repository"
)

const otpTTL = 10 * time.Minute

// ErrInvalidOTP is returned when an OTP is wrong or expired.
var ErrInvalidOTP = errors.New("invalid or expired OTP")

type AuthService struct {
	otpRepo    *repository.OTPRepository
	emailClient *email.Client
	adminEmail  string
	jwtSecret   string
	jwtExpHours int
}

func NewAuthService(
	otpRepo *repository.OTPRepository,
	emailClient *email.Client,
	adminEmail, jwtSecret string,
	jwtExpHours int,
) *AuthService {
	return &AuthService{
		otpRepo:     otpRepo,
		emailClient: emailClient,
		adminEmail:  adminEmail,
		jwtSecret:   jwtSecret,
		jwtExpHours: jwtExpHours,
	}
}

// RequestOTP generates and emails an OTP to the admin. Returns an error if the
// requested email does not match the configured admin email.
func (s *AuthService) RequestOTP(ctx context.Context, requestedEmail string) error {
	if requestedEmail != s.adminEmail {
		// Do not reveal that the email is wrong to prevent enumeration.
		return nil
	}

	code, err := auth.GenerateOTP()
	if err != nil {
		return fmt.Errorf("generate otp: %w", err)
	}

	expiresAt := time.Now().Add(otpTTL)
	if _, err := s.otpRepo.Create(ctx, requestedEmail, code, expiresAt); err != nil {
		return fmt.Errorf("store otp: %w", err)
	}

	if err := s.emailClient.SendOTP(requestedEmail, code); err != nil {
		// Логируем ошибку на сервере, но не возвращаем её наружу:
		// хендлер всегда отвечает 200, чтобы не раскрывать существование email.
		slog.Error("не удалось отправить OTP на почту",
			"recipient", requestedEmail,
			"smtp_host", s.emailClient.Host(),
			"error", err,
		)
		return nil
	}
	return nil
}

// VerifyOTP validates the OTP and returns a signed JWT on success.
func (s *AuthService) VerifyOTP(ctx context.Context, email, code string) (string, error) {
	token, err := s.otpRepo.FindActive(ctx, email, code)
	if err != nil {
		return "", ErrInvalidOTP
	}

	if err := s.otpRepo.MarkUsed(ctx, token.ID); err != nil {
		return "", fmt.Errorf("mark otp used: %w", err)
	}

	jwt, err := auth.GenerateToken(email, s.jwtSecret, s.jwtExpHours)
	if err != nil {
		return "", fmt.Errorf("generate jwt: %w", err)
	}
	return jwt, nil
}

// ValidateToken verifies a JWT and returns its claims.
func (s *AuthService) ValidateToken(tokenStr string) (*auth.Claims, error) {
	return auth.ValidateToken(tokenStr, s.jwtSecret)
}

// Ensure compile-time model import is used.
var _ = model.OTPToken{}
