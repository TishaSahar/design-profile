package handler

import (
	"errors"
	"log/slog"
	"net/http"

	"design-profile/backend/internal/service"

	"github.com/gin-gonic/gin"
)

// AuthHandler handles authentication endpoints.
type AuthHandler struct {
	svc *service.AuthService
}

func NewAuthHandler(svc *service.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

type requestOTPBody struct {
	Email string `json:"email" binding:"required,email"`
}

type verifyOTPBody struct {
	Email string `json:"email" binding:"required,email"`
	Code  string `json:"code"  binding:"required,len=6"`
}

// RequestOTP godoc
// @Summary      Request OTP
// @Description  Sends a one-time password to the admin email address. Always returns 200 to prevent email enumeration.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      requestOTPBody  true  "Email address"
// @Success      200   {object}  map[string]string
// @Failure      400   {object}  map[string]string
// @Router       /auth/request-otp [post]
func (h *AuthHandler) RequestOTP(c *gin.Context) {
	var body requestOTPBody
	if err := c.ShouldBindJSON(&body); err != nil {
		badRequest(c, err.Error())
		return
	}
	// Ignore error — always return 200 to prevent email enumeration.
	_ = h.svc.RequestOTP(c.Request.Context(), body.Email)
	c.JSON(http.StatusOK, gin.H{"message": "if the email is registered, a code has been sent"})
}

// VerifyOTP godoc
// @Summary      Verify OTP
// @Description  Verifies the one-time password and returns a JWT token.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      verifyOTPBody   true  "Email and OTP code"
// @Success      200   {object}  map[string]string
// @Failure      400   {object}  map[string]string
// @Failure      401   {object}  map[string]string
// @Router       /auth/verify-otp [post]
func (h *AuthHandler) VerifyOTP(c *gin.Context) {
	var body verifyOTPBody
	if err := c.ShouldBindJSON(&body); err != nil {
		badRequest(c, err.Error())
		return
	}

	token, err := h.svc.VerifyOTP(c.Request.Context(), body.Email, body.Code)
	if err != nil {
		if errors.Is(err, service.ErrInvalidOTP) {
			slog.Warn("неверный OTP при попытке входа",
				"ip", c.ClientIP(),
				"email", body.Email,
			)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired code"})
			return
		}
		internalError(c, "authentication failed")
		return
	}
	ok(c, gin.H{"token": token})
}
