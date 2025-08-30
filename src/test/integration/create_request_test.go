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

func TestCreateReservationRequest_HappyPath(t *testing.T) {
	_, _, _, room := SetupHostRoomAvailabilityPrice("host_001", t)

	RegisterUser("guest_001", "pass", userclient.Guest)
	jwt := LoginUser2("guest_001", "pass")

	dto := internal.CreateReservationRequestDTO{
		RoomID:     room.ID,
		DateFrom:   time.Date(2025, 9, 2, 0, 0, 0, 0, time.UTC),
		DateTo:     time.Date(2025, 9, 4, 0, 0, 0, 0, time.UTC),
		GuestCount: 2,
	}

	resp, err := CreateReservationRequest(jwt, dto)
	req := ResponseToReservationRequest(resp)
	require.NoError(t, err)
	assert.Equal(t, dto.RoomID, req.RoomID)
	assert.Equal(t, dto.GuestCount, req.GuestCount)
	assert.Equal(t, string(internal.Pending), req.Status)
}

func TestCreateReservationRequest_DuplicateRequestForSameRoom(t *testing.T) {
	_, _, _, room := SetupHostRoomAvailabilityPrice("host_002", t)

	RegisterUser("guest_002", "pass", userclient.Guest)
	jwt := LoginUser2("guest_002", "pass")

	dto := internal.CreateReservationRequestDTO{
		RoomID:     room.ID,
		DateFrom:   time.Date(2025, 9, 15, 0, 0, 0, 0, time.UTC),
		DateTo:     time.Date(2025, 9, 17, 0, 0, 0, 0, time.UTC),
		GuestCount: 1,
	}

	_, err := CreateReservationRequest(jwt, dto)
	require.NoError(t, err)

	resp, err := CreateReservationRequest(jwt, dto)
	require.NoError(t, err)
	require.Equal(t, http.StatusConflict, resp.StatusCode)
}
