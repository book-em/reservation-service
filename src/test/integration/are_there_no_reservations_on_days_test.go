package test

import (
	"bookem-reservation-service/internal"
	"bookem-reservation-service/util"
	"testing"
	"time"
)

func TestCheckRoomAvailability(t *testing.T) {
	hostUsername := "host_008"
	_, _, _, room := SetupHostRoomAvailabilityPrice(hostUsername, t)

	// Register and login guest
	RegisterUser("guest_008", "pass", util.Guest)
	guestJwt := LoginUser2("guest_008", "pass")

	// Create a reservation request for a specific date range
	dto := internal.CreateReservationRequestDTO{
		RoomID:     room.ID,
		DateFrom:   time.Date(2025, 9, 2, 0, 0, 0, 0, time.UTC),
		DateTo:     time.Date(2025, 9, 5, 0, 0, 0, 0, time.UTC),
		GuestCount: 2,
	}
	CreateReservationRequest(guestJwt, dto)

	// TODO: We can't test this properly until we implement creating reservations

	// // Check availability for a non-overlapping range (should be available)
	// resp1, err := CheckReservationAvailability(room.ID, "2025-09-4", "2025-09-6")
	// require.NoError(t, err)
	// require.Equal(t, http.StatusOK, resp1.StatusCode)

	// availability1 := ResponseToReservationAvailability(resp1)
	// assert.True(t, availability1.Available)

	// // Check availability for an overlapping range (should NOT be available)
	// resp2, err := CheckReservationAvailability(room.ID, "2025-09-2", "2025-09-5")
	// require.NoError(t, err)
	// require.Equal(t, http.StatusOK, resp2.StatusCode)

	// availability2 := ResponseToReservationAvailability(resp2)
	// assert.False(t, availability2.Available)
}
