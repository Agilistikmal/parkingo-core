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
