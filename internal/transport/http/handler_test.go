package httptransport

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"pdf-service/internal/domain"
)

type stubProposalPDFService struct {
	receivedReq domain.ProposalRequest
	response    []byte
	err         error
}

func (s *stubProposalPDFService) GenerateProposal(
	req domain.ProposalRequest,
) ([]byte, error) {
	s.receivedReq = req
	if s.err != nil {
		return nil, s.err
	}
	return s.response, nil
}

func TestGenerateProposalRejectsOversizedPayload(t *testing.T) {
	gin.SetMode(gin.TestMode)

	service := &stubProposalPDFService{response: []byte("%PDF-1.4")}
	handler := NewHandler(service)

	router := gin.New()
	router.POST("/generate-proposal", handler.GenerateProposal)

	oversizedName := strings.Repeat("A", int(maxProposalPayloadBytes)+256)
	payload := fmt.Sprintf(
		`{"clientName":"%s","propertyAddress":"Rua A, 10","propertyCity":"Goiânia","propertyState":"GO","brokerName":"Pedro","totalValue":100,"payment":{"cash":100},"validadeDias":10}`,
		oversizedName,
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/generate-proposal",
		strings.NewReader(payload),
	)
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected status %d, got %d", http.StatusRequestEntityTooLarge, res.Code)
	}
	if service.receivedReq.ClientName != "" {
		t.Fatalf("expected service not to receive payload, got %q", service.receivedReq.ClientName)
	}
}

func TestGenerateProposalRejectsInvalidPayload(t *testing.T) {
	gin.SetMode(gin.TestMode)

	service := &stubProposalPDFService{response: []byte("%PDF-1.4")}
	handler := NewHandler(service)

	router := gin.New()
	router.POST("/generate-proposal", handler.GenerateProposal)

	req := httptest.NewRequest(
		http.MethodPost,
		"/generate-proposal",
		strings.NewReader(`{"clientName":`),
	)
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, res.Code)
	}
	if body := res.Body.String(); !strings.Contains(body, "invalid payload") {
		t.Fatalf("expected invalid payload message, got %q", body)
	}
	if service.receivedReq.ClientName != "" {
		t.Fatalf("expected service not to receive payload, got %q", service.receivedReq.ClientName)
	}
}

func TestGenerateProposalRejectsValidationErrorsBeforeCallingService(t *testing.T) {
	gin.SetMode(gin.TestMode)

	service := &stubProposalPDFService{response: []byte("%PDF-1.4")}
	handler := NewHandler(service)

	router := gin.New()
	router.POST("/generate-proposal", handler.GenerateProposal)

	payload := `{
		"clientName":"Ana Silva",
		"propertyCity":"Goiânia",
		"propertyState":"GO",
		"propertyAddress":{"street":"Rua A","number":"10","city":"Goiânia","state":"GO"},
		"brokerName":"Pedro",
		"totalValue":100,
		"payment":{"cash":90},
		"validadeDias":10
	}`

	req := httptest.NewRequest(
		http.MethodPost,
		"/generate-proposal",
		strings.NewReader(payload),
	)
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, res.Code)
	}
	if body := res.Body.String(); !strings.Contains(body, "payment breakdown must match total value") {
		t.Fatalf("expected validation error payload, got %q", body)
	}
	if service.receivedReq.ClientName != "" {
		t.Fatalf("expected service not to receive payload, got %q", service.receivedReq.ClientName)
	}
}

