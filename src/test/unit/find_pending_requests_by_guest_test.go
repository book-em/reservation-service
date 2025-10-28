package test

import (
	"bookem-reservation-service/internal"
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_FindPendingRequestsByGuest_Success(t *testing.T) {
	svc, repo, userClient, roomClient, _ := CreateTestRoomService()

	room1Val := *DefaultRoom
	room1 := &room1Val
	room1.ID = 1
	room1.Deleted = true

	room2Val := *DefaultRoom
	room2 := &room2Val
	room2.ID = 2
	room2.Deleted = false

	userClient.On("FindById", context.Background(), uint(1)).Return(DefaultUser_Guest, nil)
	roomClient.On("FindById", context.Background(), room1.ID).Return(room1, nil).Once()
	roomClient.On("FindById", context.Background(), room2.ID).Return(room2, nil).Once()

	expected := []internal.ReservationRequest{
		{ID: 1, RoomID: room1.ID, GuestID: 1, Status: internal.Pending},
		{ID: 2, RoomID: room2.ID, GuestID: 1, Status: internal.Pending},
	}
	repo.On("FindPendingRequestsByGuestID", uint(1)).Return(expected, nil)

	result, err := svc.FindPendingRequestsByGuest(context.Background(), 1)

	assert.NoError(t, err)
	assert.Equal(t, 1, len(result))
	assert.Equal(t, room2.ID, result[0].ID)
	repo.AssertCalled(t, "FindPendingRequestsByGuestID", uint(1))
}

func Test_FindPendingRequestsByGuest_UserNotFound(t *testing.T) {
	svc, _, userClient, _, _ := CreateTestRoomService()

	userClient.On("FindById", context.Background(), uint(1)).Return(nil, errors.New("not found"))

	result, err := svc.FindPendingRequestsByGuest(context.Background(), 1)

	assert.Error(t, err)
	assert.Nil(t, result)
}

func Test_FindPendingRequestsByGuest_UnauthorizedRole(t *testing.T) {
	svc, _, userClient, _, _ := CreateTestRoomService()

	userClient.On("FindById", context.Background(), uint(1)).Return(DefaultUser_Host, nil)

	result, err := svc.FindPendingRequestsByGuest(context.Background(), 1)

	assert.ErrorIs(t, err, internal.ErrUnauthorized)
	assert.Nil(t, result)
}
