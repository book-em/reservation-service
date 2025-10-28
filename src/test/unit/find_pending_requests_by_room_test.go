package test

import (
	"bookem-reservation-service/internal"
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_FindPendingRequestsByRoom_Success(t *testing.T) {
	svc, repo, userClient, roomClient, _ := CreateTestRoomService()

	guest1Val := *DefaultUser_Guest
	guest1 := &guest1Val
	guest1.Id = 3
	guest1.Deleted = true

	guest2Val := *DefaultUser_Guest
	guest2 := &guest2Val
	guest2.Id = 4
	guest2.Deleted = false

	userClient.On("FindById", context.Background(), uint(2)).Return(DefaultUser_Host, nil).Once()
	roomClient.On("FindById", context.Background(), uint(1)).Return(DefaultRoom, nil)
	repo.On("FindPendingRequestsByRoomID", uint(1)).Return([]internal.ReservationRequest{
		{ID: 1, RoomID: 1, GuestID: guest1.Id, Status: internal.Pending},
		{ID: 2, RoomID: 1, GuestID: guest2.Id, Status: internal.Pending},
	}, nil)
	userClient.On("FindById", context.Background(), guest1.Id).Return(guest1, nil).Once()
	userClient.On("FindById", context.Background(), guest2.Id).Return(guest2, nil).Once()

	result, err := svc.FindPendingRequestsByRoom(context.Background(), 2, 1)

	assert.NoError(t, err)
	assert.Equal(t, 1, len(result))
	assert.Equal(t, guest2.Id, result[0].GuestID)
}

func Test_FindPendingRequestsByRoom_UserNotFound(t *testing.T) {
	svc, _, userClient, _, _ := CreateTestRoomService()

	userClient.On("FindById", context.Background(), uint(2)).Return(nil, errors.New("not found"))

	result, err := svc.FindPendingRequestsByRoom(context.Background(), 2, 1)

	assert.ErrorContains(t, err, "user")
	assert.Nil(t, result)
}

func Test_FindPendingRequestsByRoom_UnauthorizedRole(t *testing.T) {
	svc, _, userClient, _, _ := CreateTestRoomService()

	userClient.On("FindById", context.Background(), uint(1)).Return(DefaultUser_Guest, nil)

	result, err := svc.FindPendingRequestsByRoom(context.Background(), 1, 1)

	assert.ErrorIs(t, err, internal.ErrUnauthorized)
	assert.Nil(t, result)
}

func Test_FindPendingRequestsByRoom_RoomNotFound(t *testing.T) {
	svc, _, userClient, roomClient, _ := CreateTestRoomService()

	userClient.On("FindById", context.Background(), uint(2)).Return(DefaultUser_Host, nil)
	roomClient.On("FindById", context.Background(), uint(1)).Return(nil, errors.New("not found"))

	result, err := svc.FindPendingRequestsByRoom(context.Background(), 2, 1)

	assert.ErrorContains(t, err, "room")
	assert.Nil(t, result)
}

func Test_FindPendingRequestsByRoom_UnauthorizedOwnership(t *testing.T) {
	svc, _, userClient, roomClient, _ := CreateTestRoomService()

	otherRoom := *DefaultRoom
	otherRoom.HostID = 99

	userClient.On("FindById", context.Background(), uint(2)).Return(DefaultUser_Host, nil)
	roomClient.On("FindById", context.Background(), uint(1)).Return(&otherRoom, nil)

	result, err := svc.FindPendingRequestsByRoom(context.Background(), 2, 1)

	assert.ErrorIs(t, err, internal.ErrUnauthorized)
	assert.Nil(t, result)
}
