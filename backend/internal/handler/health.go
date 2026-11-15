package handler

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

const healthCheckTimeout = 2 * time.Second

// pinger is satisfied by *pgxpool.Pool.
type pinger interface {
	Ping(ctx context.Context) error
}

// HealthResponse reports the overall status and the status of each dependency.
type HealthResponse struct {
	Status string            `json:"status"          example:"ok"`
	Checks map[string]string `json:"checks"`
}

// HealthHandler reports whether the API can serve requests.
type HealthHandler struct {
	db pinger
}

func NewHealthHandler(db pinger) *HealthHandler {
	return &HealthHandler{db: db}
}

// Health godoc
// @Summary      Health check
// @Description  Reports whether the API and its dependencies (PostgreSQL) are available. Returns 503 when a dependency is down. Also served at /health outside the API prefix.
// @Tags         health
// @Produce      json
// @Success      200  {object}  HealthResponse
// @Failure      503  {object}  HealthResponse
// @Router       /health [get]
func (h *HealthHandler) Health(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), healthCheckTimeout)
	defer cancel()

	resp := HealthResponse{Status: "ok", Checks: map[string]string{"postgres": "ok"}}
	code := http.StatusOK

	if err := h.db.Ping(ctx); err != nil {
		// The cause stays in the server log; the public response only says what is down.
		slog.Error("health check: postgres ping failed", "err", err)
		resp.Status = "unavailable"
		resp.Checks["postgres"] = "unavailable"
		code = http.StatusServiceUnavailable
	}

	c.JSON(code, resp)
}
