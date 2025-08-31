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

func TestFindPendingRequestByRoom(t *testing.T) {
	hostUsername := "host_006"
	_, _, hostJwt, room := SetupHostRoomAvailabilityPrice(hostUsername, t)

	RegisterUser("guest_006", "pass", userclient.Guest)
	guestJwt := LoginUser2("guest_006", "pass")

	dto := internal.CreateReservationRequestDTO{
		RoomID:     room.ID,
		DateFrom:   time.Date(2025, 9, 3, 0, 0, 0, 0, time.UTC),
		DateTo:     time.Date(2025, 9, 5, 0, 0, 0, 0, time.UTC),
		GuestCount: 1,
	}
	CreateReservationRequest(guestJwt, dto)

	resp, err := GetPendingRequestsByRoom(hostJwt, room.ID)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	requests := ResponseToReservationRequests(resp)
	require.Len(t, requests, 1)
	assert.Equal(t, room.ID, requests[0].RoomID)
	assert.Equal(t, "pending", requests[0].Status)
}
