package service

import (
	"bytes"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/jung-kurt/gofpdf"

	"pdf-service/internal/domain"
)

type PDFService struct{}

func NewPDFService() *PDFService {
	return &PDFService{}
}

func (s *PDFService) GenerateProposal(req domain.ProposalRequest) ([]byte, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	clientName := req.ResolvedClientName()
	address := fallback(req.ResolvedPropertyAddress(), "______________________")
	city := fallback(req.ResolvedCity(), "____________")
	state := fallback(req.ResolvedState(), "__")
	validityDays := req.ResolvedValidityDays()
	totalValue := req.ResolvedTotalValue()
	payment := req.ResolvedPayments()
	brokerName := fallback(req.ResolvedBrokerName(), "Proprietário/Corretor")

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(20, 20, 20)
	pdf.SetAutoPageBreak(true, 20)
	pdf.AddPage()

	// Header
	pdf.SetFont("Arial", "B", 18)
	pdf.CellFormat(0, 10, "PROPOSTA DE COMPRA DE IMÓVEL", "", 1, "C", false, 0, "")
	pdf.Ln(8)

	// Addressee
	pdf.SetFont("Arial", "B", 13)
	pdf.CellFormat(0, 7, "Ilmo(a) Sr(a).:", "", 1, "L", false, 0, "")
	pdf.SetFont("Arial", "BI", 13)
	pdf.CellFormat(0, 7, "(PROPRIETÁRIO DO IMÓVEL)", "", 1, "L", false, 0, "")
	pdf.Ln(12)

	// Intro paragraph
	intro := fmt.Sprintf(
		"Esta proposta tem por finalidade assegurar uma oferta de compra de um imóvel de sua propriedade, situado à %s, na cidade de %s - %s, por parte do comprador, nas seguintes condições:",
		address,
		city,
		state,
	)
	pdf.SetFont("Arial", "", 12)
	pdf.MultiCell(0, 7, intro, "", "J", false)
	pdf.Ln(2)

	// Financial breakdown
	lines := []string{
		fmt.Sprintf("• Valor total da proposta: %s", formatBRL(totalValue)),
		fmt.Sprintf("• Valor em dinheiro (Sinal/Entrada): %s", formatBRL(payment.Cash)),
	}
	if payment.TradeIn > 0 {
		lines = append(lines, fmt.Sprintf("• Permuta: %s", formatBRL(payment.TradeIn)))
	}
	if payment.Financing > 0 {
		lines = append(lines, fmt.Sprintf("• Financiamento: %s", formatBRL(payment.Financing)))
	}
	if payment.Others > 0 {
		lines = append(lines, fmt.Sprintf("• Outros: %s", formatBRL(payment.Others)))
	}

	for _, line := range lines {
		pdf.MultiCell(0, 7, line, "", "L", false)
	}

	pdf.Ln(2)
	pdf.SetFont("Arial", "", 12)
	pdf.MultiCell(0, 7, fmt.Sprintf("Obs.: Esta proposta é válida por %d dias.", validityDays), "", "L", false)

	// Signatures
	currentY := pdf.GetY()
	if currentY > 240 {
		pdf.AddPage()
		currentY = 40
	}

	signatureY := currentY + 24
	leftMargin, _, rightMargin, _ := pdf.GetMargins()
	pageWidth, _ := pdf.GetPageSize()
	contentWidth := pageWidth - leftMargin - rightMargin
	gap := 18.0
	lineWidth := (contentWidth - gap) / 2
	leftX := leftMargin
	rightX := leftX + lineWidth + gap

	pdf.Line(leftX, signatureY, leftX+lineWidth, signatureY)
	pdf.Line(rightX, signatureY, rightX+lineWidth, signatureY)

	pdf.SetFont("Arial", "", 11)
	pdf.SetXY(leftX, signatureY+2)
	pdf.CellFormat(lineWidth, 6, fmt.Sprintf("%s (Proponente)", clientName), "", 0, "C", false, 0, "")
	pdf.SetXY(rightX, signatureY+2)
	pdf.CellFormat(lineWidth, 6, fmt.Sprintf("%s (Proprietário/Corretor)", brokerName), "", 0, "C", false, 0, "")

	var out bytes.Buffer
	if err := pdf.Output(&out); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

func formatBRL(value float64) string {
	sign := ""
	if value < 0 {
		sign = "-"
		value = math.Abs(value)
	}

	intPart := int64(value)
	decPart := int(math.Round((value - float64(intPart)) * 100))
	if decPart == 100 {
		intPart++
		decPart = 0
	}

	intText := strconv.FormatInt(intPart, 10)
	var grouped strings.Builder
	for i, r := range intText {
		if i > 0 && (len(intText)-i)%3 == 0 {
			grouped.WriteByte('.')
		}
		grouped.WriteRune(r)
	}

	return fmt.Sprintf("R$ %s%s,%02d", sign, grouped.String(), decPart)
}

func fallback(value, fallbackValue string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return fallbackValue
	}
	return trimmed
}
