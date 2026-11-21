package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"design-profile/backend/internal/model"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// ── Mock: otpStorer ──────────────────────────────────────────────────────────

type mockOTPStorer struct{ mock.Mock }

func (m *mockOTPStorer) Create(ctx context.Context, email, code string, expiresAt time.Time) (*model.OTPToken, error) {
	args := m.Called(ctx, email, code, expiresAt)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.OTPToken), args.Error(1)
}

func (m *mockOTPStorer) ConsumeAttempt(ctx context.Context, email string, maxAttempts int) (*model.OTPToken, error) {
	args := m.Called(ctx, email, maxAttempts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.OTPToken), args.Error(1)
}

func (m *mockOTPStorer) AttemptsSince(ctx context.Context, email string, since time.Time) (int, error) {
	args := m.Called(ctx, email, since)
	return args.Int(0), args.Error(1)
}

func (m *mockOTPStorer) CreatedSince(ctx context.Context, email string, since time.Time) (bool, error) {
	args := m.Called(ctx, email, since)
	return args.Bool(0), args.Error(1)
}

func (m *mockOTPStorer) MarkUsed(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

// ── Mock: emailSender ────────────────────────────────────────────────────────

type mockEmailSender struct{ mock.Mock }

func (m *mockEmailSender) SendOTP(to, code string) error {
	return m.Called(to, code).Error(0)
}

func (m *mockEmailSender) Host() string { return "smtp.test.local" }

// ── Helpers ──────────────────────────────────────────────────────────────────

const (
	testAdmin       = "admin@example.com"
	testJWTSecret   = "test-secret-long-enough-for-hmac-sha256-algo"
	testJWTExpHours = 1
)

// allowLogin sets up a login that is neither locked nor within the resend interval.
func allowLogin(otp *mockOTPStorer) {
	otp.On("AttemptsSince", mock.Anything, testAdmin, mock.Anything).Return(0, nil)
	otp.On("CreatedSince", mock.Anything, testAdmin, mock.Anything).Return(false, nil).Maybe() // only when a code is requested
}

func newAuthSvc(otp *mockOTPStorer, mail *mockEmailSender) *AuthService {
	return NewAuthService(otp, mail, testAdmin, testJWTSecret, testJWTExpHours)
}

// ── RequestOTP ───────────────────────────────────────────────────────────────

func TestRequestOTP_WrongEmail_Silent(t *testing.T) {
	otp := &mockOTPStorer{}
	mail := &mockEmailSender{}
	svc := newAuthSvc(otp, mail)

	err := svc.RequestOTP(context.Background(), "attacker@evil.com")

	assert.NoError(t, err)
	otp.AssertNotCalled(t, "Create")
	mail.AssertNotCalled(t, "SendOTP")
}

func TestRequestOTP_CorrectEmail_CreatesOTPAndSendsEmail(t *testing.T) {
	otp := &mockOTPStorer{}
	mail := &mockEmailSender{}
	svc := newAuthSvc(otp, mail)
	allowLogin(otp)

	fakeToken := &model.OTPToken{ID: uuid.New(), Email: testAdmin}
	otp.On("Create", mock.Anything, testAdmin, mock.AnythingOfType("string"), mock.Anything).
		Return(fakeToken, nil)
	mail.On("SendOTP", testAdmin, mock.AnythingOfType("string")).Return(nil)

	err := svc.RequestOTP(context.Background(), testAdmin)

	require.NoError(t, err)
	otp.AssertExpectations(t)
	mail.AssertExpectations(t)
}

func TestRequestOTP_SMTPError_ReturnedAsNil(t *testing.T) {
	// SMTP failures must not propagate — prevents email enumeration.
	otp := &mockOTPStorer{}
	mail := &mockEmailSender{}
	svc := newAuthSvc(otp, mail)
	allowLogin(otp)

	fakeToken := &model.OTPToken{ID: uuid.New(), Email: testAdmin}
	otp.On("Create", mock.Anything, testAdmin, mock.AnythingOfType("string"), mock.Anything).
		Return(fakeToken, nil)
	mail.On("SendOTP", testAdmin, mock.AnythingOfType("string")).
		Return(errors.New("connection refused"))

	err := svc.RequestOTP(context.Background(), testAdmin)

	assert.NoError(t, err, "SMTP failure must be silenced to prevent enumeration")
	otp.AssertExpectations(t)
	mail.AssertExpectations(t)
}

func TestRequestOTP_StoreError_ReturnsError(t *testing.T) {
	otp := &mockOTPStorer{}
	mail := &mockEmailSender{}
	svc := newAuthSvc(otp, mail)
	allowLogin(otp)

	otp.On("Create", mock.Anything, testAdmin, mock.AnythingOfType("string"), mock.Anything).
		Return(nil, errors.New("db unavailable"))

	err := svc.RequestOTP(context.Background(), testAdmin)

	require.Error(t, err)
	mail.AssertNotCalled(t, "SendOTP")
}

// ── VerifyOTP ────────────────────────────────────────────────────────────────

func TestVerifyOTP_ValidCode_ReturnsJWT(t *testing.T) {
	otp := &mockOTPStorer{}
	mail := &mockEmailSender{}
	svc := newAuthSvc(otp, mail)

	tokenID := uuid.New()
	fakeToken := &model.OTPToken{ID: tokenID, Email: testAdmin, Code: "123456"}

	allowLogin(otp)
	otp.On("ConsumeAttempt", mock.Anything, testAdmin, otpMaxAttempts).Return(fakeToken, nil)
	otp.On("MarkUsed", mock.Anything, tokenID).Return(nil)

	jwt, err := svc.VerifyOTP(context.Background(), testAdmin, "123456")

	require.NoError(t, err)
	assert.NotEmpty(t, jwt)
	otp.AssertExpectations(t)
}

func TestVerifyOTP_InvalidCode_ReturnsErrInvalidOTP(t *testing.T) {
	otp := &mockOTPStorer{}
	mail := &mockEmailSender{}
	svc := newAuthSvc(otp, mail)

	allowLogin(otp)
	otp.On("ConsumeAttempt", mock.Anything, testAdmin, otpMaxAttempts).
		Return(nil, errors.New("not found"))

	_, err := svc.VerifyOTP(context.Background(), testAdmin, "000000")

	require.ErrorIs(t, err, ErrInvalidOTP)
	otp.AssertNotCalled(t, "MarkUsed")
}

func TestVerifyOTP_MarkUsedFails_ReturnsError(t *testing.T) {
	otp := &mockOTPStorer{}
	mail := &mockEmailSender{}
	svc := newAuthSvc(otp, mail)

	tokenID := uuid.New()
	fakeToken := &model.OTPToken{ID: tokenID, Email: testAdmin, Code: "111111"}

	allowLogin(otp)
	otp.On("ConsumeAttempt", mock.Anything, testAdmin, otpMaxAttempts).Return(fakeToken, nil)
	otp.On("MarkUsed", mock.Anything, tokenID).Return(errors.New("db error"))

	_, err := svc.VerifyOTP(context.Background(), testAdmin, "111111")

	require.Error(t, err)
	assert.NotErrorIs(t, err, ErrInvalidOTP)
}

// ── Brute-force limits ───────────────────────────────────────────────────────

func TestVerifyOTP_WrongCode_ReturnsErrInvalidOTP(t *testing.T) {
	otp := &mockOTPStorer{}
	svc := newAuthSvc(otp, &mockEmailSender{})
	allowLogin(otp)
	otp.On("ConsumeAttempt", mock.Anything, testAdmin, otpMaxAttempts).
		Return(&model.OTPToken{ID: uuid.New(), Code: "123456", Attempts: 1}, nil)

	_, err := svc.VerifyOTP(context.Background(), testAdmin, "654321")

	require.ErrorIs(t, err, ErrInvalidOTP)
	otp.AssertNotCalled(t, "MarkUsed")
}

func TestVerifyOTP_Locked_ReturnsErrTooManyAttempts(t *testing.T) {
	otp := &mockOTPStorer{}
	svc := newAuthSvc(otp, &mockEmailSender{})
	otp.On("AttemptsSince", mock.Anything, testAdmin, mock.Anything).Return(otpWindowAttempts, nil)

	_, err := svc.VerifyOTP(context.Background(), testAdmin, "123456")

	require.ErrorIs(t, err, ErrTooManyAttempts)
	otp.AssertNotCalled(t, "ConsumeAttempt")
}

func TestVerifyOTP_OtherEmail_RejectedWithoutDatabase(t *testing.T) {
	otp := &mockOTPStorer{}
	svc := newAuthSvc(otp, &mockEmailSender{})

	_, err := svc.VerifyOTP(context.Background(), "attacker@evil.com", "123456")

	require.ErrorIs(t, err, ErrInvalidOTP)
	otp.AssertNotCalled(t, "AttemptsSince")
	otp.AssertNotCalled(t, "ConsumeAttempt")
}

func TestRequestOTP_Locked_SendsNothing(t *testing.T) {
	otp := &mockOTPStorer{}
	mail := &mockEmailSender{}
	svc := newAuthSvc(otp, mail)
	otp.On("AttemptsSince", mock.Anything, testAdmin, mock.Anything).Return(otpWindowAttempts, nil)

	require.NoError(t, svc.RequestOTP(context.Background(), testAdmin))
	otp.AssertNotCalled(t, "Create")
	mail.AssertNotCalled(t, "SendOTP")
}

func TestRequestOTP_WithinResendInterval_SendsNothing(t *testing.T) {
	otp := &mockOTPStorer{}
	mail := &mockEmailSender{}
	svc := newAuthSvc(otp, mail)
	otp.On("AttemptsSince", mock.Anything, testAdmin, mock.Anything).Return(0, nil)
	otp.On("CreatedSince", mock.Anything, testAdmin, mock.Anything).Return(true, nil)

	require.NoError(t, svc.RequestOTP(context.Background(), testAdmin))
	otp.AssertNotCalled(t, "Create")
	mail.AssertNotCalled(t, "SendOTP")
}
