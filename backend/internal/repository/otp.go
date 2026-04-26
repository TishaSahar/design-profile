package repository

import (
	"context"
	"fmt"
	"time"

	"design-profile/backend/internal/model"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

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
		 RETURNING id, email, code, expires_at, used, created_at`,
		email, code, expiresAt,
	).Scan(&token.ID, &token.Email, &token.Code, &token.ExpiresAt, &token.Used, &token.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("insert otp token: %w", err)
	}
	return token, nil
}

// FindActive returns the most recent active (unused, not expired) OTP for the given email and code.
func (r *OTPRepository) FindActive(ctx context.Context, email, code string) (*model.OTPToken, error) {
	token := &model.OTPToken{}
	err := r.db.QueryRow(ctx,
		`SELECT id, email, code, expires_at, used, created_at
		 FROM otp_tokens
		 WHERE email = $1 AND code = $2 AND used = FALSE AND expires_at > NOW()
		 ORDER BY created_at DESC
		 LIMIT 1`,
		email, code,
	).Scan(&token.ID, &token.Email, &token.Code, &token.ExpiresAt, &token.Used, &token.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("find active otp: %w", err)
	}
	return token, nil
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
