package test

import (
	"bookem-reservation-service/internal"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCancelReservation_Success(t *testing.T) {
	svc, mockRepo, mockUser, _, _ := CreateTestRoomService()

	res := &internal.Reservation{
		ID:        1,
		GuestID:   1,
		DateFrom:  time.Now().Add(48 * time.Hour),
		Cancelled: false,
	}

	mockUser.On("FindById", context.Background(), uint(1)).Return(DefaultUser_Guest, nil)
	mockRepo.On("FindReservationById", uint(1)).Return(res, nil)
	mockRepo.On("CancelReservation", uint(1)).Return(nil)

	err := svc.CancelReservation(context.Background(), 1, 1)

	assert.NoError(t, err)
	mockRepo.AssertCalled(t, "CancelReservation", uint(1))
}

func TestCancelReservation_AlreadyCancelled(t *testing.T) {
	svc, mockRepo, mockUser, _, _ := CreateTestRoomService()

	res := &internal.Reservation{
		ID:        2,
		GuestID:   1,
		DateFrom:  time.Now().Add(48 * time.Hour),
		Cancelled: true,
	}

	mockUser.On("FindById", context.Background(), uint(1)).Return(DefaultUser_Guest, nil)
	mockRepo.On("FindReservationById", uint(2)).Return(res, nil)

	err := svc.CancelReservation(context.Background(), 1, 2)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already cancelled")
}

func TestCancelReservation_AlreadyStarted(t *testing.T) {
	svc, mockRepo, mockUser, _, _ := CreateTestRoomService()

	res := &internal.Reservation{
		ID:        3,
		GuestID:   1,
		DateFrom:  time.Now().Add(-2 * time.Hour),
		Cancelled: false,
	}

	mockUser.On("FindById", context.Background(), uint(1)).Return(DefaultUser_Guest, nil)
	mockRepo.On("FindReservationById", uint(3)).Return(res, nil)

	err := svc.CancelReservation(context.Background(), 1, 3)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot cancel reservation that already started")
}

func TestCancelReservation_WrongGuest(t *testing.T) {
	svc, mockRepo, mockUser, _, _ := CreateTestRoomService()

	res := &internal.Reservation{
		ID:        4,
		GuestID:   99,
		DateFrom:  time.Now().Add(48 * time.Hour),
		Cancelled: false,
	}

	mockUser.On("FindById", context.Background(), uint(1)).Return(DefaultUser_Guest, nil)
	mockRepo.On("FindReservationById", uint(4)).Return(res, nil)

	err := svc.CancelReservation(context.Background(), 1, 4)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Unauthorized")
}

func TestCancelReservation_UserNotFound(t *testing.T) {
	svc, _, mockUser, _, _ := CreateTestRoomService()

	mockUser.On("FindById", context.Background(), uint(1)).Return(nil, errors.New("user not found"))

	err := svc.CancelReservation(context.Background(), 1, 5)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Unauthenticated")
}
