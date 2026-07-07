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
