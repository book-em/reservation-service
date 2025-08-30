package test

import (
	"bookem-reservation-service/internal"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_FindPendingRequestsByGuest_Success(t *testing.T) {
	svc, repo, userClient, _ := CreateTestRoomService()

	userClient.On("FindById", uint(1)).Return(DefaultUser_Guest, nil)

	expected := []internal.ReservationRequest{
		{ID: 1, RoomID: 1, GuestID: 1, Status: internal.Pending},
		{ID: 2, RoomID: 2, GuestID: 1, Status: internal.Pending},
	}
	repo.On("FindPendingRequestsByGuestID", uint(1)).Return(expected, nil)

	result, err := svc.FindPendingRequestsByGuest(1)

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
	userClient.AssertCalled(t, "FindById", uint(1))
	repo.AssertCalled(t, "FindPendingRequestsByGuestID", uint(1))
}

func Test_FindPendingRequestsByGuest_UserNotFound(t *testing.T) {
	svc, _, userClient, _ := CreateTestRoomService()

	userClient.On("FindById", uint(1)).Return(nil, errors.New("not found"))

	result, err := svc.FindPendingRequestsByGuest(1)

	assert.Error(t, err)
	assert.Nil(t, result)
	userClient.AssertCalled(t, "FindById", uint(1))
}

func Test_FindPendingRequestsByGuest_UnauthorizedRole(t *testing.T) {
	svc, _, userClient, _ := CreateTestRoomService()

	userClient.On("FindById", uint(1)).Return(DefaultUser_Host, nil)

	result, err := svc.FindPendingRequestsByGuest(1)

	assert.ErrorIs(t, err, internal.ErrUnauthorized)
	assert.Nil(t, result)
	userClient.AssertCalled(t, "FindById", uint(1))
}
