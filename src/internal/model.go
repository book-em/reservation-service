package internal

type Reservation struct {
	ID      uint `gorm:"primaryKey"`
	RoomID  uint `gorm:"not null"`
	GuestID uint `gorm:"not null"`
}
