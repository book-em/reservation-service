package test

import (
	"bookem-reservation-service/internal"
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_DeleteRequest_Success(t *testing.T) {
	svc, repo, userClient, _ := CreateTestRoomService()

	userClient.On("FindById", context.Background(), uint(1)).Return(DefaultUser_Guest, nil)
	repo.On("FindPendingRequestsByGuestID", uint(1)).Return([]internal.ReservationRequest{
		{ID: 10, GuestID: 1, RoomID: 1, Status: internal.Pending},
	}, nil)
	repo.On("DeleteRequest", uint(10)).Return(nil)

	err := svc.DeleteRequest(context.Background(), 1, 10)

	assert.NoError(t, err)
}

func Test_DeleteRequest_UserNotFound(t *testing.T) {
	svc, _, userClient, _ := CreateTestRoomService()

	userClient.On("FindById", context.Background(), uint(1)).Return(nil, errors.New("not found"))

	err := svc.DeleteRequest(context.Background(), 1, 10)

	assert.ErrorContains(t, err, "user")
}

func Test_DeleteRequest_UnauthorizedRole(t *testing.T) {
	svc, _, userClient, _ := CreateTestRoomService()

	userClient.On("FindById", context.Background(), uint(1)).Return(DefaultUser_Host, nil)

	err := svc.DeleteRequest(context.Background(), 1, 10)

	assert.ErrorIs(t, err, internal.ErrUnauthorized)
}

func Test_DeleteRequest_RequestNotFound(t *testing.T) {
	svc, repo, userClient, _ := CreateTestRoomService()

	userClient.On("FindById", context.Background(), uint(1)).Return(DefaultUser_Guest, nil)
	repo.On("FindPendingRequestsByGuestID", uint(1)).Return([]internal.ReservationRequest{
		{ID: 99, GuestID: 1, RoomID: 1, Status: internal.Pending},
	}, nil)

	err := svc.DeleteRequest(context.Background(), 1, 10)

	assert.ErrorContains(t, err, "reservation")
}

func Test_DeleteRequest_RequestAccepted(t *testing.T) {
	svc, repo, userClient, _ := CreateTestRoomService()

	userClient.On("FindById", context.Background(), uint(1)).Return(DefaultUser_Guest, nil)
	repo.On("FindPendingRequestsByGuestID", uint(1)).Return([]internal.ReservationRequest{
		{ID: 1, GuestID: 1, RoomID: 1, Status: internal.Accepted},
	}, nil)

	err := svc.DeleteRequest(context.Background(), 1, 1)

	assert.ErrorContains(t, err, "cannot cancel a handled request")
}

func Test_DeleteRequest_RequestRejected(t *testing.T) {
	svc, repo, userClient, _ := CreateTestRoomService()

	userClient.On("FindById", context.Background(), uint(1)).Return(DefaultUser_Guest, nil)
	repo.On("FindPendingRequestsByGuestID", uint(1)).Return([]internal.ReservationRequest{
		{ID: 1, GuestID: 1, RoomID: 1, Status: internal.Rejected},
	}, nil)

	err := svc.DeleteRequest(context.Background(), 1, 1)

	assert.ErrorContains(t, err, "cannot cancel a handled request")
}
