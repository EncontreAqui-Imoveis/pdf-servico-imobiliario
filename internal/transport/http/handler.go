package httptransport

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"pdf-service/internal/domain"
)

type ProposalPDFService interface {
	GenerateProposal(req domain.ProposalRequest) ([]byte, error)
}

type Handler struct {
	pdfService ProposalPDFService
}

func NewHandler(pdfService ProposalPDFService) *Handler {
	return &Handler{pdfService: pdfService}
}

func (h *Handler) GenerateProposal(c *gin.Context) {
	var req domain.ProposalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	if err := req.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	pdfBytes, err := h.pdfService.GenerateProposal(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate pdf"})
		return
	}

	c.Data(http.StatusOK, "application/pdf", pdfBytes)
}
