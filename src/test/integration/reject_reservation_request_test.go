package test

import (
	"bookem-reservation-service/internal"
	"bookem-reservation-service/util"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRejectReservationRequest(t *testing.T) {
	hostUsername := "host_010"
	_, _, hostJwt, room := SetupHostRoomAvailabilityPrice(hostUsername, t)

	RegisterUser("guest_010", "pass", util.Guest)
	guestJwt := LoginUser2("guest_010", "pass")

	dto := internal.CreateReservationRequestDTO{
		RoomID:     room.ID,
		DateFrom:   time.Date(2025, 9, 22, 0, 0, 0, 0, time.UTC),
		DateTo:     time.Date(2025, 9, 24, 0, 0, 0, 0, time.UTC),
		GuestCount: 2,
	}

	resp, err := CreateReservationRequest(guestJwt, dto)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	req := ResponseToReservationRequest(resp)

	rejectURL := URL_reservation + "req/" + strconv.FormatUint(uint64(req.ID), 10) + "/reject"
	request, err := http.NewRequest(http.MethodPut, rejectURL, nil)
	require.NoError(t, err)
	request.Header.Add("Authorization", "Bearer "+hostJwt)

	rejectResp, err := http.DefaultClient.Do(request)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, rejectResp.StatusCode)

	requestsResp, err := GetPendingRequestsByRoom(hostJwt, room.ID)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, requestsResp.StatusCode)

	requests := ResponseToReservationRequests(requestsResp)
	if len(requests) > 0 {
		assert.Equal(t, "rejected", requests[0].Status)
	}
}
