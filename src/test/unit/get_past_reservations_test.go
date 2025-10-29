package test

import (
	"bookem-reservation-service/internal"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_GetPastReservationsByGuest_Success(t *testing.T) {
	svc, repo, _, _, _ := CreateTestRoomService()

	before := time.Now().UTC()
	guestID := uint(10)

	reservations := []internal.Reservation{
		{
			ID:         1,
			RoomID:     5,
			DateFrom:   before.AddDate(0, -1, 0),
			DateTo:     before.AddDate(0, -1, 1),
			GuestCount: 2,
			GuestID:    guestID,
			Cancelled:  false,
			Cost:       250,
		},
		{
			ID:         2,
			RoomID:     6,
			DateFrom:   before.AddDate(0, -2, 0),
			DateTo:     before.AddDate(0, -2, 1),
			GuestCount: 3,
			GuestID:    guestID,
			Cancelled:  false,
			Cost:       400,
		},
	}

	repo.On("GetAllPastReservationsByGuest", guestID, before).
		Return(reservations, nil)

	out, err := svc.GetPastReservationsByGuest(context.Background(), guestID, before)

	assert.NoError(t, err)
	assert.Len(t, out, 2)

	assert.Equal(t, reservations[0].ID, out[0].ID)
	assert.Equal(t, reservations[1].RoomID, out[1].RoomID)
	assert.Equal(t, reservations[1].Cost, out[1].Cost)

	repo.AssertCalled(t, "GetAllPastReservationsByGuest", guestID, before)
}

func Test_GetPastReservationsByGuest_RepoError(t *testing.T) {
	svc, repo, _, _, _ := CreateTestRoomService()

	before := time.Now().UTC()
	guestID := uint(5)

	repo.On("GetAllPastReservationsByGuest", guestID, before).
		Return([]internal.Reservation{}, errors.New("db failure"))

	out, err := svc.GetPastReservationsByGuest(context.Background(), guestID, before)

	assert.Error(t, err)
	assert.Nil(t, out)
	assert.Contains(t, err.Error(), "db failure")
	repo.AssertCalled(t, "GetAllPastReservationsByGuest", guestID, before)
}

func Test_GetPastReservationsByGuest_EmptyList(t *testing.T) {
	svc, repo, _, _, _ := CreateTestRoomService()

	before := time.Now().UTC()
	guestID := uint(3)

	repo.On("GetAllPastReservationsByGuest", guestID, before).
		Return([]internal.Reservation{}, nil)

	out, err := svc.GetPastReservationsByGuest(context.Background(), guestID, before)

	assert.NoError(t, err)
	assert.Len(t, out, 0)
	repo.AssertCalled(t, "GetAllPastReservationsByGuest", guestID, before)
}
