package test

import (
	"bookem-reservation-service/client/roomclient"
	"bookem-reservation-service/internal"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_GetActiveHostReservations_FindRoomsError(t *testing.T) {
	svc, mockRepo, _, mockRoomClient := CreateTestRoomService()

	host := DefaultUser_Host

	mockRoomClient.On("FindByHostId", context.Background(), host.Id).Return(nil, errors.New("rooms are not found"))

	reservations, err := svc.GetActiveHostReservations(context.Background(), host.Id)

	assert.Error(t, err)
	assert.Nil(t, reservations)
	mockRoomClient.AssertNumberOfCalls(t, "FindByHostId", 1)
	mockRoomClient.AssertExpectations(t)
	mockRepo.AssertNumberOfCalls(t, "FindReservationsByRoomID", 0)
	mockRepo.AssertExpectations(t)
}

func Test_GetActiveHostReservations_NoRoomsSuccess(t *testing.T) {
	svc, mockRepo, _, mockRoomClient := CreateTestRoomService()

	host := DefaultUser_Host

	mockRoomClient.On("FindByHostId", context.Background(), host.Id).Return(nil, nil)

	reservations, err := svc.GetActiveHostReservations(context.Background(), host.Id)

	assert.NoError(t, err)
	assert.Nil(t, reservations)
	mockRoomClient.AssertNumberOfCalls(t, "FindByHostId", 1)
	mockRoomClient.AssertExpectations(t)
	mockRepo.AssertNumberOfCalls(t, "FindReservationsByRoomID", 0)
	mockRepo.AssertExpectations(t)
}

func Test_GetActiveHostReservations_NoneRoomHasAnyReservation(t *testing.T) {
	svc, mockRepo, _, mockRoomClient := CreateTestRoomService()

	host := DefaultUser_Host
	room1 := *DefaultRoom
	room1.ID = 1
	room2 := *DefaultRoom
	room2.ID = 2
	rooms := []roomclient.RoomDTO{room1, room2}
	reservartion1 := *DefaultReservation
	reservartion1.ID = room1.ID
	reservartion1.Cancelled = true
	reservations1 := []internal.Reservation{reservartion1}
	reservartion2 := *DefaultReservation
	reservartion2.ID = room2.ID
	reservartion2.Cancelled = true
	reservations2 := []internal.Reservation{reservartion2}

	mockRoomClient.On("FindByHostId", context.Background(), host.Id).Return(rooms, nil)
	mockRepo.On("FindReservationsByRoomID", room1.ID).Return(reservations1, nil)
	mockRepo.On("FindReservationsByRoomID", room2.ID).Return(reservations2, nil)

	reservartionsGot, err := svc.GetActiveHostReservations(context.Background(), host.Id)

	assert.NoError(t, err)
	assert.Nil(t, reservartionsGot)
	mockRoomClient.AssertNumberOfCalls(t, "FindByHostId", 1)
	mockRoomClient.AssertExpectations(t)
	mockRepo.AssertNumberOfCalls(t, "FindReservationsByRoomID", 2)
	mockRepo.AssertExpectations(t)
}

func Test_GetActiveHostReservations_ReservationsErr(t *testing.T) {
	svc, mockRepo, _, mockRoomClient := CreateTestRoomService()

	host := DefaultUser_Host
	room1 := *DefaultRoom
	room1.ID = 1
	rooms := []roomclient.RoomDTO{room1}
	var nilReservations []internal.Reservation = nil

	mockRoomClient.On("FindByHostId", context.Background(), host.Id).Return(rooms, nil)
	mockRepo.On("FindReservationsByRoomID", room1.ID).Return(nilReservations, errors.New("reservations are not found"))

	reservartionsGot, err := svc.GetActiveHostReservations(context.Background(), host.Id)

	assert.Error(t, err)
	assert.Nil(t, reservartionsGot)
	mockRoomClient.AssertNumberOfCalls(t, "FindByHostId", 1)
	mockRoomClient.AssertExpectations(t)
	mockRepo.AssertNumberOfCalls(t, "FindReservationsByRoomID", 1)
	mockRepo.AssertExpectations(t)
}

// This test automatically covers enough of `ExtractActiveReservations`,
// as its logic is exceptionally simple.
func Test_GetActiveHostReservations_Success(t *testing.T) {
	svc, mockRepo, _, mockRoomClient := CreateTestRoomService()

	host := DefaultUser_Host
	room1 := *DefaultRoom
	room1.ID = 1
	room2 := *DefaultRoom
	room2.ID = 2
	rooms := []roomclient.RoomDTO{room1, room2}
	reservation1 := *DefaultReservation
	reservation1.RoomID = room1.ID
	reservation1.Cancelled = false
	reservation1.DateTo = time.Now().Add(24 * time.Hour)
	reservations1 := []internal.Reservation{reservation1}
	reservation2 := *DefaultReservation
	reservation2.RoomID = room2.ID
	reservation2.Cancelled = true
	reservation2.DateTo = time.Now().Add(24 * time.Hour)
	reservations2 := []internal.Reservation{reservation2}

	mockRoomClient.On("FindByHostId", context.Background(), host.Id).Return(rooms, nil)
	mockRepo.On("FindReservationsByRoomID", room1.ID).Return(reservations1, nil)
	mockRepo.On("FindReservationsByRoomID", room2.ID).Return(reservations2, nil)

	reservationsGot, err := svc.GetActiveHostReservations(context.Background(), host.Id)

	assert.NoError(t, err)
	assert.Equal(t, reservations1, reservationsGot)
	mockRoomClient.AssertNumberOfCalls(t, "FindByHostId", 1)
	mockRoomClient.AssertExpectations(t)
	mockRepo.AssertNumberOfCalls(t, "FindReservationsByRoomID", 2)
	mockRepo.AssertExpectations(t)
}
