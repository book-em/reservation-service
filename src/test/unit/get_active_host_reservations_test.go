package test

import (
	"bookem-reservation-service/internal"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_GetActiveHostReservations_UserNotFound(t *testing.T) {
	svc, mockRepo, mockUserClient, _ := CreateTestRoomService()

	host := DefaultUser_Host
	var roomIDs []uint

	mockUserClient.On("FindById", context.Background(), host.Id).Return(nil, errors.New("User is not found"))

	reservations, err := svc.GetActiveHostReservations(context.Background(), host.Id, roomIDs)

	assert.Error(t, err)
	assert.Nil(t, reservations)
	mockUserClient.AssertNumberOfCalls(t, "FindById", 1)
	mockUserClient.AssertExpectations(t)
	mockRepo.AssertNumberOfCalls(t, "FindReservationsByRoomID", 0)
	mockRepo.AssertExpectations(t)
}

func Test_GetActiveHostReservations_NoRoomsNoReservations(t *testing.T) {
	svc, mockRepo, mockUserClient, _ := CreateTestRoomService()

	host := DefaultUser_Host
	var roomIDs []uint

	mockUserClient.On("FindById", context.Background(), host.Id).Return(host, nil)

	reservations, err := svc.GetActiveHostReservations(context.Background(), host.Id, roomIDs)

	assert.NoError(t, err)
	assert.Nil(t, reservations)
	mockUserClient.AssertNumberOfCalls(t, "FindById", 1)
	mockUserClient.AssertExpectations(t)
	mockRepo.AssertNumberOfCalls(t, "FindReservationsByRoomID", 0)
	mockRepo.AssertExpectations(t)
}

func Test_GetActiveHostReservations_NoneRoomHasAnyReservation(t *testing.T) {
	svc, mockRepo, mockUserClient, _ := CreateTestRoomService()

	host := DefaultUser_Host
	roomIDs := []uint{1, 2}
	reservartion1 := DefaultReservation
	reservartion1.Cancelled = true
	reservations1 := []internal.Reservation{*reservartion1}
	reservartion2 := DefaultReservation
	reservartion2.Cancelled = true
	reservations2 := []internal.Reservation{*reservartion1}

	mockUserClient.On("FindById", context.Background(), host.Id).Return(host, nil)
	mockRepo.On("FindReservationsByRoomID", roomIDs[0]).Return(reservations1, nil)
	mockRepo.On("FindReservationsByRoomID", roomIDs[1]).Return(reservations2, nil)

	reservartionsGot, err := svc.GetActiveHostReservations(context.Background(), host.Id, roomIDs)

	assert.NoError(t, err)
	assert.Nil(t, reservartionsGot)
	mockUserClient.AssertNumberOfCalls(t, "FindById", 1)
	mockUserClient.AssertExpectations(t)
	mockRepo.AssertNumberOfCalls(t, "FindReservationsByRoomID", 2)
	mockRepo.AssertExpectations(t)
}

// This test automatically covers enough of `ExtractActiveReservations`,
// as its logic is exceptionally simple.
func Test_GetActiveHostReservations_Success(t *testing.T) {
	svc, mockRepo, mockUserClient, _ := CreateTestRoomService()

	host := DefaultUser_Host
	roomIDs := []uint{1, 2}
	reservartion1 := *DefaultReservation
	reservartion1.Cancelled = false
	reservartion1.DateTo = time.Now().Add(24 * time.Hour)
	reservations1 := []internal.Reservation{reservartion1}
	reservartion2 := *DefaultReservation
	reservartion2.Cancelled = true
	reservartion2.DateTo = time.Now().Add(24 * time.Hour)
	reservations2 := []internal.Reservation{reservartion2}

	mockUserClient.On("FindById", context.Background(), host.Id).Return(host, nil)
	mockRepo.On("FindReservationsByRoomID", roomIDs[0]).Return(reservations1, nil)
	mockRepo.On("FindReservationsByRoomID", roomIDs[1]).Return(reservations2, nil)

	reservartionsGot, err := svc.GetActiveHostReservations(context.Background(), host.Id, roomIDs)

	assert.NoError(t, err)
	assert.Equal(t, reservations1, reservartionsGot)
	mockUserClient.AssertNumberOfCalls(t, "FindById", 1)
	mockUserClient.AssertExpectations(t)
	mockRepo.AssertNumberOfCalls(t, "FindReservationsByRoomID", 2)
	mockRepo.AssertExpectations(t)
}
