package jobs

import (
	"time"

	"github.com/agilistikmal/parkingo-core/internal/app/models"
	"github.com/agilistikmal/parkingo-core/internal/app/services"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
)

type BookingJob struct {
	BookingService *services.BookingService
	ParkingService *services.ParkingService
	TimeLocation   *time.Location
}

func NewBookingJob(bookingService *services.BookingService, parkingService *services.ParkingService) *BookingJob {
	loc, _ := time.LoadLocation("Asia/Jakarta")
	return &BookingJob{
		BookingService: bookingService,
		ParkingService: parkingService,
		TimeLocation:   loc,
	}
}

func (j *BookingJob) checkBookingStatus() {
	logrus.Info("Checking booking status")
	bookings, err := j.BookingService.GetBookings(&models.BookingFilter{
		Status: "PAID",
	})
	if err != nil {
		logrus.Error("Failed to get bookings: ", err)
		return
	}

	logrus.Info("Found ", len(bookings), " bookings")

	for _, booking := range bookings {
		if booking.Status == "PAID" && booking.EndAt.Before(time.Now()) && booking.Status != "COMPLETED" {
			logrus.Infof("Updating booking %s status to COMPLETED", booking.PaymentReference)
			_, err = j.BookingService.UpdateBooking(booking.ID, &models.UpdateBookingRequest{
				Status: "COMPLETED",
			})
			if err != nil {
				logrus.Error("Failed to update booking: ", err)
			}

			_, err = j.ParkingService.UpdateParkingSlot(booking.SlotID, &models.UpdateParkingSlotRequest{
				Status: "AVAILABLE",
			})
			if err != nil {
				logrus.Error("Failed to update parking slot: ", err)
			}
		}
	}
}

func (j *BookingJob) RunCheckBookingStatus() {
	logrus.Info("Running check booking status")
	c := cron.New(cron.WithLocation(j.TimeLocation))
	_, err := c.AddFunc("*/5 * * * *", j.checkBookingStatus)
	if err != nil {
		logrus.Error("Failed to add check booking status to cron: ", err)
		return
	}
	c.Start()
}
