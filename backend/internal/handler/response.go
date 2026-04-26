package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ok sends a 200 JSON response with the given data payload.
func ok(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, gin.H{"data": data})
}

// created sends a 201 JSON response.
func created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, gin.H{"data": data})
}

// badRequest sends a 400 JSON error response.
func badRequest(c *gin.Context, msg string) {
	c.JSON(http.StatusBadRequest, gin.H{"error": msg})
}

// notFound sends a 404 JSON error response.
func notFound(c *gin.Context, msg string) {
	c.JSON(http.StatusNotFound, gin.H{"error": msg})
}

// internalError sends a 500 JSON error response.
func internalError(c *gin.Context, msg string) {
	c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
}
