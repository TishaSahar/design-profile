package repository

import (
	"context"
	"fmt"

	"design-profile/backend/internal/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ContactRepository struct {
	db *pgxpool.Pool
}

func NewContactRepository(db *pgxpool.Pool) *ContactRepository {
	return &ContactRepository{db: db}
}

// Get returns the designer's contact information (single row).
func (r *ContactRepository) Get(ctx context.Context) (*model.Contacts, error) {
	c := &model.Contacts{}
	err := r.db.QueryRow(ctx,
		`SELECT telegram, instagram, email FROM contacts LIMIT 1`,
	).Scan(&c.Telegram, &c.Instagram, &c.Email)
	if err != nil {
		return nil, fmt.Errorf("get contacts: %w", err)
	}
	return c, nil
}

// Update modifies the designer's contact information.
func (r *ContactRepository) Update(ctx context.Context, c *model.Contacts) (*model.Contacts, error) {
	result := &model.Contacts{}
	err := r.db.QueryRow(ctx,
		`UPDATE contacts SET telegram = $1, instagram = $2, email = $3, updated_at = NOW()
		 RETURNING telegram, instagram, email`,
		c.Telegram, c.Instagram, c.Email,
	).Scan(&result.Telegram, &result.Instagram, &result.Email)
	if err != nil {
		return nil, fmt.Errorf("update contacts: %w", err)
	}
	return result, nil
}
