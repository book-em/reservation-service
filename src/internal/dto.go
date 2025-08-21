package internal

import "time"

type CreateReservationRequestDTO struct {
	RoomID     uint      `json:"roomId"`
	DateFrom   time.Time `json:"dateFrom"`
	DateTo     time.Time `json:"dateTo"`
	GuestCount uint      `json:"guestCount"`
}

type ReservationRequestDTO struct {
	ID         uint      `json:"id"`
	RoomID     uint      `json:"roomId"`
	DateFrom   time.Time `json:"dateFrom"`
	DateTo     time.Time `json:"dateTo"`
	GuestCount uint      `json:"guestCount"`
	GuestID    uint      `json:"guestId"`
	Status     string    `json:"status"`
	Cost       uint      `json:"cost"`
}

type ReservationDTO struct {
	ID         uint      `json:"id"`
	RoomID     uint      `json:"roomId"`
	DateFrom   time.Time `json:"dateFrom"`
	DateTo     time.Time `json:"dateTo"`
	GuestCount uint      `json:"guestCount"`
	GuestID    uint      `json:"guestId"`
	Cancelled  bool      `json:"cancelled"`
	Cost       uint      `json:"cost"`
}

func ToReservationRequestDTO(r ReservationRequest) ReservationRequestDTO {
	return ReservationRequestDTO{
		ID:         r.ID,
		RoomID:     r.RoomID,
		DateFrom:   r.DateFrom,
		DateTo:     r.DateTo,
		GuestCount: r.GuestCount,
		GuestID:    r.GuestID,
		Status:     string(r.Status),
	}
}

func ToReservationDTO(r Reservation) ReservationDTO {
	return ReservationDTO{
		ID:         r.ID,
		RoomID:     r.RoomID,
		DateFrom:   r.DateFrom,
		DateTo:     r.DateTo,
		GuestCount: r.GuestCount,
		GuestID:    r.GuestID,
		Cancelled:  r.Cancelled,
	}
}
