package services

import (
	"context"
	"fmt"

	"github.com/agilistikmal/parkingo-core/internal/app/models"
	"github.com/agilistikmal/parkingo-core/internal/app/pkg"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"github.com/xendit/xendit-go/v6"
	"github.com/xendit/xendit-go/v6/invoice"
	"gorm.io/gorm"
)

type BookingService struct {
	DB           *gorm.DB
	Validate     *validator.Validate
	XenditClient *xendit.APIClient
}

func NewBookingService(db *gorm.DB, validate *validator.Validate, xenditClient *xendit.APIClient) *BookingService {
	return &BookingService{
		DB:           db,
		Validate:     validate,
		XenditClient: xenditClient,
	}
}

func (s *BookingService) GetBookings() ([]*models.Booking, error) {
	var bookings []*models.Booking
	err := s.DB.Find(&bookings).Error
	if err != nil {
		return nil, err
	}

	return bookings, nil
}

func (s *BookingService) GetBookingByID(id int) (*models.Booking, error) {
	var booking *models.Booking
	err := s.DB.First(&booking, id).Error
	if err != nil {
		return nil, err
	}

	return booking, nil
}

func (s *BookingService) CreateBooking(userID int, req *models.CreateBookingRequest) (*models.Booking, error) {
	err := s.Validate.Struct(req)
	if err != nil {
		return nil, err
	}

	var user *models.User
	err = s.DB.First(&user, userID).Error
	if err != nil {
		return nil, err
	}

	paymentReference := "PKGO-" + pkg.RandomString(8)
	invoiceRequest := *invoice.NewCreateInvoiceRequest(paymentReference, req.TotalFee)
	invoiceRequest.SetPayerEmail(user.Email)
	invoiceRequest.SetDescription(fmt.Sprintf("Parking fee for %s", req.PlateNumber))
	invoiceRequest.SetCurrency("IDR")

	paymentInvoice, _, err := s.XenditClient.InvoiceApi.CreateInvoice(context.Background()).
		CreateInvoiceRequest(invoiceRequest).
		Execute()
	if err != nil {
		logrus.Error(err)
		return nil, err
	}

	booking := models.Booking{
		UserID:           userID,
		ParkingID:        req.ParkingID,
		SlotID:           req.SlotID,
		PlateNumber:      req.PlateNumber,
		StartAt:          req.StartAt,
		EndAt:            req.EndAt,
		TotalHours:       req.TotalHours,
		TotalFee:         req.TotalFee,
		PaymentReference: paymentInvoice.ExternalId,
		PaymentLink:      paymentInvoice.InvoiceUrl,
		Status:           req.Status,
	}

	err = s.DB.Create(&booking).Error
	if err != nil {
		return nil, err
	}

	return &booking, nil
}

func (s *BookingService) UpdateBooking(id int, req *models.UpdateBookingRequest) (*models.Booking, error) {
	err := s.Validate.Struct(req)
	if err != nil {
		return nil, err
	}

	booking, err := s.GetBookingByID(id)
	if err != nil {
		return nil, err
	}

	if req.PlateNumber != "" {
		booking.PlateNumber = req.PlateNumber
	}

	if !req.StartAt.IsZero() {
		booking.StartAt = req.StartAt
	}

	if !req.EndAt.IsZero() {
		booking.EndAt = req.EndAt
	}

	if req.TotalHours != 0 {
		booking.TotalHours = req.TotalHours
	}

	if req.TotalFee != 0 {
		booking.TotalFee = req.TotalFee
	}

	if req.Status != "" {
		booking.Status = req.Status
	}

	err = s.DB.Save(&booking).Error
	if err != nil {
		return nil, err
	}

	return booking, nil
}

func (s *BookingService) DeleteBooking(id int) error {
	booking, err := s.GetBookingByID(id)
	if err != nil {
		return err
	}

	err = s.DB.Delete(&booking).Error
	if err != nil {
		return err
	}

	return nil
}
