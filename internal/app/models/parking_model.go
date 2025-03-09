package models

import "time"

type Parking struct {
	ID        int                      `json:"id"`
	AuthorID  int                      `json:"author_id"`
	Author    User                     `gorm:"foreignKey:author_id;references:ID"`
	Name      string                   `json:"name"`
	Address   string                   `json:"address"`
	Latitude  float64                  `json:"latitude"`
	Longitude float64                  `json:"longitude"`
	Layout    []map[string]interface{} `json:"layout"`
	Slots     []ParkingSlot            `json:"slots" gorm:"foreignKey:parking_id;references:ID"`
	CreatedAt time.Time                `json:"created_at"`
	UpdatedAt time.Time                `json:"updated_at"`
	DeletedAt *time.Time               `json:"deleted_at"`
}

type ParkingSlot struct {
	ID        int        `json:"id"`
	ParkingID int        `json:"parking_id"`
	Parking   Parking    `gorm:"foreignKey:parking_id;references:ID"`
	Name      string     `json:"name"`
	Status    string     `json:"status"`
	Fee       float64    `json:"fee"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at"`
}

type CreateParkingRequest struct {
	Name      string                   `json:"name" validate:"required,min=3,max=255"`
	Address   string                   `json:"address" validate:"required,min=3,max=255"`
	Latitude  float64                  `json:"latitude" validate:"required"`
	Longitude float64                  `json:"longitude" validate:"required"`
	Layout    []map[string]interface{} `json:"layout" validate:"required"`
}

type UpdateParkingRequest struct {
	Name      string                   `json:"name" validate:"omitempty,min=3,max=255"`
	Address   string                   `json:"address" validate:"omitempty,min=3,max=255"`
	Latitude  float64                  `json:"latitude" validate:"omitempty"`
	Longitude float64                  `json:"longitude" validate:"omitempty"`
	Layout    []map[string]interface{} `json:"layout" validate:"omitempty"`
}

type CreateParkingSlotRequest struct {
	ParkingID int     `json:"parking_id" validate:"required"`
	Name      string  `json:"name" validate:"required,min=1,max=8"`
	Status    string  `json:"status" validate:"required,oneof=AVAILABLE BOOKED OCCUPIED"`
	Fee       float64 `json:"fee" validate:"required,min=0"`
}

type UpdateParkingSlotRequest struct {
	Name   string  `json:"name" validate:"omitempty,min=1,max=8"`
	Status string  `json:"status" validate:"omitempty,oneof=AVAILABLE BOOKED OCCUPIED"`
	Fee    float64 `json:"fee" validate:"omitempty,min=0"`
}
