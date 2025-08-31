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

func TestDeleteReservationRequest(t *testing.T) {
	_, _, _, room := SetupHostRoomAvailabilityPrice("host_007", t)

	RegisterUser("guest_007", "pass", userclient.Guest)
	jwt := LoginUser2("guest_007", "pass")

	dto := internal.CreateReservationRequestDTO{
		RoomID:     room.ID,
		DateFrom:   time.Date(2025, 9, 6, 0, 0, 0, 0, time.UTC),
		DateTo:     time.Date(2025, 9, 8, 0, 0, 0, 0, time.UTC),
		GuestCount: 1,
	}
	resp, err := CreateReservationRequest(jwt, dto)
	req := ResponseToReservationRequest(resp)
	require.NoError(t, err)

	// Delete the request
	delResp, err := DeleteReservationRequest(jwt, req.ID)
	require.NoError(t, err)
	require.Equal(t, http.StatusNoContent, delResp.StatusCode)

	// Verify it's gone
	listResp, err := GetPendingRequestsByGuest(jwt)
	require.NoError(t, err)
	requests := ResponseToReservationRequests(listResp)
	assert.Len(t, requests, 0)
}
