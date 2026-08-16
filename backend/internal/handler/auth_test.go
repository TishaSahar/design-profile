package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"design-profile/backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ── Stub auth service ────────────────────────────────────────────────────────

type stubAuthSvc struct {
	requestOTPErr error
	verifyOTPJWT  string
	verifyOTPErr  error
}

func (s *stubAuthSvc) RequestOTP(_ context.Context, _ string) error {
	return s.requestOTPErr
}

func (s *stubAuthSvc) VerifyOTP(_ context.Context, _, _ string) (string, error) {
	return s.verifyOTPJWT, s.verifyOTPErr
}

// ── Test helpers ─────────────────────────────────────────────────────────────

func newAuthTestRouter(h *AuthHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/auth/request-otp", h.RequestOTP)
	r.POST("/auth/verify-otp", h.VerifyOTP)
	return r
}

func jsonBody(t *testing.T, v any) *bytes.Reader {
	t.Helper()
	b, err := json.Marshal(v)
	require.NoError(t, err)
	return bytes.NewReader(b)
}

// ── RequestOTP handler ───────────────────────────────────────────────────────

func TestRequestOTPHandler_ValidEmail_Returns200(t *testing.T) {
	r := newAuthTestRouter(NewAuthHandler(&stubAuthSvc{}))

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/auth/request-otp",
		jsonBody(t, map[string]string{"email": "admin@example.com"}))
	req.Header.Set("Content-Type", "application/json")

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRequestOTPHandler_ServiceError_StillReturns200(t *testing.T) {
	// The handler must always return 200, even when the service fails internally.
	svc := &stubAuthSvc{requestOTPErr: assert.AnError}
	r := newAuthTestRouter(NewAuthHandler(svc))

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/auth/request-otp",
		jsonBody(t, map[string]string{"email": "admin@example.com"}))
	req.Header.Set("Content-Type", "application/json")

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "service error must not leak to client")
}

func TestRequestOTPHandler_InvalidEmail_Returns400(t *testing.T) {
	r := newAuthTestRouter(NewAuthHandler(&stubAuthSvc{}))

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/auth/request-otp",
		jsonBody(t, map[string]string{"email": "not-an-email"}))
	req.Header.Set("Content-Type", "application/json")

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRequestOTPHandler_MissingEmail_Returns400(t *testing.T) {
	r := newAuthTestRouter(NewAuthHandler(&stubAuthSvc{}))

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/auth/request-otp",
		jsonBody(t, map[string]string{}))
	req.Header.Set("Content-Type", "application/json")

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ── VerifyOTP handler ────────────────────────────────────────────────────────

func TestVerifyOTPHandler_ValidCode_Returns200WithToken(t *testing.T) {
	svc := &stubAuthSvc{verifyOTPJWT: "signed.jwt.token"}
	r := newAuthTestRouter(NewAuthHandler(svc))

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/auth/verify-otp",
		jsonBody(t, map[string]string{"email": "admin@example.com", "code": "123456"}))
	req.Header.Set("Content-Type", "application/json")

	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var resp map[string]map[string]string
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "signed.jwt.token", resp["data"]["token"])
}

func TestVerifyOTPHandler_InvalidCode_Returns401(t *testing.T) {
	svc := &stubAuthSvc{verifyOTPErr: service.ErrInvalidOTP}
	r := newAuthTestRouter(NewAuthHandler(svc))

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/auth/verify-otp",
		jsonBody(t, map[string]string{"email": "admin@example.com", "code": "000000"}))
	req.Header.Set("Content-Type", "application/json")

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestVerifyOTPHandler_CodeNot6Digits_Returns400(t *testing.T) {
	r := newAuthTestRouter(NewAuthHandler(&stubAuthSvc{}))

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/auth/verify-otp",
		jsonBody(t, map[string]string{"email": "admin@example.com", "code": "123"}))
	req.Header.Set("Content-Type", "application/json")

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestVerifyOTPHandler_MissingCode_Returns400(t *testing.T) {
	r := newAuthTestRouter(NewAuthHandler(&stubAuthSvc{}))

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/auth/verify-otp",
		jsonBody(t, map[string]string{"email": "admin@example.com"}))
	req.Header.Set("Content-Type", "application/json")

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
