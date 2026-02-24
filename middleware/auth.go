package middleware

import (
	"crypto/subtle"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

const (
	// APIKeyHeader is the header name used to pass the API key.
	APIKeyHeader = "X-API-Key"

	// APIKeyEnvVar is the environment variable name that holds the expected API key.
	APIKeyEnvVar = "API_KEY"
)

// APIKeyAuth returns a Gin middleware that validates the X-API-Key header
// against the value stored in the API_KEY environment variable.
//
// If API_KEY is not set, the server will reject all requests with 500.
// If the header is missing, it returns 401 Unauthorized.
// If the header value does not match, it returns 403 Forbidden.
func APIKeyAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		expectedKey := os.Getenv(APIKeyEnvVar)
		if expectedKey == "" {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "API key is not configured on the server",
			})
			c.Abort()
			return
		}

		providedKey := c.GetHeader(APIKeyHeader)
		if providedKey == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Missing API key. Please provide the X-API-Key header.",
			})
			c.Abort()
			return
		}

		// Use constant-time comparison to prevent timing attacks
		if subtle.ConstantTimeCompare([]byte(expectedKey), []byte(providedKey)) != 1 {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Invalid API key.",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
