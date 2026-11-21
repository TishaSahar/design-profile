package repository

import (
	"context"
	"fmt"
	"time"

	"design-profile/backend/internal/model"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

const otpColumns = `id, email, code, expires_at, used, created_at, attempts`

// otpScanTargets returns scan destinations matching otpColumns.
func otpScanTargets(t *model.OTPToken) []any {
	return []any{&t.ID, &t.Email, &t.Code, &t.ExpiresAt, &t.Used, &t.CreatedAt, &t.Attempts}
}

type OTPRepository struct {
	db *pgxpool.Pool
}

func NewOTPRepository(db *pgxpool.Pool) *OTPRepository {
	return &OTPRepository{db: db}
}

// Create inserts a new OTP token, invalidating any previous unused tokens for the same email.
func (r *OTPRepository) Create(ctx context.Context, email, code string, expiresAt time.Time) (*model.OTPToken, error) {
	// Invalidate previous tokens for this email.
	_, err := r.db.Exec(ctx,
		`UPDATE otp_tokens SET used = TRUE WHERE email = $1 AND used = FALSE`,
		email,
	)
	if err != nil {
		return nil, fmt.Errorf("invalidate old tokens: %w", err)
	}

	token := &model.OTPToken{}
	err = r.db.QueryRow(ctx,
		`INSERT INTO otp_tokens (email, code, expires_at) VALUES ($1, $2, $3)
		 RETURNING `+otpColumns,
		email, code, expiresAt,
	).Scan(otpScanTargets(token)...)
	if err != nil {
		return nil, fmt.Errorf("insert otp token: %w", err)
	}
	return token, nil
}

// ConsumeAttempt takes one check of the newest active code of an email: it
// returns that code with its attempts counted, or pgx.ErrNoRows when there is no
// active code or its maxAttempts are used up. Concurrent checks cannot exceed
// maxAttempts: the row is re-checked under its lock.
func (r *OTPRepository) ConsumeAttempt(ctx context.Context, email string, maxAttempts int) (*model.OTPToken, error) {
	token := &model.OTPToken{}
	err := r.db.QueryRow(ctx,
		`UPDATE otp_tokens SET attempts = attempts + 1
		 WHERE id = (SELECT id FROM otp_tokens
		             WHERE email = $1 AND used = FALSE AND expires_at > NOW()
		             ORDER BY created_at DESC LIMIT 1)
		   AND used = FALSE AND attempts < $2
		 RETURNING `+otpColumns,
		email, maxAttempts,
	).Scan(otpScanTargets(token)...)
	if err != nil {
		return nil, fmt.Errorf("consume otp attempt: %w", err)
	}
	return token, nil
}

// AttemptsSince returns how many times the codes of an email were checked since a moment.
func (r *OTPRepository) AttemptsSince(ctx context.Context, email string, since time.Time) (int, error) {
	var n int
	err := r.db.QueryRow(ctx,
		`SELECT COALESCE(SUM(attempts), 0) FROM otp_tokens WHERE email = $1 AND created_at > $2`,
		email, since,
	).Scan(&n)
	if err != nil {
		return 0, fmt.Errorf("count otp attempts: %w", err)
	}
	return n, nil
}

// CreatedSince reports whether a code was issued to an email since a moment.
func (r *OTPRepository) CreatedSince(ctx context.Context, email string, since time.Time) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx,
		`SELECT EXISTS (SELECT 1 FROM otp_tokens WHERE email = $1 AND created_at > $2)`,
		email, since,
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check recent otp: %w", err)
	}
	return exists, nil
}

// MarkUsed marks the OTP token with the given ID as used.
func (r *OTPRepository) MarkUsed(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx,
		`UPDATE otp_tokens SET used = TRUE WHERE id = $1`,
		id,
	)
	if err != nil {
		return fmt.Errorf("mark otp used: %w", err)
	}
	return nil
}
