package httptransport

import (
	"crypto/subtle"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		expectedAPIKey := strings.TrimSpace(os.Getenv("INTERNAL_API_KEY"))
		if expectedAPIKey == "" {
			c.AbortWithStatus(http.StatusServiceUnavailable)
			return
		}

		receivedAPIKey := strings.TrimSpace(c.GetHeader("X-Internal-API-Key"))
		if receivedAPIKey == "" {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		if subtle.ConstantTimeCompare([]byte(receivedAPIKey), []byte(expectedAPIKey)) != 1 {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		c.Next()
	}
}
