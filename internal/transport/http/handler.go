package httptransport

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"pdf-service/internal/domain"
)

type PDFService interface {
	GenerateProposal(req domain.ProposalRequest) ([]byte, error)
	GenerateContract(req domain.ContractRequest) ([]byte, error)
}

type Handler struct {
	pdfService PDFService
}

const maxProposalPayloadBytes int64 = 1 << 20 // 1MB

func NewHandler(pdfService PDFService) *Handler {
	return &Handler{pdfService: pdfService}
}

func (h *Handler) GenerateContract(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxProposalPayloadBytes)

	var req domain.ContractRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "payload too large"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	if err := req.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	pdfBytes, err := h.pdfService.GenerateContract(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate pdf"})
		return
	}

	filename := "minuta_contrato_compra_venda.pdf"
	if req.DealType == "rent" {
		filename = "minuta_contrato_locacao.pdf"
	}
	c.Header("Content-Disposition", `attachment; filename="`+filename+`"`)
	c.Data(http.StatusOK, "application/pdf", pdfBytes)
}

func (h *Handler) GenerateProposal(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxProposalPayloadBytes)

	var req domain.ProposalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "payload too large"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	req.Sanitize()

	if err := req.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	pdfBytes, err := h.pdfService.GenerateProposal(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate pdf"})
		return
	}

	c.Header("Content-Disposition", `attachment; filename="proposta_compra_imovel.pdf"`)
	c.Data(http.StatusOK, "application/pdf", pdfBytes)
}
