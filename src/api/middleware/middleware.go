package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ErrorHandler middleware recovers from panics and returns a 500 error
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Log the error
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error": "Internal Server Error",
				})
			}
		}()
		c.Next()
	}
}

// AuthMiddleware checks for valid authentication
// To be expanded with actual auth implementation
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement proper authentication
		// This is just a placeholder for now
		c.Next()
	}
}
