package test

import (
	"bookem-reservation-service/internal"
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_RejectReservationRequest_Success(t *testing.T) {
	svc, repo, _, roomClient, _ := CreateTestRoomService()

	req := &internal.ReservationRequest{ID: 1, RoomID: 1}
	room := *DefaultRoom
	room.HostID = 2

	repo.On("FindRequestByID", uint(1)).Return(req, nil)
	roomClient.On("FindById", context.Background(), uint(1)).Return(&room, nil)
	repo.On("SetRequestStatus", uint(1), internal.Rejected).Return(nil)

	callerID := 2
	err := svc.RejectReservationRequest(context.Background(), uint(callerID), 1)

	assert.NoError(t, err)
	repo.AssertCalled(t, "SetRequestStatus", uint(1), internal.Rejected)
}

func Test_RejectReservationRequest_RequestNotFound(t *testing.T) {
	svc, repo, _, _, _ := CreateTestRoomService()

	repo.On("FindRequestByID", uint(1)).Return(nil, errors.New("not found"))

	callerID := 2
	err := svc.RejectReservationRequest(context.Background(), uint(callerID), 1)

	assert.Error(t, err)
}

func Test_RejectReservationRequest_UnauthorizedHost(t *testing.T) {
	svc, repo, _, roomClient, _ := CreateTestRoomService()

	req := &internal.ReservationRequest{ID: 1, RoomID: 1}
	room := *DefaultRoom
	room.HostID = 99

	repo.On("FindRequestByID", uint(1)).Return(req, nil)
	roomClient.On("FindById", context.Background(), uint(1)).Return(&room, nil)

	callerID := 2
	err := svc.RejectReservationRequest(context.Background(), uint(callerID), 1)

	assert.ErrorIs(t, err, internal.ErrUnauthorized)
}

func Test_RejectReservationRequest_SetStatusFails(t *testing.T) {
	svc, repo, _, roomClient, _ := CreateTestRoomService()

	req := &internal.ReservationRequest{ID: 1, RoomID: 1}
	room := *DefaultRoom
	room.HostID = 2

	repo.On("FindRequestByID", uint(1)).Return(req, nil)
	roomClient.On("FindById", context.Background(), uint(1)).Return(&room, nil)
	repo.On("SetRequestStatus", uint(1), internal.Rejected).Return(errors.New("db error"))

	callerID := 2
	err := svc.RejectReservationRequest(context.Background(), uint(callerID), 1)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "db error")
}
