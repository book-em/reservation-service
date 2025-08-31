package test

import (
	"bookem-reservation-service/client/userclient"
	"bookem-reservation-service/internal"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFindPendingRequestsByGuest(t *testing.T) {
	_, _, _, room := SetupHostRoomAvailabilityPrice("host_005", t)

	RegisterUser("guest_005", "pass", userclient.Guest)
	jwt := LoginUser2("guest_005", "pass")

	dto := internal.CreateReservationRequestDTO{
		RoomID:     room.ID,
		DateFrom:   time.Date(2025, 9, 2, 0, 0, 0, 0, time.UTC),
		DateTo:     time.Date(2025, 9, 4, 0, 0, 0, 0, time.UTC),
		GuestCount: 2,
	}
	CreateReservationRequest(jwt, dto)

	resp, err := GetPendingRequestsByGuest(jwt)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	requests := ResponseToReservationRequests(resp)
	require.Len(t, requests, 1)
	assert.Equal(t, room.ID, requests[0].RoomID)
	assert.Equal(t, "pending", requests[0].Status)
}
