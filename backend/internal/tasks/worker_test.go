package tasks

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"design-profile/backend/internal/model"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// fakeStore hands out its tasks once and records what happened to each.
type fakeStore struct {
	mu      sync.Mutex
	tasks   []model.Task
	deleted []uuid.UUID
	errored map[uuid.UUID]time.Time
	fatal   map[uuid.UUID]string
}

func newFakeStore(tasks ...model.Task) *fakeStore {
	return &fakeStore{tasks: tasks, errored: map[uuid.UUID]time.Time{}, fatal: map[uuid.UUID]string{}}
}

func (s *fakeStore) add(t model.Task) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tasks = append(s.tasks, t)
}

func (s *fakeStore) Claim(_ context.Context, limit int, _ time.Duration) ([]model.Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	n := min(limit, len(s.tasks))
	batch := s.tasks[:n]
	s.tasks = s.tasks[n:]
	return batch, nil
}
func (s *fakeStore) Delete(_ context.Context, id uuid.UUID) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.deleted = append(s.deleted, id)
	return nil
}
func (s *fakeStore) MarkError(_ context.Context, id uuid.UUID, _ string, runAt time.Time) error {
	s.errored[id] = runAt
	return nil
}
func (s *fakeStore) MarkFatal(_ context.Context, id uuid.UUID, reason string) error {
	s.fatal[id] = reason
	return nil
}

func task(taskType string, attempts int) model.Task {
	return model.Task{ID: uuid.New(), Type: taskType, Attempts: attempts, Body: []byte(`{}`)}
}

var testSettings = Settings{Delay: time.Second, ErrorDelay: 3 * time.Minute, NoTasksDelay: 20 * time.Second}

// runAll steps the worker until no task is due and returns the pause after each step.
func runAll(w *Worker) []time.Duration {
	var pauses []time.Duration
	for {
		o := w.Step(context.Background())
		pauses = append(pauses, w.pause(o))
		if o == noTasks {
			return pauses
		}
	}
}

func TestStep_OutcomeDecidesTaskFateAndPause(t *testing.T) {
	ok, temporary, permanent, unknown := task("ok", 1), task("temporary", 3), task("permanent", 1), task("unknown", 1)
	store := newFakeStore(ok, temporary, permanent, unknown)
	w := NewWorker(store, testSettings)
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	w.now = func() time.Time { return now }
	w.Handle("ok", func(context.Context, []byte) error { return nil })
	w.Handle("temporary", func(context.Context, []byte) error { return errors.New("server busy") })
	w.Handle("permanent", func(context.Context, []byte) error { return Permanent(errors.New("bad address")) })

	pauses := runAll(w)

	assert.Equal(t, []uuid.UUID{ok.ID}, store.deleted, "a task that succeeded is deleted")
	assert.Equal(t, now.Add(testSettings.ErrorDelay), store.errored[temporary.ID], "retried after error_delay")
	assert.Contains(t, store.fatal[permanent.ID], "bad address")
	assert.Contains(t, store.fatal[unknown.ID], "no handler")
	assert.Equal(t, []time.Duration{
		testSettings.Delay,        // ok
		testSettings.ErrorDelay,   // temporary failure
		testSettings.Delay,        // fatal: the next task need not wait longer
		testSettings.Delay,        // fatal (no handler)
		testSettings.NoTasksDelay, // nothing left
	}, pauses)
}

func TestStep_OneTaskAtATime(t *testing.T) {
	store := newFakeStore(task("ok", 1), task("ok", 1), task("ok", 1))
	w := NewWorker(store, testSettings)
	w.Handle("ok", func(context.Context, []byte) error { return nil })

	assert.Equal(t, done, w.Step(context.Background()))
	assert.Len(t, store.deleted, 1)
	assert.Len(t, store.tasks, 2, "the others wait for the next step")
}

func TestRun_WakeSkipsTheWait(t *testing.T) {
	store := newFakeStore()
	w := NewWorker(store, Settings{NoTasksDelay: time.Hour})
	ran := make(chan struct{}, 1)
	w.Handle("ok", func(context.Context, []byte) error { ran <- struct{}{}; return nil })
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go w.Run(ctx)

	time.Sleep(50 * time.Millisecond) // the worker found nothing and waits an hour
	store.add(task("ok", 1))
	w.Wake()

	select {
	case <-ran:
	case <-time.After(2 * time.Second):
		t.Fatal("a queued task did not wake the worker")
	}
}

func TestWake_NeverBlocks(t *testing.T) {
	w := NewWorker(newFakeStore(), testSettings)
	w.Wake()
	w.Wake() // the first wake-up is still pending
}
