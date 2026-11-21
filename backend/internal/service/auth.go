package service

import (
	"context"
	"crypto/subtle"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"design-profile/backend/internal/auth"
	"design-profile/backend/internal/model"

	"github.com/google/uuid"
)

// Limits that make guessing a 6-digit login code hopeless: at most 10 guesses an
// hour out of a million codes. Without them a script could try every code
// while one is valid.
const (
	otpTTL = 10 * time.Minute
	// otpMaxAttempts is how many times one code can be checked; then it is spent.
	otpMaxAttempts = 5
	// otpResendInterval is the least time between two codes: within it a new
	// request sends nothing and the code already sent stays valid.
	otpResendInterval = time.Minute
	// otpWindowAttempts checks of all codes within otpWindow lock the login until
	// older checks leave the window. Also keeps the admin's mailbox from a flood.
	otpWindow         = time.Hour
	otpWindowAttempts = 10
)

var (
	// ErrInvalidOTP is returned when an OTP is wrong, expired or spent.
	ErrInvalidOTP = errors.New("invalid or expired OTP")
	// ErrTooManyAttempts is returned while the login is locked after too many checks.
	ErrTooManyAttempts = errors.New("too many login attempts, try again later")
)

type otpStorer interface {
	Create(ctx context.Context, email, code string, expiresAt time.Time) (*model.OTPToken, error)
	ConsumeAttempt(ctx context.Context, email string, maxAttempts int) (*model.OTPToken, error)
	AttemptsSince(ctx context.Context, email string, since time.Time) (int, error)
	CreatedSince(ctx context.Context, email string, since time.Time) (bool, error)
	MarkUsed(ctx context.Context, id uuid.UUID) error
}

type emailSender interface {
	SendOTP(to, code string) error
	Host() string
}

type AuthService struct {
	otpRepo     otpStorer
	emailClient emailSender
	adminEmail  string
	jwtSecret   string
	jwtExpHours int
}

func NewAuthService(
	otpRepo otpStorer,
	emailClient emailSender,
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

// locked reports whether the login is locked after too many checks within otpWindow.
func (s *AuthService) locked(ctx context.Context, email string) (bool, error) {
	n, err := s.otpRepo.AttemptsSince(ctx, email, time.Now().Add(-otpWindow))
	if err != nil {
		return false, err
	}
	return n >= otpWindowAttempts, nil
}

// RequestOTP generates and emails an OTP to the admin. A wrong email, a request
// within otpResendInterval of the previous code and a locked login send nothing,
// and the caller is not told which, so nobody can probe the admin email.
func (s *AuthService) RequestOTP(ctx context.Context, requestedEmail string) error {
	if requestedEmail != s.adminEmail {
		// Do not reveal that the email is wrong to prevent enumeration.
		return nil
	}

	locked, err := s.locked(ctx, requestedEmail)
	if err != nil {
		return fmt.Errorf("check login lock: %w", err)
	}
	if locked {
		slog.Warn("код входа не отправлен: слишком много попыток за час")
		return nil
	}
	recent, err := s.otpRepo.CreatedSince(ctx, requestedEmail, time.Now().Add(-otpResendInterval))
	if err != nil {
		return fmt.Errorf("check recent otp: %w", err)
	}
	if recent {
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

// VerifyOTP validates the OTP and returns a signed JWT on success. Every check
// counts towards otpMaxAttempts of the code and otpWindowAttempts of the hour.
func (s *AuthService) VerifyOTP(ctx context.Context, email, code string) (string, error) {
	if email != s.adminEmail {
		return "", ErrInvalidOTP
	}
	locked, err := s.locked(ctx, email)
	if err != nil {
		return "", fmt.Errorf("check login lock: %w", err)
	}
	if locked {
		return "", ErrTooManyAttempts
	}

	token, err := s.otpRepo.ConsumeAttempt(ctx, email, otpMaxAttempts)
	if err != nil {
		// No active code, or its attempts are used up.
		return "", ErrInvalidOTP
	}
	if subtle.ConstantTimeCompare([]byte(token.Code), []byte(code)) != 1 {
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

