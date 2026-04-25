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

func TestBuildProponentSignatureLabelUsesClientName(t *testing.T) {
	got := buildProponentSignatureLabel("Joana Pereira")

	if got != "Joana Pereira (Proponente)" {
		t.Fatalf("expected proponent signature label to use client name, got %q", got)
	}
}

func TestBuildInstitutionalSignatureLabelUsesInstitutionalParty(t *testing.T) {
	got := buildInstitutionalSignatureLabel()

	if got != "ENCONTREAQUI IMÓVEIS LTDA (Imobiliária)" {
		t.Fatalf("expected institutional signature label, got %q", got)
	}
}

// Regressão: captador/corretor não entra no corpo do PDF; legado "vendedor" não deve vazar no binário.
func TestGenerateProposalDoesNotEmbedBrokerNameInPDFBytes(t *testing.T) {
	svc := NewPDFService()
	uniqueBroker := "BROKER_MUST_NOT_APPEAR_IN_PDF_77219"

	req := domain.ProposalRequest{
		ClientName:    "Comprador Proponente",
		ClientCPF:     "123.456.789-00",
		BrokerName:    uniqueBroker,
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
	if bytes.Contains(pdfBytes, []byte(uniqueBroker)) {
		t.Fatalf("PDF must not contain broker string %q (template must stay non-leaky for legacy vendedor/captador)", uniqueBroker)
	}
}
