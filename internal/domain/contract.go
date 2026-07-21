package domain

import (
	"errors"
	"strings"
)

// ContractParty contains only the qualification needed to render a draft.
// Authorization and account identifiers stay in the backend.
type ContractParty struct {
	Name  string `json:"name"`
	CPF   string `json:"cpf"`
	Email string `json:"email"`
	Phone string `json:"phone"`
}

type ContractRequest struct {
	ContractID      string           `json:"contract_id"`
	DealType        string           `json:"deal_type"`
	PropertyTitle   string           `json:"property_title"`
	PropertyAddress string           `json:"property_address"`
	Seller          ContractParty    `json:"seller"`
	Buyer           ContractParty    `json:"buyer"`
	SaleTerms       PaymentBreakdown `json:"sale_terms"`
	RentalTerms     RentalTerms      `json:"rental_terms"`
}

func (p *ContractParty) Sanitize() {
	p.Name = sanitizeText(p.Name)
	p.CPF = sanitizeText(p.CPF)
	p.Email = sanitizeText(p.Email)
	p.Phone = sanitizeText(p.Phone)
}

func (r *ContractRequest) Sanitize() {
	r.ContractID = sanitizeText(r.ContractID)
	r.DealType = strings.ToLower(sanitizeText(r.DealType))
	r.PropertyTitle = sanitizeText(r.PropertyTitle)
	r.PropertyAddress = sanitizeText(r.PropertyAddress)
	r.Seller.Sanitize()
	r.Buyer.Sanitize()
	r.RentalTerms.Sanitize()
}

func (r *ContractRequest) Validate() error {
	r.Sanitize()
	if r.DealType != "sale" && r.DealType != "rent" {
		return errors.New("deal_type must be sale or rent")
	}
	if r.PropertyTitle == "" || r.PropertyAddress == "" {
		return errors.New("property_title and property_address are required")
	}
	if r.Seller.Name == "" || r.Buyer.Name == "" {
		return errors.New("seller.name and buyer.name are required")
	}
	if r.DealType == "rent" && r.RentalTerms.MonthlyRent <= 0 {
		return errors.New("rental_terms.monthly_rent must be greater than zero")
	}
	return nil
}
