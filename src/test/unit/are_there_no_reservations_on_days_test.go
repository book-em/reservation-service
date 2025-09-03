package test

import (
	"bookem-reservation-service/internal"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_AreThereReservationsOnDays_AllDaysFree(t *testing.T) {
	svc, repo, _, _ := CreateTestRoomService()

	// Mock: No reservations on any day
	repo.On("FindReservationsByRoomIDForDay", uint(1), mock.Anything).Return([]internal.Reservation{}, nil)

	from := time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2025, 9, 3, 0, 0, 0, 0, time.UTC)

	result, err := svc.AreThereReservationsOnDays(1, from, to)

	assert.NoError(t, err)
	assert.False(t, result)
}

func Test_AreThereReservationsOnDays_OneDayBooked(t *testing.T) {
	svc, repo, _, _ := CreateTestRoomService()

	// Mock: First two days free, third day has a reservation
	repo.On("FindReservationsByRoomIDForDay", uint(1), time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC)).Return([]internal.Reservation{}, nil)
	repo.On("FindReservationsByRoomIDForDay", uint(1), time.Date(2025, 9, 2, 0, 0, 0, 0, time.UTC)).Return([]internal.Reservation{}, nil)
	repo.On("FindReservationsByRoomIDForDay", uint(1), time.Date(2025, 9, 3, 0, 0, 0, 0, time.UTC)).Return([]internal.Reservation{
		{RoomID: 1, DateFrom: time.Date(2025, 9, 3, 0, 0, 0, 0, time.UTC), DateTo: time.Date(2025, 9, 5, 0, 0, 0, 0, time.UTC)},
	}, nil)

	from := time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2025, 9, 3, 0, 0, 0, 0, time.UTC)

	result, err := svc.AreThereReservationsOnDays(1, from, to)

	assert.NoError(t, err)
	assert.True(t, result)
}

func Test_AreThereReservationsOnDays_RepoError(t *testing.T) {
	svc, repo, _, _ := CreateTestRoomService()

	// Mock: Error on first day
	repo.On("FindReservationsByRoomIDForDay", uint(1), time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC)).Return([]internal.Reservation{}, errors.New("db error"))

	from := time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2025, 9, 2, 0, 0, 0, 0, time.UTC)

	result, err := svc.AreThereReservationsOnDays(1, from, to)

	assert.Error(t, err)
	assert.False(t, result)
}
