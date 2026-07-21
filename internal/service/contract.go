package service

import (
	"bytes"
	"fmt"

	"github.com/jung-kurt/gofpdf"

	"pdf-service/internal/domain"
)

// GenerateContract produces a non-signed draft. The backend records its
// template provenance and controls who may retrieve it.
func (s *PDFService) GenerateContract(req domain.ContractRequest) ([]byte, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(20, 20, 20)
	pdf.SetAutoPageBreak(true, 20)
	pdf.SetCompression(false)
	pdf.AddPage()
	tr := pdf.UnicodeTranslatorFromDescriptor("")

	title := "MINUTA DE CONTRATO DE COMPRA E VENDA"
	sellerRole, buyerRole := "VENDEDOR", "COMPRADOR"
	if req.DealType == "rent" {
		title = "MINUTA DE CONTRATO DE LOCAÇÃO"
		sellerRole, buyerRole = "LOCADOR", "LOCATÁRIO"
	}

	pdf.SetFont("Arial", "B", 16)
	pdf.CellFormat(0, 10, tr(title), "", 1, "C", false, 0, "")
	pdf.Ln(5)

	pdf.SetFont("Arial", "", 11)
	pdf.MultiCell(0, 6, tr(fmt.Sprintf(
		"MINUTA NÃO ASSINADA. Imóvel: %s. Endereço: %s.",
		req.PropertyTitle,
		req.PropertyAddress,
	)), "", "J", false)
	pdf.Ln(3)

	writeContractParty(pdf, tr, sellerRole, req.Seller)
	writeContractParty(pdf, tr, buyerRole, req.Buyer)

	pdf.SetFont("Arial", "B", 12)
	pdf.CellFormat(0, 7, tr("CLÁUSULAS COMERCIAIS"), "", 1, "L", false, 0, "")
	pdf.SetFont("Arial", "", 11)
	for _, line := range buildContractCommercialTerms(req) {
		pdf.MultiCell(0, 6, tr(line), "", "J", false)
	}
	pdf.Ln(3)
	pdf.MultiCell(0, 6, tr("As partes reconhecem que esta minuta deverá ser revisada pela imobiliária e formalizada presencialmente, em papel, antes de produzir efeitos definitivos."), "", "J", false)
	pdf.Ln(18)
	pdf.Line(25, pdf.GetY(), 90, pdf.GetY())
	pdf.Line(120, pdf.GetY(), 185, pdf.GetY())
	pdf.SetFont("Arial", "", 9)
	pdf.CellFormat(65, 5, tr(sellerRole), "", 0, "C", false, 0, "")
	pdf.CellFormat(30, 5, "", "", 0, "C", false, 0, "")
	pdf.CellFormat(65, 5, tr(buyerRole), "", 1, "C", false, 0, "")

	var out bytes.Buffer
	if err := pdf.Output(&out); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

func writeContractParty(pdf *gofpdf.Fpdf, tr func(string) string, role string, party domain.ContractParty) {
	pdf.SetFont("Arial", "B", 12)
	pdf.CellFormat(0, 7, tr(role), "", 1, "L", false, 0, "")
	pdf.SetFont("Arial", "", 11)
	pdf.MultiCell(0, 6, tr(fmt.Sprintf("Nome: %s\nCPF: %s\nE-mail: %s\nTelefone: %s", fallback(party.Name, "______________________"), fallback(party.CPF, "______________________"), fallback(party.Email, "______________________"), fallback(party.Phone, "______________________"))), "", "L", false)
	pdf.Ln(3)
}

func buildContractCommercialTerms(req domain.ContractRequest) []string {
	if req.DealType == "rent" {
		terms := req.RentalTerms
		lines := []string{fmt.Sprintf("1. Valor mensal da locação: %s.", formatBRL(terms.MonthlyRent))}
		if terms.GuaranteeType != "" {
			line := fmt.Sprintf("2. Garantia locatícia: %s", terms.GuaranteeType)
			if terms.GuaranteeAmount > 0 {
				line += fmt.Sprintf(" no valor de %s", formatBRL(terms.GuaranteeAmount))
			}
			lines = append(lines, line+".")
		}
		if terms.LeaseTermMonths > 0 {
			lines = append(lines, fmt.Sprintf("3. Prazo de locação: %d meses.", terms.LeaseTermMonths))
		}
		if terms.ExpectedStartDate != "" {
			lines = append(lines, fmt.Sprintf("4. Início previsto: %s.", formatISODateForDisplay(terms.ExpectedStartDate)))
		}
		if terms.MonthlyDueDay > 0 {
			lines = append(lines, fmt.Sprintf("5. Vencimento mensal: dia %d.", terms.MonthlyDueDay))
		}
		if terms.CondominiumResponsibility != "" {
			lines = append(lines, fmt.Sprintf("Responsabilidade pelo condomínio: %s.", terms.CondominiumResponsibility))
		}
		if terms.PropertyTaxResponsibility != "" {
			lines = append(lines, fmt.Sprintf("Responsabilidade pelo IPTU: %s.", terms.PropertyTaxResponsibility))
		}
		return lines
	}

	proposal := domain.ProposalRequest{Payment: req.SaleTerms}
	payment := proposal.ResolvedPayments()
	total := payment.Cash + payment.TradeIn + payment.Financing + payment.Others
	return buildSaleProposalTerms(total, payment)
}
