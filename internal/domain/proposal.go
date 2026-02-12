package domain

import (
	"errors"
	"strings"
)

type ProposalRequest struct {
	ClientName        string  `json:"client_name"`
	ClientCPF         string  `json:"client_cpf"`
	PropertyAddress   string  `json:"property_address"`
	BrokerName        string  `json:"broker_name"`
	SellingBrokerName string  `json:"selling_broker_name"`
	Value             float64 `json:"value"`
	PaymentMethod     string  `json:"payment_method"`
	ValidityDays      int     `json:"validity_days"`
}

func (p *ProposalRequest) Validate() error {
	if strings.TrimSpace(p.ClientName) == "" {
		return errors.New("client_name is required")
	}
	if strings.TrimSpace(p.PaymentMethod) == "" {
		return errors.New("payment_method is required")
	}
	if p.Value <= 0 {
		return errors.New("value must be greater than zero")
	}
	return nil
}
