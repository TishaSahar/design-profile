// Package tasks runs the background tasks of the outbox table: it picks up due
// tasks, hands each to the handler of its type and records the outcome —
// deleted on success, retried later on a temporary failure (status error),
// given up on a permanent one (status fatal).
package tasks

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"design-profile/backend/internal/model"

	"github.com/google/uuid"
)

// Handler runs one task from its JSON body. Wrap an error in Permanent when
// running the task again cannot succeed; any other error is retried.
type Handler func(ctx context.Context, body []byte) error

// PermanentError marks a failure that retrying cannot fix.
type PermanentError struct{ Err error }

func (e *PermanentError) Error() string { return e.Err.Error() }
func (e *PermanentError) Unwrap() error { return e.Err }

// Permanent marks err as a failure that retrying cannot fix: the task becomes fatal.
func Permanent(err error) error { return &PermanentError{Err: err} }

type store interface {
	Claim(ctx context.Context, limit int, lease time.Duration) ([]model.Task, error)
	Delete(ctx context.Context, id uuid.UUID) error
	MarkError(ctx context.Context, id uuid.UUID, reason string, runAt time.Time) error
	MarkFatal(ctx context.Context, id uuid.UUID, reason string) error
}

const (
	// pollInterval: how often due tasks are looked for when nothing wakes the worker.
	pollInterval = 15 * time.Second
	batchSize    = 10
	// lease: a claimed task is not taken again for this long, so if the backend
	// stops mid-task the task still runs later — possibly twice, never zero times.
	lease = 5 * time.Minute
	// Retries wait 1, 2, 4… minutes after a failure, at most an hour.
	firstRetry = time.Minute
	maxRetry   = time.Hour
)

// Worker runs due tasks until its context is cancelled.
type Worker struct {
	store    store
	handlers map[string]Handler
	wake     chan struct{}
	now      func() time.Time
}

func NewWorker(store store) *Worker {
	return &Worker{
		store:    store,
		handlers: map[string]Handler{},
		wake:     make(chan struct{}, 1),
		now:      time.Now,
	}
}

// Handle registers the handler of a task type.
func (w *Worker) Handle(taskType string, h Handler) { w.handlers[taskType] = h }

// Wake makes the worker look for due tasks now instead of at its next poll,
// e.g. right after a task was queued. It never blocks.
func (w *Worker) Wake() {
	select {
	case w.wake <- struct{}{}:
	default:
	}
}

// Run processes due tasks until ctx is cancelled.
func (w *Worker) Run(ctx context.Context) {
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()
	for {
		w.RunDue(ctx)
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		case <-w.wake:
		}
	}
}

// RunDue processes the tasks due now, batch after batch, until none is left.
func (w *Worker) RunDue(ctx context.Context) {
	for ctx.Err() == nil {
		batch, err := w.store.Claim(ctx, batchSize, lease)
		if err != nil {
			slog.Error("claim tasks", "err", err)
			return
		}
		for _, t := range batch {
			w.run(ctx, t)
		}
		if len(batch) < batchSize {
			return
		}
	}
}

func (w *Worker) run(ctx context.Context, t model.Task) {
	log := slog.With("task_id", t.ID, "type", t.Type, "attempt", t.Attempts)
	h, ok := w.handlers[t.Type]
	var err error
	if !ok {
		err = Permanent(fmt.Errorf("no handler for task type %q", t.Type))
	} else {
		err = h(ctx, t.Body)
	}

	var permanent *PermanentError
	switch {
	case err == nil:
		if err := w.store.Delete(ctx, t.ID); err != nil {
			// The task ran; it runs again once its lease is over.
			log.Error("delete finished task", "err", err)
			return
		}
		log.Info("task done")
	case errors.As(err, &permanent):
		log.Error("task failed for good", "err", err)
		if err := w.store.MarkFatal(ctx, t.ID, err.Error()); err != nil {
			log.Error("mark task fatal", "err", err)
		}
	default:
		next := w.now().Add(retryDelay(t.Attempts))
		log.Warn("task failed, will retry", "err", err, "retry_at", next)
		if err := w.store.MarkError(ctx, t.ID, err.Error(), next); err != nil {
			log.Error("mark task error", "err", err)
		}
	}
}

// retryDelay is the wait after the given failed attempt (1-based).
func retryDelay(attempt int) time.Duration {
	d := firstRetry
	for i := 1; i < attempt && d < maxRetry; i++ {
		d *= 2
	}
	return min(d, maxRetry)
}
