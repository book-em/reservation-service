package test

import (
	"bookem-reservation-service/internal"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_GetActiveGuestReservations_UserNotFound(t *testing.T) {
	svc, mockRepo, mockUserClient, _ := CreateTestRoomService()

	guest := DefaultUser_Guest

	mockUserClient.On("FindById", guest.Id).Return(nil, errors.New("User is not found"))

	reservations, err := svc.GetActiveGuestReservations(guest.Id)

	assert.Error(t, err)
	assert.Nil(t, reservations)
	mockUserClient.AssertNumberOfCalls(t, "FindById", 1)
	mockUserClient.AssertExpectations(t)
	mockRepo.AssertNumberOfCalls(t, "FindReservationsByGuestID", 0)
	mockRepo.AssertExpectations(t)
}

func Test_GetActiveGuestReservations_FindReservationsError(t *testing.T) {
	svc, mockRepo, mockUserClient, _ := CreateTestRoomService()

	guest := DefaultUser_Guest
	var reservations []internal.Reservation

	mockUserClient.On("FindById", guest.Id).Return(guest, nil)
	mockRepo.On("FindReservationsByGuestID", guest.Id).Return(reservations, errors.New("Reservations are not found"))

	reservations, err := svc.GetActiveGuestReservations(guest.Id)

	assert.Error(t, err)
	assert.Nil(t, reservations)
	mockUserClient.AssertNumberOfCalls(t, "FindById", 1)
	mockUserClient.AssertExpectations(t)
	mockRepo.AssertNumberOfCalls(t, "FindReservationsByGuestID", 1)
	mockRepo.AssertExpectations(t)
}

// This test automatically encapsulates enough of `ExtractActiveReservations`,
// as its logic is exceptionally simple.
func Test_GetActiveGuestReservations_SuccessNoneActive(t *testing.T) {
	svc, mockRepo, mockUserClient, _ := CreateTestRoomService()

	guest := DefaultUser_Guest
	reservartion := *DefaultReservation
	reservartion.Cancelled = true
	reservartion.DateTo = time.Now().Add(24 * time.Hour)
	reservations := []internal.Reservation{reservartion}

	mockUserClient.On("FindById", guest.Id).Return(guest, nil)
	mockRepo.On("FindReservationsByGuestID", guest.Id).Return(reservations, nil)

	reservartionsGot, err := svc.GetActiveGuestReservations(guest.Id)

	assert.NoError(t, err)
	assert.Nil(t, reservartionsGot)
	mockUserClient.AssertNumberOfCalls(t, "FindById", 1)
	mockUserClient.AssertExpectations(t)
	mockRepo.AssertNumberOfCalls(t, "FindReservationsByGuestID", 1)
	mockRepo.AssertExpectations(t)
}

// This test automatically encapsulates enough of `ExtractActiveReservations`,
// as its logic is exceptionally simple.
func Test_GetActiveGuestReservations_Success(t *testing.T) {
	svc, mockRepo, mockUserClient, _ := CreateTestRoomService()

	guest := DefaultUser_Guest
	reservartion1 := *DefaultReservation
	reservartion1.Cancelled = false
	reservartion1.DateTo = time.Now().Add(24 * time.Hour)
	reservartion2 := *DefaultReservation
	reservartion2.Cancelled = true
	reservartion2.DateTo = time.Now().Add(24 * time.Hour)
	reservations := []internal.Reservation{reservartion1, reservartion2}
	expectedReservations := []internal.Reservation{reservartion1}

	mockUserClient.On("FindById", guest.Id).Return(guest, nil)
	mockRepo.On("FindReservationsByGuestID", guest.Id).Return(reservations, nil)

	reservartionsGot, err := svc.GetActiveGuestReservations(guest.Id)

	assert.NoError(t, err)
	assert.Equal(t, expectedReservations, reservartionsGot)
	mockUserClient.AssertNumberOfCalls(t, "FindById", 1)
	mockUserClient.AssertExpectations(t)
	mockRepo.AssertNumberOfCalls(t, "FindReservationsByGuestID", 1)
	mockRepo.AssertExpectations(t)
}
