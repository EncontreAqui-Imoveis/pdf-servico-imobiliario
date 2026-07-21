package service

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"pdf-service/internal/domain"
)

func TestGenerateProposalReturnsPDFBytesForValidRequest(t *testing.T) {
	svc := NewPDFService()

	req := domain.ProposalRequest{
		ClientName: "Ana Silva",
		ClientCPF:  "123.456.789-00",
		BrokerName: "Pedro Souza",
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
		ClientName: "Ana Silva",
		ClientCPF:  "123.456.789-00",
		BrokerName: "Pedro Souza",
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

	if got != "Encontre Aqui Imóveis Ltda (Imobiliária)" {
		t.Fatalf("expected institutional signature label, got %q", got)
	}
}

// Regressão: captador/corretor não entra no corpo do PDF; legado "vendedor" não deve vazar no binário.
func TestGenerateProposalDoesNotEmbedBrokerNameInPDFBytes(t *testing.T) {
	svc := NewPDFService()
	uniqueBroker := "BROKER_MUST_NOT_APPEAR_IN_PDF_77219"

	req := domain.ProposalRequest{
		ClientName: "Comprador Proponente",
		ClientCPF:  "123.456.789-00",
		BrokerName: uniqueBroker,
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

func TestBuildFooterBrandLabelUsesSeparatedBrandName(t *testing.T) {
	got := buildFooterBrandLabel()

	if got != "Encontre Aqui Imóveis" {
		t.Fatalf("expected footer brand label to be separated, got %q", got)
	}
}

func TestBuildInstitutionalAddresseeLabelUsesUppercaseBrand(t *testing.T) {
	got := buildInstitutionalAddresseeLabel()

	if got != "ENCONTRE AQUI IMÓVEIS LTDA" {
		t.Fatalf("expected addressee label to be uppercase, got %q", got)
	}
}

func TestBuildProposalTitleUsesRentForRentalProposals(t *testing.T) {
	req := domain.ProposalRequest{
		DealType: "rent",
	}

	got := buildProposalTitle(req)

	if got != "PROPOSTA DE LOCAÇÃO DE IMÓVEL" {
		t.Fatalf("expected rent proposal title, got %q", got)
	}
}

func TestBuildIntroParagraphUsesRentVocabularyForRentalProposals(t *testing.T) {
	req := domain.ProposalRequest{
		DealType: "rent",
	}

	got := buildIntroParagraph(req, "Rua A, 10", "Rio Verde", "GO")

	if !strings.Contains(got, "oferta de locação") {
		t.Fatalf("expected rental intro paragraph, got %q", got)
	}
	if strings.Contains(got, "oferta de compra") {
		t.Fatalf("expected rental intro paragraph to avoid purchase wording, got %q", got)
	}
}

func TestBuildRentalProposalTermsUsesOnlyRentalVocabulary(t *testing.T) {
	lines := buildRentalProposalTerms(domain.RentalTerms{
		MonthlyRent:               2500,
		GuaranteeType:             "Seguro-fiança",
		GuaranteeAmount:           2500,
		LeaseTermMonths:           30,
		ExpectedStartDate:         "2026-08-01",
		MonthlyDueDay:             10,
		CondominiumResponsibility: "Locatário",
		PropertyTaxResponsibility: "Locador",
		Observations:              "Sem animais.",
	})
	joined := strings.Join(lines, "\n")

	for _, expected := range []string{
		"Valor mensal do aluguel: R$ 2.500,00",
		"Garantia locatícia: Seguro-fiança (R$ 2.500,00)",
		"Prazo de locação: 30 meses",
		"Início previsto da locação: 01/08/2026",
		"Vencimento mensal: dia 10",
		"Responsabilidade pelo condomínio: Locatário",
		"Responsabilidade pelo IPTU: Locador",
	} {
		if !strings.Contains(joined, expected) {
			t.Fatalf("expected rental terms to contain %q, got %q", expected, joined)
		}
	}
	for _, forbidden := range []string{"Sinal/Entrada", "Financiamento", "Permuta"} {
		if strings.Contains(joined, forbidden) {
			t.Fatalf("rental terms must not contain sale term %q: %q", forbidden, joined)
		}
	}
}

func TestGenerateProposalRendersRentalTermsWithoutSaleTerms(t *testing.T) {
	svc := NewPDFService()
	req := domain.ProposalRequest{
		ClientName:      "Ana Silva",
		ClientCPF:       "123.456.789-00",
		BrokerName:      "Pedro Souza",
		DealType:        "rent",
		PropertyAddress: domain.FlexibleAddress{Raw: "Rua A, 10, Centro, Goiânia, GO"},
		RentalTerms: domain.RentalTerms{
			MonthlyRent:     2500,
			GuaranteeType:   "Caução",
			LeaseTermMonths: 12,
		},
		ValidityDays: 10,
	}

	pdfBytes, err := svc.GenerateProposal(req)
	if err != nil {
		t.Fatalf("expected rental PDF generation, got error: %v", err)
	}
	pdfText := string(pdfBytes)
	if !strings.Contains(pdfText, "Valor mensal do aluguel") {
		t.Fatalf("expected rental PDF content, got %q", pdfText)
	}
	for _, forbidden := range []string{"Sinal/Entrada", "Financiamento", "Permuta"} {
		if strings.Contains(pdfText, forbidden) {
			t.Fatalf("rental PDF must not contain sale term %q", forbidden)
		}
	}
}
