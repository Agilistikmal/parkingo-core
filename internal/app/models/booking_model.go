package models

import (
	"time"
)

type Booking struct {
	ID               uint        `json:"id"`
	UserID           uint        `json:"user_id"`
	User             User        `gorm:"foreignKey:UserID" json:"user"`
	ParkingID        uint        `json:"parking_id"`
	Parking          Parking     `gorm:"foreignKey:ParkingID" json:"parking"`
	SlotID           uint        `json:"slot_id"`
	Slot             ParkingSlot `gorm:"foreignKey:SlotID" json:"slot"`
	PlateNumber      string      `json:"plate_number"`
	StartAt          time.Time   `json:"start_at"`
	EndAt            time.Time   `json:"end_at"`
	TotalHours       int         `json:"total_hours"`
	TotalFee         float64     `json:"total_fee"`
	PaymentReference string      `json:"payment_reference"`
	Status           string      `json:"status"`
	CreatedAt        time.Time   `json:"created_at"`
	UpdatedAt        time.Time   `json:"updated_at"`
	DeletedAt        *time.Time  `json:"deleted_at"`
}
