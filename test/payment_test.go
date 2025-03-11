package test

import (
	"context"
	"testing"

	"github.com/agilistikmal/parkingo-core/internal/app/pkg"
	"github.com/agilistikmal/parkingo-core/internal/infrastructure/config"
	"github.com/agilistikmal/parkingo-core/internal/infrastructure/paymentgateway"
	"github.com/sirupsen/logrus"
	"github.com/xendit/xendit-go/v6/invoice"
)

func TestPayment(t *testing.T) {
	config.Load()
	xdt := paymentgateway.NewXendit()

	paymentReference := "PKGO-" + pkg.RandomString(8)
	invoiceRequest := *invoice.NewCreateInvoiceRequest(paymentReference, 1000)
	invoiceRequest.SetPayerEmail("agilistikmal3@gmail.com")
	invoiceRequest.SetDescription("Testing Parking Fee KB 1234 AGL")
	invoiceRequest.SetCurrency("IDR")

	paymentInvoice, _, err := xdt.InvoiceApi.CreateInvoice(context.Background()).
		CreateInvoiceRequest(invoiceRequest).
		Execute()
	if err != nil {
		logrus.Error(err)
	}

	logrus.Info(paymentInvoice)
}
