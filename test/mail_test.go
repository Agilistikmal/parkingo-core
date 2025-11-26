package test

import (
	"fmt"
	"testing"

	"github.com/agilistikmal/parkingo-core/internal/app/services"
	"github.com/agilistikmal/parkingo-core/internal/infrastructure/config"
)

func TestMailService_SendMailOvertime(t *testing.T) {
	config.Load()
	mailService := services.NewMailService()
	mailService.SendMail("agilistikmal3@gmail.com", fmt.Sprintf("Booking Overtime %s", "PKGO-1234567890"), fmt.Sprintf("Your booking is overtime. You might charged extra fee for this booking. Booking invoice and detail: https://parkingo.agil.zip/b/PKGO-1234567890"))
}
