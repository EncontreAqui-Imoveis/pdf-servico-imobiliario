package service

import (
	"fmt"
	"strings"

	"github.com/johnfercher/maroto/pkg/consts"
	"github.com/johnfercher/maroto/pkg/pdf"
	"github.com/johnfercher/maroto/pkg/props"

	"pdf-service/internal/domain"
)

type PDFService struct{}

func NewPDFService() *PDFService {
	return &PDFService{}
}

func (s *PDFService) GenerateProposal(req domain.ProposalRequest) ([]byte, error) {
	m := pdf.NewMaroto(consts.Portrait, consts.A4)
	m.SetPageMargins(20, 15, 20)

	m.Row(12, func() {
		m.Col(12, func() {
			m.Text("PROPOSTA DE AQUISIÇÃO", props.Text{
				Align: consts.Center,
				Style: consts.Bold,
				Size:  16,
			})
		})
	})

	m.Row(6, func() {
		m.Col(12, func() {
			m.Text("Dados do Cliente", props.Text{Style: consts.Bold, Size: 12})
		})
	})

	m.Row(6, func() {
		m.Col(12, func() {
			m.Text(fmt.Sprintf("Nome: %s", req.ClientName), props.Text{Size: 10})
		})
	})

	if strings.TrimSpace(req.ClientCPF) != "" {
		m.Row(6, func() {
			m.Col(12, func() {
				m.Text(fmt.Sprintf("CPF: %s", req.ClientCPF), props.Text{Size: 10})
			})
		})
	}

	m.Row(6, func() {
		m.Col(12, func() {
			m.Text("Dados do Imóvel", props.Text{Style: consts.Bold, Size: 12})
		})
	})

	m.Row(6, func() {
		m.Col(12, func() {
			m.Text(fmt.Sprintf("Endereço: %s", req.PropertyAddress), props.Text{Size: 10})
		})
	})

	if strings.TrimSpace(req.BrokerName) != "" {
		m.Row(6, func() {
			m.Col(12, func() {
				m.Text(fmt.Sprintf("Corretor: %s", req.BrokerName), props.Text{Size: 10})
			})
		})
	}

	if strings.TrimSpace(req.SellingBrokerName) != "" && req.SellingBrokerName != req.BrokerName {
		m.Row(6, func() {
			m.Col(12, func() {
				m.Text(fmt.Sprintf("Corretor Vendedor: %s", req.SellingBrokerName), props.Text{Size: 10})
			})
		})
	}

	m.Line(1)

	m.Row(8, func() {
		m.Col(4, func() {
			m.Text(fmt.Sprintf("Valor: R$ %.2f", req.Value), props.Text{Size: 10, Style: consts.Bold})
		})
		m.Col(4, func() {
			m.Text(fmt.Sprintf("Pagamento: %s", req.PaymentMethod), props.Text{Size: 10})
		})
		m.Col(4, func() {
			m.Text(fmt.Sprintf("Validade: %d dias", req.ValidityDays), props.Text{Size: 10})
		})
	})

	m.Row(12, func() {
		m.Col(6, func() {
			m.Text("_____________________________", props.Text{Align: consts.Center})
			m.Text("Comprador", props.Text{Align: consts.Center, Size: 10})
		})
		m.Col(6, func() {
			m.Text("_____________________________", props.Text{Align: consts.Center})
			m.Text("Corretor", props.Text{Align: consts.Center, Size: 10})
		})
	})

	output, err := m.Output()
	if err != nil {
		return nil, err
	}

	return output.Bytes(), nil
}
