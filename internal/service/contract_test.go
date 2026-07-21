package service

import (
	"strings"
	"testing"

	"pdf-service/internal/domain"
)

func TestGenerateContractUsesRentalTerminology(t *testing.T) {
	pdf, err := NewPDFService().GenerateContract(domain.ContractRequest{
		ContractID:      "contract-1",
		DealType:        "rent",
		PropertyTitle:   "Casa de teste",
		PropertyAddress: "Rua A, 10, Goiânia, GO",
		Seller:          domain.ContractParty{Name: "Locador"},
		Buyer:           domain.ContractParty{Name: "Locatário"},
		RentalTerms:     domain.RentalTerms{MonthlyRent: 1500},
	})
	if err != nil {
		t.Fatalf("GenerateContract() error = %v", err)
	}
	text := string(pdf)
	if !strings.Contains(text, "MINUTA DE CONTRATO DE LOCA") || strings.Contains(text, "COMPRA E VENDA") {
		t.Fatalf("expected rental-only contract text, got %q", text)
	}
}

func TestGenerateContractUsesSaleTerminology(t *testing.T) {
	pdf, err := NewPDFService().GenerateContract(domain.ContractRequest{
		ContractID:      "contract-1",
		DealType:        "sale",
		PropertyTitle:   "Casa de teste",
		PropertyAddress: "Rua A, 10, Goiânia, GO",
		Seller:          domain.ContractParty{Name: "Vendedor"},
		Buyer:           domain.ContractParty{Name: "Comprador"},
		SaleTerms:       domain.PaymentBreakdown{Cash: 100000},
	})
	if err != nil {
		t.Fatalf("GenerateContract() error = %v", err)
	}
	text := string(pdf)
	if !strings.Contains(text, "MINUTA DE CONTRATO DE COMPRA E VENDA") || strings.Contains(text, "LOCA") {
		t.Fatalf("expected sale-only contract text, got %q", text)
	}
}
