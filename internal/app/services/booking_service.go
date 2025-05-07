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
	MailService  *MailService
}

func NewBookingService(db *gorm.DB, validate *validator.Validate, xenditClient *xendit.APIClient, mailService *MailService) *BookingService {
	return &BookingService{
		DB:           db,
		Validate:     validate,
		XenditClient: xenditClient,
		MailService:  mailService,
	}
}

func (s *BookingService) GetBookings(filter *models.BookingFilter) ([]*models.Booking, error) {
	var bookings []*models.Booking
	query := s.DB.Preload("Parking").Preload("Slot").Preload("User")

	if filter != nil {
		if filter.UserID != 0 {
			query = query.Where("user_id = ?", filter.UserID)
		}
		if filter.ParkingID != 0 {
			query = query.Or("parking_id = ?", filter.ParkingID)
		}
		if filter.Status != "" {
			query = query.Or("status = ?", filter.Status)
		}
		if filter.SortBy != "" {
			query = query.Order(fmt.Sprintf("%s %s", filter.SortBy, filter.SortOrder))
		}
		if filter.Limit != 0 {
			query = query.Limit(filter.Limit)
		}
		if filter.Page != 0 {
			query = query.Offset((filter.Page - 1) * filter.Limit)
		}
	}

	err := query.Find(&bookings).Error
	if err != nil {
		return nil, err
	}

	return bookings, nil
}

func (s *BookingService) GetBookingByID(id int) (*models.Booking, error) {
	var booking *models.Booking
	err := s.DB.Preload("User").Preload("Parking").Preload("Slot").First(&booking, id).Error
	if err != nil {
		return nil, err
	}

	return booking, nil
}

func (s *BookingService) GetBookingByReference(reference string) (*models.Booking, error) {
	var booking *models.Booking
	err := s.DB.Preload("User").Preload("Parking").Preload("Slot").Where("payment_reference = ?", reference).First(&booking).Error
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

	now := pkg.GetCurrentTime()
	if req.StartAt.Before(now) {
		return nil, fmt.Errorf("start time must be in the future")
	}

	var user *models.User
	err = s.DB.First(&user, userID).Error
	if err != nil {
		return nil, err
	}

	// Check if the parking slot is available
	var similarBookings []models.Booking
	err = s.DB.Where("slot_id = ? AND (start_at BETWEEN ? AND ? OR end_at BETWEEN ? AND ?) AND status NOT IN ('CANCELED', 'EXPIRED', 'COMPLETED')", req.SlotID, req.StartAt, req.EndAt, req.StartAt, req.EndAt).Find(&similarBookings).Error
	if err != nil {
		return nil, err
	}

	if len(similarBookings) > 0 {
		return nil, fmt.Errorf("slot is already booked")
	}

	// Check if the parking slot is available
	var parkingSlot models.ParkingSlot
	err = s.DB.Preload("Parking").First(&parkingSlot, req.SlotID).Error
	if err != nil {
		return nil, err
	}

	totalHours := int(req.EndAt.Sub(req.StartAt).Hours())
	if totalHours < 3 {
		return nil, fmt.Errorf("minimum booking time is 3 hours")
	}

	totalFee := float64(totalHours) * parkingSlot.Fee

	customer := *invoice.NewCustomerObject()
	customer.SetEmail(user.Email)
	customer.SetGivenNames(user.FullName)
	customer.SetId(string(user.ID))

	items := []invoice.InvoiceItem{
		{
			Name:        fmt.Sprintf("%s | %s | %s", parkingSlot.Parking.Name, parkingSlot.Name, req.PlateNumber),
			Price:       float32(parkingSlot.Fee),
			Quantity:    float32(totalHours),
			ReferenceId: &parkingSlot.Parking.Slug,
		},
	}

	paymentReference := "PKGO-" + pkg.RandomString(8)
	invoiceRequest := *invoice.NewCreateInvoiceRequest(paymentReference, totalFee)
	invoiceRequest.SetPayerEmail(user.Email)
	invoiceRequest.SetDescription(fmt.Sprintf("Parking fee for %s", req.PlateNumber))
	invoiceRequest.SetCurrency("IDR")
	invoiceRequest.SetSuccessRedirectUrl(fmt.Sprintf("https://parkingo.agil.zip/b/%s", paymentReference))
	invoiceRequest.SetInvoiceDuration("600")
	invoiceRequest.SetCustomer(customer)
	invoiceRequest.SetItems(items)

	paymentInvoice, _, err := s.XenditClient.InvoiceApi.CreateInvoice(context.Background()).
		CreateInvoiceRequest(invoiceRequest).
		Execute()
	if paymentInvoice.Id == nil {
		logrus.Error("Payment Error:", err)
		return nil, err
	}

	booking := models.Booking{
		UserID:           userID,
		ParkingID:        req.ParkingID,
		SlotID:           req.SlotID,
		PlateNumber:      req.PlateNumber,
		StartAt:          req.StartAt,
		EndAt:            req.EndAt,
		PaymentReference: paymentInvoice.ExternalId,
		PaymentLink:      paymentInvoice.InvoiceUrl,
		PaymentExpiredAt: paymentInvoice.ExpiryDate,
		Status:           "UNPAID",
		TotalHours:       totalHours,
		TotalFee:         totalFee,
	}

	err = s.DB.Transaction(func(tx *gorm.DB) error {
		err = tx.Create(&booking).Error
		if err != nil {
			return err
		}

		parkingSlot.Status = "BOOKED"
		err = tx.Save(&parkingSlot).Error
		if err != nil {
			return err
		}

		// Send email to user
		go s.MailService.SendMail(user.Email, fmt.Sprintf("Booking Confirmation %s", booking.PaymentReference), fmt.Sprintf("Your booking is confirmed. Booking invoice and detail: https://parkingo.agil.zip/b/%s", booking.PaymentReference))

		return nil
	})
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

	parkingSlot := booking.Slot

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

		if req.Status == "CANCELED" {
			parkingSlot.Status = "AVAILABLE"
		}
		if req.Status == "EXPIRED" {
			parkingSlot.Status = "AVAILABLE"
		}
		if req.Status == "COMPLETED" {
			parkingSlot.Status = "AVAILABLE"
		}
		if req.Status == "UNPAID" {
			parkingSlot.Status = "BOOKED"
		}
		if req.Status == "PAID" {
			parkingSlot.Status = "OCCUPIED"
		}
	}

	err = s.DB.Transaction(func(tx *gorm.DB) error {
		err = tx.Save(&booking).Error
		if err != nil {
			return err
		}

		err = tx.Save(&parkingSlot).Error
		if err != nil {
			return err
		}

		return nil
	})
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
