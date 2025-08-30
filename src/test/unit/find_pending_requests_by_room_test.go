package test

import (
	"bookem-reservation-service/internal"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_FindPendingRequestsByRoom_Success(t *testing.T) {
	svc, repo, userClient, roomClient := CreateTestRoomService()

	userClient.On("FindById", uint(2)).Return(DefaultUser_Host, nil)
	roomClient.On("FindById", uint(1)).Return(DefaultRoom, nil)
	repo.On("FindPendingRequestsByRoomID", uint(1)).Return([]internal.ReservationRequest{
		{ID: 1, RoomID: 1, GuestID: 1, Status: internal.Pending},
	}, nil)

	result, err := svc.FindPendingRequestsByRoom(2, 1)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
}

func Test_FindPendingRequestsByRoom_UserNotFound(t *testing.T) {
	svc, _, userClient, _ := CreateTestRoomService()

	userClient.On("FindById", uint(2)).Return(nil, errors.New("not found"))

	result, err := svc.FindPendingRequestsByRoom(2, 1)

	assert.ErrorContains(t, err, "user")
	assert.Nil(t, result)
}

func Test_FindPendingRequestsByRoom_UnauthorizedRole(t *testing.T) {
	svc, _, userClient, _ := CreateTestRoomService()

	userClient.On("FindById", uint(1)).Return(DefaultUser_Guest, nil)

	result, err := svc.FindPendingRequestsByRoom(1, 1)

	assert.ErrorIs(t, err, internal.ErrUnauthorized)
	assert.Nil(t, result)
}

func Test_FindPendingRequestsByRoom_RoomNotFound(t *testing.T) {
	svc, _, userClient, roomClient := CreateTestRoomService()

	userClient.On("FindById", uint(2)).Return(DefaultUser_Host, nil)
	roomClient.On("FindById", uint(1)).Return(nil, errors.New("not found"))

	result, err := svc.FindPendingRequestsByRoom(2, 1)

	assert.ErrorContains(t, err, "room")
	assert.Nil(t, result)
}

func Test_FindPendingRequestsByRoom_UnauthorizedOwnership(t *testing.T) {
	svc, _, userClient, roomClient := CreateTestRoomService()

	otherRoom := *DefaultRoom
	otherRoom.HostID = 99

	userClient.On("FindById", uint(2)).Return(DefaultUser_Host, nil)
	roomClient.On("FindById", uint(1)).Return(&otherRoom, nil)

	result, err := svc.FindPendingRequestsByRoom(2, 1)

	assert.ErrorIs(t, err, internal.ErrUnauthorized)
	assert.Nil(t, result)
}
