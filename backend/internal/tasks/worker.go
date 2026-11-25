// Package tasks runs the background tasks of the outbox table: it picks up due
// tasks one at a time, hands each to the handler of its type and records the
// outcome — deleted on success, retried after Settings.ErrorDelay on a temporary
// failure (status error), given up on a permanent one (status fatal).
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

// lease: a claimed task is not taken again for this long, so if the backend stops
// mid-task the task still runs later — possibly twice, never zero times.
const lease = 5 * time.Minute

// outcome is what happened in one step of the worker; it decides the pause after it.
type outcome int

const (
	noTasks outcome = iota
	done            // succeeded, or failed for good (fatal)
	failed          // failed for a temporary reason, or the table could not be read
)

// Worker runs due tasks until its context is cancelled.
type Worker struct {
	store    store
	settings Settings
	handlers map[string]Handler
	wake     chan struct{}
	now      func() time.Time
}

func NewWorker(store store, settings Settings) *Worker {
	return &Worker{
		store:    store,
		settings: settings,
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

// Run processes due tasks one at a time until ctx is cancelled, pausing after
// each step as its outcome and the Settings say.
func (w *Worker) Run(ctx context.Context) {
	for ctx.Err() == nil {
		pause := w.pause(w.Step(ctx))
		timer := time.NewTimer(pause)
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
		case <-w.wake:
			timer.Stop()
		}
	}
}

func (w *Worker) pause(o outcome) time.Duration {
	switch o {
	case noTasks:
		return w.settings.NoTasksDelay
	case failed:
		return w.settings.ErrorDelay
	default:
		return w.settings.Delay
	}
}

// Step runs the next due task, if any.
func (w *Worker) Step(ctx context.Context) outcome {
	batch, err := w.store.Claim(ctx, 1, lease)
	if err != nil {
		slog.Error("claim tasks", "err", err)
		return failed
	}
	if len(batch) == 0 {
		return noTasks
	}
	return w.run(ctx, batch[0])
}

func (w *Worker) run(ctx context.Context, t model.Task) outcome {
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
			return failed
		}
		log.Info("task done")
		return done
	case errors.As(err, &permanent):
		log.Error("task failed for good", "err", err)
		if err := w.store.MarkFatal(ctx, t.ID, err.Error()); err != nil {
			log.Error("mark task fatal", "err", err)
		}
		return done
	default:
		next := w.now().Add(w.settings.ErrorDelay)
		log.Warn("task failed, will retry", "err", err, "retry_at", next)
		if err := w.store.MarkError(ctx, t.ID, err.Error(), next); err != nil {
			log.Error("mark task error", "err", err)
		}
		return failed
	}
}
