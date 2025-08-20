package internal

type ReservationDTO struct {
	ID      uint `json:"id"`
	RoomID  uint `json:"roomId"`
	GuestID uint `json:"guestId"`
}
