package domain

import (
	"strings"
	"testing"
)

func TestProposalRequestSanitizeRemovesUnsafeCharacters(t *testing.T) {
	req := ProposalRequest{
		ClientName:          "  Ana \u202e Silva  ",
		ClientCPF:           "123.456.789-00",
		BrokerName:          " Pedro \u200f Souza ",
		PropertyCity:        " Goiânia ",
		PropertyState:       " go ",
		PaymentMethodLegacy: " Dinheiro: R$ 100,00 ",
		PropertyAddress: FlexibleAddress{
			Street: " Rua 1 ",
			Number: " 10 ",
			City:   " Goiânia ",
			State:  " go ",
		},
	}

	req.Sanitize()

	if got := req.ClientName; got != "Ana Silva" {
		t.Fatalf("expected sanitized client name %q, got %q", "Ana Silva", got)
	}
	if got := req.BrokerName; got != "Pedro Souza" {
		t.Fatalf("expected sanitized broker name %q, got %q", "Pedro Souza", got)
	}
	if got := req.PropertyState; got != "GO" {
		t.Fatalf("expected sanitized state %q, got %q", "GO", got)
	}
	if got := req.PropertyAddress.State; got != "GO" {
		t.Fatalf("expected sanitized address state %q, got %q", "GO", got)
	}
}

func TestProposalRequestValidateRejectsClientNameAboveMaxLength(t *testing.T) {
	req := ProposalRequest{
		ClientName:    strings.Repeat("A", maxClientNameLength+1),
		PropertyCity:  "Goiânia",
		PropertyState: "GO",
		PropertyAddress: FlexibleAddress{
			Street: "Rua A",
			Number: "10",
			City:   "Goiânia",
			State:  "GO",
		},
		BrokerName:   "Pedro",
		TotalValue:   100,
		Payment:      PaymentBreakdown{Cash: 100},
		ValidityDays: 10,
	}

	err := req.Validate()
	if err == nil {
		t.Fatal("expected validation error for oversized client name")
	}
	if !strings.Contains(err.Error(), "client_name exceeds max length") {
		t.Fatalf("unexpected validation error: %v", err)
	}
}
