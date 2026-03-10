package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"

	"pdf-service/internal/config"
	"pdf-service/internal/service"
	httptransport "pdf-service/internal/transport/http"
)

func main() {
	if config.InternalAPIKey() == "" {
		log.Fatal("INTERNAL_API_KEY or PDF_INTERNAL_API_KEY must be configured")
	}

	router := gin.Default()
	if err := router.SetTrustedProxies(nil); err != nil {
		log.Fatalf("failed to configure trusted proxies: %v", err)
	}
	router.Use(httptransport.AuthMiddleware())

	pdfService := service.NewPDFService()
	handler := httptransport.NewHandler(pdfService)

	router.POST("/generate-proposal", handler.GenerateProposal)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           router,
		ReadTimeout:       5 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       15 * time.Second,
		MaxHeaderBytes:    1 << 20, // 1 MB
	}

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("failed to start server: %v", err)
	}
}
