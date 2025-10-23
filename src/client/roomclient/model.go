package roomclient

import "time"

type RoomDTO struct {
	ID          uint     `json:"id"`
	HostID      uint     `json:"hostID"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Address     string   `json:"address"`
	MinGuests   uint     `json:"minGuests"`
	MaxGuests   uint     `json:"maxGuests"`
	Photos      []string `json:"photos"`
	Commodities []string `json:"commodities"`
	AutoApprove bool     `json:"autoApprove"`
}

type CreateRoomDTO struct {
	HostID        uint     `json:"hostID"`
	Name          string   `json:"name"`
	Description   string   `json:"description"`
	Address       string   `json:"address"`
	MinGuests     uint     `json:"minGuests"`
	MaxGuests     uint     `json:"maxGuests"`
	PhotosPayload []string `json:"photosPayload"`
	Commodities   []string `json:"commodities"`
	AutoApprove   bool     `json:"autoApprove"`
}

// ---------------------------------------------------------------

type CreateRoomAvailabilityListDTO struct {
	RoomID uint                            `json:"roomId"`
	Items  []CreateRoomAvailabilityItemDTO `json:"items"`
}

type RoomAvailabilityListDTO struct {
	ID            uint                      `json:"id"`
	RoomID        uint                      `json:"roomId"`
	EffectiveFrom time.Time                 `json:"effectiveFrom"`
	Items         []RoomAvailabilityItemDTO `json:"items"`
}

type RoomAvailabilityItemDTO struct {
	ID        uint      `json:"id"`
	DateFrom  time.Time `json:"dateFrom"`
	DateTo    time.Time `json:"dateTo"`
	Available bool      `json:"available"`
}

type CreateRoomAvailabilityItemDTO struct {
	// ExistingID is either the ID of an RoomAvailabilityItem that already
	// exists, or 0 if this is a new item. When 0, a new one will be created in
	// the DB. When not 0, it will reuse the existing object.
	ExistingID uint      `json:"existingId"`
	DateFrom   time.Time `json:"dateFrom"`
	DateTo     time.Time `json:"dateTo"`
	Available  bool      `json:"available"`
}

// ---------------------------------------------------------------

type CreateRoomPriceListDTO struct {
	RoomID    uint                     `json:"roomId"`
	Items     []CreateRoomPriceItemDTO `json:"items"`
	BasePrice uint                     `json:"basePrice"`
	PerGuest  bool                     `json:"perGuest"`
}

type RoomPriceListDTO struct {
	ID            uint               `json:"id"`
	RoomID        uint               `json:"roomId"`
	EffectiveFrom time.Time          `json:"effectiveFrom"`
	BasePrice     uint               `json:"basePrice"`
	Items         []RoomPriceItemDTO `json:"items"`
	PerGuest      bool               `json:"perGuest"`
}

type RoomPriceItemDTO struct {
	ID       uint      `json:"id"`
	DateFrom time.Time `json:"dateFrom"`
	DateTo   time.Time `json:"dateTo"`
	Price    uint      `json:"price"`
}

type CreateRoomPriceItemDTO struct {
	// ExistingID is either the ID of an RoomPriceItem that already
	// exists, or 0 if this is a new item. When 0, a new one will be created in
	// the DB. When not 0, it will reuse the existing object.
	ExistingID uint      `json:"existingId"`
	DateFrom   time.Time `json:"dateFrom"`
	DateTo     time.Time `json:"dateTo"`
	Price      uint      `json:"price"`
}

// ---------------------------------------------------------------

type RoomReservationQueryDTO struct {
	RoomID     uint      `json:"roomId"`
	DateFrom   time.Time `json:"dateFrom"`
	DateTo     time.Time `json:"dateTo"`
	GuestCount uint      `json:"guestCount"`
}

type RoomReservationQueryResponseDTO struct {
	Available bool `json:"available"`
	TotalCost uint `json:"totalCost"`
}
