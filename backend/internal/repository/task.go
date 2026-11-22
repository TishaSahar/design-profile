package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"design-profile/backend/internal/model"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const taskColumns = `id, type, status, body, attempts, last_error, run_at, created_at, updated_at`

func taskScanTargets(t *model.Task) []any {
	return []any{&t.ID, &t.Type, &t.Status, &t.Body, &t.Attempts, &t.LastError, &t.RunAt, &t.CreatedAt, &t.UpdatedAt}
}

// TaskRepository is the outbox of background tasks (the tasks table).
type TaskRepository struct {
	db *pgxpool.Pool
}

func NewTaskRepository(db *pgxpool.Pool) *TaskRepository {
	return &TaskRepository{db: db}
}

// insertTasks queues tasks within a transaction, so they are stored together with
// the change they belong to — or not at all.
func insertTasks(ctx context.Context, tx pgx.Tx, tasks []model.NewTask) error {
	for _, t := range tasks {
		body, err := json.Marshal(t.Body)
		if err != nil {
			return fmt.Errorf("encode %s task: %w", t.Type, err)
		}
		if _, err := tx.Exec(ctx, `INSERT INTO tasks (type, body) VALUES ($1, $2)`, t.Type, body); err != nil {
			return fmt.Errorf("queue %s task: %w", t.Type, err)
		}
	}
	return nil
}

// Claim takes up to limit due tasks (new or error, run_at passed) for one run,
// oldest first, counting the attempt. A claimed task is put off by lease: if the
// backend stops before finishing it, the task runs again once the lease is over.
// SKIP LOCKED keeps two backends from taking the same task.
func (r *TaskRepository) Claim(ctx context.Context, limit int, lease time.Duration) ([]model.Task, error) {
	rows, err := r.db.Query(ctx,
		`UPDATE tasks SET attempts = attempts + 1, run_at = NOW() + $2::interval, updated_at = NOW()
		 WHERE id IN (SELECT id FROM tasks
		              WHERE status IN ('new', 'error') AND run_at <= NOW()
		              ORDER BY run_at LIMIT $1
		              FOR UPDATE SKIP LOCKED)
		 RETURNING `+taskColumns,
		limit, lease.String(),
	)
	if err != nil {
		return nil, fmt.Errorf("claim tasks: %w", err)
	}
	defer rows.Close()
	var tasks []model.Task
	for rows.Next() {
		var t model.Task
		if err := rows.Scan(taskScanTargets(&t)...); err != nil {
			return nil, fmt.Errorf("scan task: %w", err)
		}
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}

// Delete removes a task that succeeded.
func (r *TaskRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if _, err := r.db.Exec(ctx, `DELETE FROM tasks WHERE id = $1`, id); err != nil {
		return fmt.Errorf("delete task: %w", err)
	}
	return nil
}

// MarkError records a temporary failure; the task runs again at runAt.
func (r *TaskRepository) MarkError(ctx context.Context, id uuid.UUID, reason string, runAt time.Time) error {
	_, err := r.db.Exec(ctx,
		`UPDATE tasks SET status = 'error', last_error = $2, run_at = $3, updated_at = NOW() WHERE id = $1`,
		id, reason, runAt,
	)
	if err != nil {
		return fmt.Errorf("mark task error: %w", err)
	}
	return nil
}

// MarkFatal records a failure that retrying cannot fix; the task is not run again.
func (r *TaskRepository) MarkFatal(ctx context.Context, id uuid.UUID, reason string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE tasks SET status = 'fatal', last_error = $2, updated_at = NOW() WHERE id = $1`,
		id, reason,
	)
	if err != nil {
		return fmt.Errorf("mark task fatal: %w", err)
	}
	return nil
}
