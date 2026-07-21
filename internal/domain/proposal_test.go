package domain

import "testing"

func TestProposalRequestResolvedDealTypeUsesExplicitRent(t *testing.T) {
	req := ProposalRequest{
		DealType: "aluguel",
	}

	if got := req.ResolvedDealType(); got != "rent" {
		t.Fatalf("expected rent deal type, got %q", got)
	}
}

func TestProposalRequestResolvedDealTypeDefaultsToSale(t *testing.T) {
	req := ProposalRequest{}

	if got := req.ResolvedDealType(); got != "sale" {
		t.Fatalf("expected sale deal type fallback, got %q", got)
	}
}

func TestRentalProposalValidationRejectsInvalidCommercialTerms(t *testing.T) {
	req := ProposalRequest{
		ClientName:            "Ana Silva",
		PropertyAddressLegacy: "Rua A, 10",
		BrokerName:            "Corretor",
		DealType:              "rent",
		ValidityDays:          10,
		RentalTerms: RentalTerms{
			MonthlyRent:       2500,
			ExpectedStartDate: "2026-02-30",
		},
	}

	if err := req.Validate(); err == nil {
		t.Fatal("expected invalid rental date to be rejected")
	}
}
