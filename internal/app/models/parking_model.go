package models

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Parking struct {
	ID         int            `json:"id"`
	AuthorID   int            `json:"author_id"`
	Author     *User          `gorm:"foreignKey:author_id;references:ID" json:"author,omitempty"`
	Slug       string         `json:"slug"`
	Name       string         `json:"name"`
	Address    string         `json:"address"`
	DefaultFee float64        `json:"default_fee"`
	Latitude   float64        `json:"latitude"`
	Longitude  float64        `json:"longitude"`
	Layout     datatypes.JSON `json:"layout" gorm:"type:jsonb"`
	Slots      []ParkingSlot  `json:"slots" gorm:"foreignKey:parking_id;references:ID"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  *time.Time     `json:"deleted_at"`
}

type ParkingSlot struct {
	ID         int            `json:"id"`
	ParkingID  int            `json:"parking_id"`
	Parking    *Parking       `gorm:"foreignKey:parking_id;references:ID" json:"parking,omitempty"`
	Name       string         `json:"name"`
	Status     string         `json:"status"`
	Fee        float64        `json:"fee"`
	Row        int            `json:"row"`
	Col        int            `json:"col"`
	ESPHmac    string         `json:"esp_hmac"`
	PreviewURL string         `json:"preview_url"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `json:"deleted_at"`
}

type CreateParkingRequest struct {
	Slug       string         `json:"slug" validate:"required,min=3,max=255"`
	Name       string         `json:"name" validate:"required,min=3,max=255"`
	Address    string         `json:"address" validate:"required,min=3,max=255"`
	DefaultFee float64        `json:"default_fee" validate:"required,min=0"`
	Latitude   float64        `json:"latitude" validate:"required"`
	Longitude  float64        `json:"longitude" validate:"required"`
	Layout     datatypes.JSON `json:"layout" validate:"required"`
}

type UpdateParkingRequest struct {
	Slug       string         `json:"slug" validate:"omitempty,min=3,max=255"`
	Name       string         `json:"name" validate:"omitempty,min=3,max=255"`
	Address    string         `json:"address" validate:"omitempty,min=3,max=255"`
	DefaultFee float64        `json:"default_fee" validate:"omitempty,min=0"`
	Latitude   float64        `json:"latitude" validate:"omitempty"`
	Longitude  float64        `json:"longitude" validate:"omitempty"`
	Layout     datatypes.JSON `json:"layout" validate:"omitempty"`
}

type CreateParkingSlotRequest struct {
	ParkingID int     `json:"parking_id" validate:"required"`
	Name      string  `json:"name" validate:"required,min=1,max=8"`
	Status    string  `json:"status" validate:"required,oneof=AVAILABLE BOOKED OCCUPIED"`
	Fee       float64 `json:"fee" validate:"required,min=0"`
	Row       int     `json:"row" validate:"required"`
	Col       int     `json:"col" validate:"required"`
	ESPHmac   string  `json:"esp_hmac" validate:"omitempty"`
}

type UpdateParkingSlotRequest struct {
	Name       string  `json:"name" validate:"omitempty,min=1,max=8"`
	Status     string  `json:"status" validate:"omitempty,oneof=AVAILABLE BOOKED OCCUPIED"`
	Fee        float64 `json:"fee" validate:"omitempty,min=0"`
	Row        int     `json:"row" validate:"omitempty"`
	Col        int     `json:"col" validate:"omitempty"`
	ESPHmac    string  `json:"esp_hmac" validate:"omitempty"`
	PreviewURL string  `json:"preview_url" validate:"omitempty,url"`
}
