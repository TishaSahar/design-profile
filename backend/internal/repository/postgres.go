package repository

import (
	"context"
	"fmt"
	"io/fs"
	"log"
	"sort"

	"github.com/jackc/pgx/v5/pgxpool"
)

// NewPool creates and validates a new PostgreSQL connection pool.
func NewPool(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}
	return pool, nil
}

// Migrate runs all SQL files from the provided FS in lexicographic order.
func Migrate(ctx context.Context, pool *pgxpool.Pool, migrationsFS fs.FS) error {
	entries, err := fs.ReadDir(migrationsFS, ".")
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}

	// Sort to guarantee execution order.
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		// Skip non-SQL files (e.g. the .go embedding file).
		if len(name) < 4 || name[len(name)-4:] != ".sql" {
			continue
		}

		sql, err := fs.ReadFile(migrationsFS, name)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", name, err)
		}
		if _, err := pool.Exec(ctx, string(sql)); err != nil {
			return fmt.Errorf("apply migration %s: %w", name, err)
		}
		log.Printf("migration applied: %s", name)
	}
	return nil
}
