package tasks

import (
	"context"
	"errors"
	"testing"
	"time"

	"design-profile/backend/internal/model"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// fakeStore hands out its tasks once and records what happened to each.
type fakeStore struct {
	tasks   []model.Task
	deleted []uuid.UUID
	errored map[uuid.UUID]time.Time
	fatal   map[uuid.UUID]string
}

func newFakeStore(tasks ...model.Task) *fakeStore {
	return &fakeStore{tasks: tasks, errored: map[uuid.UUID]time.Time{}, fatal: map[uuid.UUID]string{}}
}

func (s *fakeStore) Claim(_ context.Context, limit int, _ time.Duration) ([]model.Task, error) {
	n := min(limit, len(s.tasks))
	batch := s.tasks[:n]
	s.tasks = s.tasks[n:]
	return batch, nil
}
func (s *fakeStore) Delete(_ context.Context, id uuid.UUID) error {
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

func TestRunDue_OutcomeDecidesTaskFate(t *testing.T) {
	ok, temporary, permanent, unknown := task("ok", 1), task("temporary", 3), task("permanent", 1), task("unknown", 1)
	store := newFakeStore(ok, temporary, permanent, unknown)
	w := NewWorker(store)
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	w.now = func() time.Time { return now }
	w.Handle("ok", func(context.Context, []byte) error { return nil })
	w.Handle("temporary", func(context.Context, []byte) error { return errors.New("server busy") })
	w.Handle("permanent", func(context.Context, []byte) error { return Permanent(errors.New("bad address")) })

	w.RunDue(context.Background())

	assert.Equal(t, []uuid.UUID{ok.ID}, store.deleted, "a task that succeeded is deleted")
	assert.Equal(t, now.Add(4*time.Minute), store.errored[temporary.ID], "third failure waits 4 minutes")
	assert.Contains(t, store.fatal[permanent.ID], "bad address")
	assert.Contains(t, store.fatal[unknown.ID], "no handler")
	assert.Len(t, store.errored, 1)
	assert.Len(t, store.fatal, 2)
}

func TestRunDue_ProcessesEveryBatch(t *testing.T) {
	var all []model.Task
	for range batchSize*2 + 3 {
		all = append(all, task("ok", 1))
	}
	store := newFakeStore(all...)
	w := NewWorker(store)
	w.Handle("ok", func(context.Context, []byte) error { return nil })

	w.RunDue(context.Background())

	assert.Len(t, store.deleted, len(all))
}

func TestRetryDelay_DoublesUpToAnHour(t *testing.T) {
	assert.Equal(t, time.Minute, retryDelay(1))
	assert.Equal(t, 2*time.Minute, retryDelay(2))
	assert.Equal(t, 32*time.Minute, retryDelay(6))
	assert.Equal(t, time.Hour, retryDelay(7))
	assert.Equal(t, time.Hour, retryDelay(500))
}

func TestWake_NeverBlocks(t *testing.T) {
	w := NewWorker(newFakeStore())
	w.Wake()
	w.Wake() // the first wake-up is still pending
}
