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
	address, city, state := resolveIntroLocation(req)
	validityDays := req.ResolvedValidityDays()
	totalValue := req.ResolvedTotalValue()
	payment := req.ResolvedPayments()
	brokerName := fallback(req.ResolvedBrokerName(), "Proprietário/Corretor")
	sellingBrokerName := strings.ToUpper(
		fallback(req.ResolvedSellingBrokerName(), "PROPRIETÁRIO DO IMÓVEL"),
	)

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(20, 20, 20)
	pdf.SetAutoPageBreak(true, 20)
	pdf.AddPage()
	tr := pdf.UnicodeTranslatorFromDescriptor("")

	// Header
	pdf.SetFont("Arial", "B", 18)
	pdf.CellFormat(0, 10, tr("PROPOSTA DE COMPRA DE IMÓVEL"), "", 1, "C", false, 0, "")
	pdf.Ln(8)

	// Addressee
	pdf.SetFont("Arial", "B", 13)
	pdf.CellFormat(0, 7, tr("Ilmo(a) Sr(a).:"), "", 1, "L", false, 0, "")
	pdf.SetFont("Arial", "BI", 13)
	pdf.CellFormat(0, 7, tr(sellingBrokerName), "", 1, "L", false, 0, "")
	pdf.Ln(12)

	// Intro paragraph
	intro := fmt.Sprintf(
		"Esta proposta tem por finalidade assegurar uma oferta de compra de um imóvel de sua propriedade, situado à %s, na cidade de %s - %s, por parte do comprador, nas seguintes condições:",
		address,
		city,
		state,
	)
	pdf.SetFont("Arial", "", 12)
	pdf.MultiCell(0, 7, tr(intro), "", "J", false)
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
		pdf.MultiCell(0, 7, tr(line), "", "L", false)
	}

	pdf.Ln(2)
	pdf.SetFont("Arial", "", 12)
	pdf.MultiCell(0, 7, tr(fmt.Sprintf("Esta proposta é válida por %d dias.", validityDays)), "", "L", false)

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
	pdf.CellFormat(lineWidth, 6, tr(fmt.Sprintf("%s (Proponente)", clientName)), "", 0, "C", false, 0, "")
	pdf.SetXY(rightX, signatureY+2)
	pdf.CellFormat(lineWidth, 6, tr(fmt.Sprintf("%s (Proprietário/Corretor)", brokerName)), "", 0, "C", false, 0, "")

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

func resolveIntroLocation(req domain.ProposalRequest) (string, string, string) {
	fullAddress := fallback(req.ResolvedPropertyAddress(), "______________________")
	city := strings.TrimSpace(req.ResolvedCity())
	state := strings.TrimSpace(req.ResolvedState())

	parts := strings.Split(fullAddress, ",")
	trimmedParts := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			trimmedParts = append(trimmedParts, trimmed)
		}
	}

	usedIndexes := map[int]struct{}{}
	stateIndex := -1
	if state == "" {
		for i := len(trimmedParts) - 1; i >= 0; i-- {
			candidate := normalizeStateToken(trimmedParts[i])
			if candidate != "" {
				state = candidate
				stateIndex = i
				usedIndexes[i] = struct{}{}
				break
			}
		}
	}

	if city == "" {
		cityIndex := -1
		if stateIndex > 0 {
			cityIndex = stateIndex - 1
		} else if len(trimmedParts) >= 2 {
			cityIndex = len(trimmedParts) - 2
		}
		if cityIndex >= 0 && cityIndex < len(trimmedParts) {
			city = trimmedParts[cityIndex]
			usedIndexes[cityIndex] = struct{}{}
		}
	}

	addressParts := make([]string, 0, len(trimmedParts))
	for i, part := range trimmedParts {
		if _, used := usedIndexes[i]; used {
			continue
		}
		addressParts = append(addressParts, part)
	}
	address := strings.Join(addressParts, ", ")
	if strings.TrimSpace(address) == "" {
		address = fullAddress
	}

	return fallback(address, "______________________"), fallback(city, "____________"), fallback(strings.ToUpper(state), "__")
}

func normalizeStateToken(value string) string {
	trimmed := strings.TrimSpace(strings.ToUpper(value))
	trimmed = strings.TrimPrefix(trimmed, "UF ")
	trimmed = strings.TrimPrefix(trimmed, "UF:")
	trimmed = strings.Trim(trimmed, "- ")
	if len(trimmed) == 2 {
		return trimmed
	}
	return ""
}