func TestGenerateProposalSanitizesPayloadBeforeCallingService(t *testing.T) {
	gin.SetMode(gin.TestMode)

	service := &stubProposalPDFService{response: []byte("%PDF-1.4 fake")}
	handler := NewHandler(service)

	router := gin.New()
	router.POST("/generate-proposal", handler.GenerateProposal)

	payload := `{
		"clientName":"  Ana \u202e Silva  ",
		"clientCpf":"123.456.789-00",
		"propertyAddress":{"street":" Rua 1 ","number":"10","city":" Goiânia ","state":"go"},
		"brokerName":" Pedro \u200f Souza ",
		"sellingBrokerName":" Maria ",
		"totalValue":100,
		"payment":{"cash":100},
		"validadeDias":10
	}`

	req := httptest.NewRequest(
		http.MethodPost,
		"/generate-proposal",
		strings.NewReader(payload),
	)
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, res.Code)
	}
	if got := service.receivedReq.ClientName; got != "Ana Silva" {
		t.Fatalf("expected sanitized client name %q, got %q", "Ana Silva", got)
	}
	if got := service.receivedReq.BrokerName; got != "Pedro Souza" {
		t.Fatalf("expected sanitized broker name %q, got %q", "Pedro Souza", got)
	}
	if got := service.receivedReq.PropertyAddress.State; got != "GO" {
		t.Fatalf("expected normalized state %q, got %q", "GO", got)
	}
	if got := res.Header().Get("Content-Disposition"); got == "" {
		t.Fatal("expected content disposition header to be set")
	}
	if body := res.Body.String(); !strings.HasPrefix(body, "%PDF-1.4") {
		t.Fatalf("expected PDF body, got %q", body)
	}
}

func TestGenerateProposalAcceptsBackendSnakeCasePayloadShape(t *testing.T) {
	gin.SetMode(gin.TestMode)

	service := &stubProposalPDFService{response: []byte("%PDF-1.4 legacy")}
	handler := NewHandler(service)

	router := gin.New()
	router.POST("/generate-proposal", handler.GenerateProposal)

	payload := `{
		"client_name":"Ana Silva",
		"client_cpf":"123.456.789-00",
		"property_address":"Rua A, 10, Centro, Goiânia, GO",
		"broker_name":"Pedro",
		"selling_broker_name":"Maria",
		"value":250000,
		"payment":{"cash":50000,"trade_in":25000,"financing":150000,"others":25000},
		"validity_days":10
	}`

	req := httptest.NewRequest(
		http.MethodPost,
		"/generate-proposal",
		strings.NewReader(payload),
	)
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, res.Code)
	}
	if got := service.receivedReq.ResolvedClientName(); got != "Ana Silva" {
		t.Fatalf("expected client name %q, got %q", "Ana Silva", got)
	}
	if got := service.receivedReq.ResolvedPropertyAddress(); got != "Rua A, 10, Centro, Goiânia, GO" {
		t.Fatalf("expected property address %q, got %q", "Rua A, 10, Centro, Goiânia, GO", got)
	}
	payments := service.receivedReq.ResolvedPayments()
	if payments.TradeIn != 25000 {
		t.Fatalf("expected trade-in %v, got %v", 25000, payments.TradeIn)
	}
	if got := service.receivedReq.ResolvedValidityDays(); got != 10 {
		t.Fatalf("expected validity %d, got %d", 10, got)
	}
	if body := res.Body.String(); !strings.HasPrefix(body, "%PDF-1.4") {
		t.Fatalf("expected PDF body, got %q", body)
	}
}

func TestGenerateProposalReturnsServerErrorWhenPDFServiceFails(t *testing.T) {
	gin.SetMode(gin.TestMode)

	service := &stubProposalPDFService{err: errors.New("generator failed")}
	handler := NewHandler(service)

	router := gin.New()
	router.POST("/generate-proposal", handler.GenerateProposal)

	payload := `{
		"clientName":"Ana Silva",
		"propertyCity":"Goiânia",
		"propertyState":"GO",
		"propertyAddress":{"street":"Rua A","number":"10","city":"Goiânia","state":"GO"},
		"brokerName":"Pedro",
		"totalValue":100,
		"payment":{"cash":100},
		"validadeDias":10
	}`

	req := httptest.NewRequest(
		http.MethodPost,
		"/generate-proposal",
		strings.NewReader(payload),
	)
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, res.Code)
	}
	if body := res.Body.String(); !strings.Contains(body, "failed to generate pdf") {
		t.Fatalf("expected server error payload, got %q", body)
	}
}
