package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
)

type FlexibleAddress struct {
	Raw          string
	Street       string
	Number       string
	Neighborhood string
	City         string
	State        string
	Complement   string
}

func (a *FlexibleAddress) UnmarshalJSON(data []byte) error {
	raw := strings.TrimSpace(string(data))
	if raw == "" || raw == "null" {
		return nil
	}

	if strings.HasPrefix(raw, "\"") {
		unquoted, err := strconv.Unquote(raw)
		if err != nil {
			return err
		}
		a.Raw = strings.TrimSpace(unquoted)
		return nil
	}

	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		return err
	}

	read := func(keys ...string) string {
		for _, key := range keys {
			if value, ok := payload[key]; ok && value != nil {
				parsed := strings.TrimSpace(fmt.Sprint(value))
				if parsed != "" {
					return parsed
				}
			}
		}
		return ""
	}

	a.Street = read("street", "logradouro", "address", "line1")
	a.Number = read("number", "numero")
	a.Neighborhood = read("neighborhood", "bairro")
	a.City = read("city", "localidade")
	a.State = strings.ToUpper(read("state", "uf"))
	a.Complement = read("complement", "complemento")
	a.Raw = read("formatted", "full", "raw")
	return nil
}

type PaymentBreakdown struct {
	Cash          float64 `json:"cash"`
	TradeIn       float64 `json:"tradeIn"`
	Financing     float64 `json:"financing"`
	Others        float64 `json:"others"`
	Dinheiro      float64 `json:"dinheiro"`
	Permuta       float64 `json:"permuta"`
	Financiamento float64 `json:"financiamento"`
	Outros        float64 `json:"outros"`
}

type PaymentValues struct {
	Cash      float64
	TradeIn   float64
	Financing float64
	Others    float64
}

type ProposalRequest struct {
	ClientNameLegacy      string  `json:"client_name"`
	ClientCPFLegacy       string  `json:"client_cpf"`
	PropertyAddressLegacy string  `json:"property_address"`
	BrokerNameLegacy      string  `json:"broker_name"`
	TotalValueLegacy      float64 `json:"value"`
	PaymentMethodLegacy   string  `json:"payment_method"`
	ValidityDaysLegacy    int     `json:"validity_days"`

	ClientName      string           `json:"clientName"`
	ClientCPF       string           `json:"clientCpf"`
	PropertyAddress FlexibleAddress  `json:"propertyAddress"`
	BrokerName      string           `json:"brokerName"`
	TotalValue      float64          `json:"totalValue"`
	Payment         PaymentBreakdown `json:"payment"`
	ValidityDays    int              `json:"validadeDias"`

	PropertyCity  string `json:"propertyCity"`
	PropertyState string `json:"propertyState"`
}

func (p *ProposalRequest) Validate() error {
	if strings.TrimSpace(p.ResolvedClientName()) == "" {
		return errors.New("client_name is required")
	}
	if strings.TrimSpace(p.ResolvedPropertyAddress()) == "" {
		return errors.New("property_address is required")
	}
	total := p.ResolvedTotalValue()
	if total <= 0 {
		return errors.New("value must be greater than zero")
	}
	validity := p.ResolvedValidityDays()
	if validity <= 0 {
		return errors.New("validity_days must be greater than zero")
	}

	payments := p.ResolvedPayments()
	sum := payments.Cash + payments.TradeIn + payments.Financing + payments.Others
	if sum <= 0 {
		return errors.New("payment breakdown is required")
	}

	if math.Abs(sum-total) > 0.01 {
		return errors.New("payment breakdown must match total value")
	}

	return nil
}

func (p *ProposalRequest) ResolvedClientName() string {
	return firstNonBlank(p.ClientName, p.ClientNameLegacy)
}

func (p *ProposalRequest) ResolvedClientCPF() string {
	return firstNonBlank(p.ClientCPF, p.ClientCPFLegacy)
}

func (p *ProposalRequest) ResolvedBrokerName() string {
	return firstNonBlank(p.BrokerName, p.BrokerNameLegacy)
}

func (p *ProposalRequest) ResolvedValidityDays() int {
	if p.ValidityDays > 0 {
		return p.ValidityDays
	}
	if p.ValidityDaysLegacy > 0 {
		return p.ValidityDaysLegacy
	}
	return 10
}

func (p *ProposalRequest) ResolvedTotalValue() float64 {
	if p.TotalValue > 0 {
		return p.TotalValue
	}
	if p.TotalValueLegacy > 0 {
		return p.TotalValueLegacy
	}
	payments := p.ResolvedPayments()
	return payments.Cash + payments.TradeIn + payments.Financing + payments.Others
}

func (p *ProposalRequest) ResolvedPropertyAddress() string {
	if strings.TrimSpace(p.PropertyAddress.Raw) != "" {
		return strings.TrimSpace(p.PropertyAddress.Raw)
	}
	if strings.TrimSpace(p.PropertyAddressLegacy) != "" {
		return strings.TrimSpace(p.PropertyAddressLegacy)
	}

	parts := []string{
		p.PropertyAddress.Street,
		withPrefixIfPresent("Nº ", p.PropertyAddress.Number),
		p.PropertyAddress.Neighborhood,
		p.PropertyAddress.City,
		p.PropertyAddress.State,
		p.PropertyAddress.Complement,
	}

	filtered := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			filtered = append(filtered, trimmed)
		}
	}
	return strings.Join(filtered, ", ")
}

func (p *ProposalRequest) ResolvedCity() string {
	return firstNonBlank(p.PropertyAddress.City, p.PropertyCity)
}

func (p *ProposalRequest) ResolvedState() string {
	return strings.ToUpper(firstNonBlank(p.PropertyAddress.State, p.PropertyState))
}

func (p *ProposalRequest) ResolvedPayments() PaymentValues {
	values := PaymentValues{
		Cash:      firstPositive(p.Payment.Cash, p.Payment.Dinheiro),
		TradeIn:   firstPositive(p.Payment.TradeIn, p.Payment.Permuta),
		Financing: firstPositive(p.Payment.Financing, p.Payment.Financiamento),
		Others:    firstPositive(p.Payment.Others, p.Payment.Outros),
	}

	if values.Cash+values.TradeIn+values.Financing+values.Others > 0 {
		return values
	}

	legacy := strings.TrimSpace(p.PaymentMethodLegacy)
	if legacy == "" {
		return values
	}

	values.Cash = extractLegacyPaymentValue(legacy, "Dinheiro")
	values.TradeIn = extractLegacyPaymentValue(legacy, "Permuta")
	values.Financing = extractLegacyPaymentValue(legacy, "Financiamento")
	values.Others = extractLegacyPaymentValue(legacy, "Outros")
	return values
}

func withPrefixIfPresent(prefix, value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}
	return prefix + trimmed
}

func firstNonBlank(values ...string) string {
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func firstPositive(values ...float64) float64 {
	for _, value := range values {
		if value > 0 {
			return value
		}
	}
	return 0
}

func extractLegacyPaymentValue(source string, label string) float64 {
	pattern := regexp.MustCompile(label + `\s*:\s*R\$\s*([0-9\.,]+)`)
	matches := pattern.FindStringSubmatch(source)
	if len(matches) < 2 {
		return 0
	}

	value := strings.ReplaceAll(matches[1], ".", "")
	value = strings.ReplaceAll(value, ",", ".")
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0
	}
	return parsed
}
