package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"

	"pdf-service/internal/service"
	httptransport "pdf-service/internal/transport/http"
)

func main() {
	router := gin.Default()
	router.SetTrustedProxies(nil)
	router.Use(httptransport.AuthMiddleware())

	pdfService := service.NewPDFService()
	handler := httptransport.NewHandler(pdfService)

	router.POST("/generate-proposal", handler.GenerateProposal)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	if err := router.Run(":" + port); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
