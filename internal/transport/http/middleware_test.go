package httptransport

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestAuthMiddlewareRejectsWhenInternalAPIKeyIsNotConfigured(t *testing.T) {
	t.Setenv("INTERNAL_API_KEY", "")
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(AuthMiddleware())
	router.GET("/ping", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.Header.Set("X-Internal-API-Key", "super-secret-key")
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status %d, got %d", http.StatusServiceUnavailable, res.Code)
	}
}

func TestAuthMiddlewareRejectsMissingAPIKeyHeader(t *testing.T) {
	t.Setenv("INTERNAL_API_KEY", "super-secret-key")
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(AuthMiddleware())
	router.GET("/ping", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, res.Code)
	}
}

func TestAuthMiddlewareRejectsInvalidAPIKeyHeader(t *testing.T) {
	t.Setenv("INTERNAL_API_KEY", "super-secret-key")
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(AuthMiddleware())
	router.GET("/ping", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.Header.Set("X-Internal-API-Key", "wrong-key")
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, res.Code)
	}
}

func TestAuthMiddlewareAllowsRequestWithValidAPIKey(t *testing.T) {
	t.Setenv("INTERNAL_API_KEY", "super-secret-key")
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(AuthMiddleware())
	router.GET("/ping", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.Header.Set("X-Internal-API-Key", "super-secret-key")
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, res.Code)
	}
}
