package models

import (
	"time"
)

type Booking struct {
	ID               int         `json:"id"`
	UserID           int         `json:"user_id"`
	User             User        `gorm:"foreignKey:UserID" json:"user"`
	ParkingID        int         `json:"parking_id"`
	Parking          Parking     `gorm:"foreignKey:ParkingID" json:"parking"`
	SlotID           int         `json:"slot_id"`
	Slot             ParkingSlot `gorm:"foreignKey:SlotID" json:"slot"`
	PlateNumber      string      `json:"plate_number"`
	StartAt          time.Time   `json:"start_at"`
	EndAt            time.Time   `json:"end_at"`
	TotalHours       int         `json:"total_hours"`
	TotalFee         float64     `json:"total_fee"`
	PaymentReference string      `json:"payment_reference"`
	PaymentLink      string      `json:"payment_link"`
	Status           string      `json:"status"`
	CreatedAt        time.Time   `json:"created_at"`
	UpdatedAt        time.Time   `json:"updated_at"`
	DeletedAt        *time.Time  `json:"deleted_at"`
}

type CreateBookingRequest struct {
	ParkingID   int       `json:"parking_id" validate:"required"`
	SlotID      int       `json:"slot_id" validate:"required"`
	PlateNumber string    `json:"plate_number" validate:"required,min=3,max=16"`
	StartAt     time.Time `json:"start_at" validate:"required"`
	EndAt       time.Time `json:"end_at" validate:"required"`
	TotalHours  int       `json:"total_hours" validate:"required"`
	TotalFee    float64   `json:"total_fee" validate:"required"`
	Status      string    `json:"status" validate:"required,oneof=UNPAID PAID CANCELLED EXPIRED COMPLETED"`
}

type UpdateBookingRequest struct {
	PlateNumber string    `json:"plate_number" validate:"omitempty,min=3,max=16"`
	StartAt     time.Time `json:"start_at" validate:"omitempty"`
	EndAt       time.Time `json:"end_at" validate:"omitempty"`
	TotalHours  int       `json:"total_hours" validate:"omitempty"`
	TotalFee    float64   `json:"total_fee" validate:"omitempty"`
	Status      string    `json:"status" validate:"omitempty,oneof=UNPAID PAID CANCELLED EXPIRED COMPLETED"`
}
