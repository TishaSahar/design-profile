package middleware

import (
	"net/http"
	"strings"

	"design-profile/backend/internal/auth"

	"github.com/gin-gonic/gin"
)

const claimsKey = "claims"

// RequireAuth validates the Bearer JWT in the Authorization header.
func RequireAuth(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authorization header required"})
			return
		}

		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization format"})
			return
		}

		claims, err := auth.ValidateToken(parts[1], jwtSecret)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		c.Set(claimsKey, claims)
		c.Next()
	}
}

// GetClaims retrieves the JWT claims stored by RequireAuth.
func GetClaims(c *gin.Context) *auth.Claims {
	val, _ := c.Get(claimsKey)
	claims, _ := val.(*auth.Claims)
	return claims
}
