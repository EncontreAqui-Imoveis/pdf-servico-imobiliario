package service

import (
	"bytes"
	"testing"
	"time"

	"pdf-service/internal/domain"
)

func TestGenerateProposalReturnsPDFBytesForValidRequest(t *testing.T) {
	svc := NewPDFService()

	req := domain.ProposalRequest{
		ClientName:    "Ana Silva",
		ClientCPF:     "123.456.789-00",
		BrokerName:    "Pedro Souza",
		SellingBroker: "Maria Souza",
		PropertyAddress: domain.FlexibleAddress{
			Street:       "Rua A",
			Number:       "10",
			Neighborhood: "Centro",
			City:         "Goiânia",
			State:        "GO",
		},
		PropertyCity:  "Goiânia",
		PropertyState: "GO",
		TotalValue:    150000,
		Payment: domain.PaymentBreakdown{
			Cash:      50000,
			Financing: 100000,
		},
		ValidityDays: 10,
	}

	pdfBytes, err := svc.GenerateProposal(req)
	if err != nil {
		t.Fatalf("expected valid PDF generation, got error: %v", err)
	}
	if len(pdfBytes) == 0 {
		t.Fatal("expected generated PDF bytes, got empty output")
	}
	if !bytes.HasPrefix(pdfBytes, []byte("%PDF")) {
		t.Fatalf("expected PDF signature prefix, got %q", pdfBytes[:4])
	}
}

func TestGenerateProposalCompletesWithinNominalBudget(t *testing.T) {
	svc := NewPDFService()

	req := domain.ProposalRequest{
		ClientName:    "Ana Silva",
		ClientCPF:     "123.456.789-00",
		BrokerName:    "Pedro Souza",
		SellingBroker: "Maria Souza",
		PropertyAddress: domain.FlexibleAddress{
			Street:       "Rua A",
			Number:       "10",
			Neighborhood: "Centro",
			City:         "Goiânia",
			State:        "GO",
		},
		PropertyCity:  "Goiânia",
		PropertyState: "GO",
		TotalValue:    150000,
		Payment: domain.PaymentBreakdown{
			Cash:      50000,
			Financing: 100000,
		},
		ValidityDays: 10,
	}

	startedAt := time.Now()
	_, err := svc.GenerateProposal(req)
	elapsed := time.Since(startedAt)

	if err != nil {
		t.Fatalf("expected valid PDF generation, got error: %v", err)
	}
	if elapsed > 1500*time.Millisecond {
		t.Fatalf("expected PDF generation under 1500ms, got %s", elapsed)
	}
}
