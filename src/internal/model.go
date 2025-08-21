package internal

import "time"

type ReservationRequestStatus string

const (
	Pending  ReservationRequestStatus = "pending"
	Accepted ReservationRequestStatus = "accepted"
	Rejected ReservationRequestStatus = "rejected"
)

type ReservationRequest struct {
	ID                 uint                     `gorm:"primaryKey"`
	RoomID             uint                     `gorm:"not null"`
	RoomAvailabilityID uint                     `gorm:"not null"`
	RoomPriceID        uint                     `gorm:"not null"`
	DateFrom           time.Time                `gorm:"not null"` // Including year
	DateTo             time.Time                `gorm:"not null"` // Including year
	GuestCount         uint                     `gorm:"not null"`
	GuestID            uint                     `gorm:"not null"` // User who made the request
	Status             ReservationRequestStatus `gorm:"not null"`
	Cost               uint                     `gorm:"not null"` // Computed field
}

type Reservation struct {
	ID                 uint      `gorm:"primaryKey"`
	RoomID             uint      `gorm:"not null"`
	RoomAvailabilityID uint      `gorm:"not null"`
	RoomPriceID        uint      `gorm:"not null"`
	GuestID            uint      `gorm:"not null"` // User who made the request
	DateFrom           time.Time `gorm:"not null"` // Including year
	DateTo             time.Time `gorm:"not null"` // Including year
	GuestCount         uint      `gorm:"not null"`
	Cancelled          bool      `gorm:"not null"`
	Cost               uint      `gorm:"not null"` // Computed field
}
