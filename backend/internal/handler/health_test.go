package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubPinger struct {
	err      error
	deadline bool
}

func (p *stubPinger) Ping(ctx context.Context) error {
	_, p.deadline = ctx.Deadline()
	return p.err
}

func getHealth(t *testing.T, db pinger) (int, HealthResponse) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/health", NewHealthHandler(db).Health)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/health", nil))

	var body HealthResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	return w.Code, body
}

func TestHealth(t *testing.T) {
	t.Run("ok when postgres answers", func(t *testing.T) {
		db := &stubPinger{}
		code, body := getHealth(t, db)
		assert.Equal(t, http.StatusOK, code)
		assert.Equal(t, "ok", body.Status)
		assert.Equal(t, "ok", body.Checks["postgres"])
		assert.True(t, db.deadline, "ping must be bounded by a timeout")
	})

	t.Run("503 without error details when postgres is down", func(t *testing.T) {
		db := &stubPinger{err: errors.New("dial tcp 10.0.0.5:5432: connect: connection refused")}
		code, body := getHealth(t, db)
		assert.Equal(t, http.StatusServiceUnavailable, code)
		assert.Equal(t, "unavailable", body.Status)
		assert.Equal(t, "unavailable", body.Checks["postgres"])
	})
}
