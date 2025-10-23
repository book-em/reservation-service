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

func TestApproveReservationRequest(t *testing.T) {
	hostUsername := "host_009"
	_, _, hostJwt, room := SetupHostRoomAvailabilityPrice(hostUsername, t)

	RegisterUser("guest_009", "pass", util.Guest)
	guestJwt := LoginUser2("guest_009", "pass")

	dto := internal.CreateReservationRequestDTO{
		RoomID:     room.ID,
		DateFrom:   time.Date(2025, 9, 6, 0, 0, 0, 0, time.UTC),
		DateTo:     time.Date(2025, 9, 8, 0, 0, 0, 0, time.UTC),
		GuestCount: 2,
	}

	resp, err := CreateReservationRequest(guestJwt, dto)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	req := ResponseToReservationRequest(resp)

	approveURL := URL_reservation + "req/" + strconv.FormatUint(uint64(req.ID), 10) + "/approve"
	request, err := http.NewRequest(http.MethodPut, approveURL, nil)
	require.NoError(t, err)
	request.Header.Add("Authorization", "Bearer "+hostJwt)

	approveResp, err := http.DefaultClient.Do(request)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, approveResp.StatusCode)

	requestsResp, err := GetPendingRequestsByRoom(hostJwt, room.ID)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, requestsResp.StatusCode)

	requests := ResponseToReservationRequests(requestsResp)
	if len(requests) > 0 {
		assert.Equal(t, "accepted", requests[0].Status)
	}
}
