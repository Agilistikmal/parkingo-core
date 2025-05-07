package main

import (
	"github.com/agilistikmal/parkingo-core/injector"
	"github.com/agilistikmal/parkingo-core/internal/app/queues"
	"github.com/agilistikmal/parkingo-core/internal/app/services"
	"github.com/agilistikmal/parkingo-core/internal/infrastructure/config"
)

func main() {
	config.Load()

	routes := injector.InjectRoutes()
	routes.RegisterRoutes()

	s3Service := services.NewS3Service()
	mqttScanner := queues.NewScannerMQTT(s3Service)

	go mqttScanner.Subscribe()
	go routes.BookingJob.RunCheckBookingStatus()

	routes.FiberApp.Listen(":3000")
}
