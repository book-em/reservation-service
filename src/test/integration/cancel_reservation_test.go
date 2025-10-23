package test

import (
	"bookem-reservation-service/internal"
	"bookem-reservation-service/util"

	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCancelReservation_Success(t *testing.T) {
	hostUsername := "host_cancel_001"
	_, _, hostJwt, room := SetupHostRoomAvailabilityPrice(hostUsername, t)

	RegisterUser("guest_cancel_001", "pass", util.Guest)
	guestJwt := LoginUser2("guest_cancel_001", "pass")

	dto := internal.CreateReservationRequestDTO{
		RoomID:     room.ID,
		DateFrom:   time.Date(2025, 12, 6, 0, 0, 0, 0, time.UTC),
		DateTo:     time.Date(2025, 12, 8, 0, 0, 0, 0, time.UTC),
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

	guestActiveResp, err := GetActiveGuestReservations(guestJwt)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, guestActiveResp.StatusCode)

	reservations := ResponseToReservations(guestActiveResp)
	require.Greater(t, len(reservations), 0)

	reservationID := reservations[0].ID
	cancelURL := URL_reservation + "reservations/" + strconv.FormatUint(uint64(reservationID), 10) + "/cancel"

	cancelReq, err := http.NewRequest(http.MethodDelete, cancelURL, nil)
	require.NoError(t, err)
	cancelReq.Header.Add("Authorization", "Bearer "+guestJwt)

	cancelResp, err := http.DefaultClient.Do(cancelReq)
	require.NoError(t, err)
	require.Equal(t, http.StatusNoContent, cancelResp.StatusCode)
}

func TestCancelReservation_UnauthorizedGuest(t *testing.T) {
	hostUsername := "host_cancel_003"
	_, _, hostJwt, room := SetupHostRoomAvailabilityPrice(hostUsername, t)

	RegisterUser("guest_owner_001", "pass", util.Guest)
	guest1Jwt := LoginUser2("guest_owner_001", "pass")

	RegisterUser("guest_intruder_001", "pass", util.Guest)
	guest2Jwt := LoginUser2("guest_intruder_001", "pass")

	dto := internal.CreateReservationRequestDTO{
		RoomID:     room.ID,
		DateFrom:   time.Date(2025, 12, 6, 0, 0, 0, 0, time.UTC),
		DateTo:     time.Date(2025, 12, 8, 0, 0, 0, 0, time.UTC),
		GuestCount: room.MinGuests,
	}

	resp, err := CreateReservationRequest(guest1Jwt, dto)
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

	guest1ActiveResp, err := GetActiveGuestReservations(guest1Jwt)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, guest1ActiveResp.StatusCode)

	reservations := ResponseToReservations(guest1ActiveResp)
	require.Greater(t, len(reservations), 0)

	reservationID := reservations[0].ID

	cancelURL := URL_reservation + "reservations/" + strconv.FormatUint(uint64(reservationID), 10) + "/cancel"
	cancelReq, err := http.NewRequest(http.MethodDelete, cancelURL, nil)
	require.NoError(t, err)
	cancelReq.Header.Add("Authorization", "Bearer "+guest2Jwt)

	cancelResp, err := http.DefaultClient.Do(cancelReq)
	require.NoError(t, err)

	require.Equal(t, http.StatusUnauthorized, cancelResp.StatusCode)
}
