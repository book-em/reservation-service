package test

import (
	"bookem-reservation-service/client/notificationclient"
	"bookem-reservation-service/internal"
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_RejectReservationRequest_Success(t *testing.T) {
	svc, repo, userClient, roomClient, notifClient := CreateTestRoomService()

	req := &internal.ReservationRequest{ID: 1, RoomID: 1}
	req.GuestID = 11
	room := *DefaultRoom
	room.HostID = 2
	guestVal := *DefaultUser_Guest
	guest := &guestVal
	guest.Id = 11
	guest.Deleted = false

	repo.On("FindRequestByID", uint(1)).Return(req, nil)
	roomClient.On("FindById", context.Background(), uint(1)).Return(&room, nil)
	userClient.On("FindById", context.Background(), req.GuestID).Return(guest, nil)
	repo.On("SetRequestStatus", uint(1), internal.Rejected).Return(nil)

	callerID := 2
	notifClient.On("CreateNotification", mock.Anything, mock.Anything, mock.Anything).
		Return(&notificationclient.NotificationDTO{}, nil)
	err := svc.RejectReservationRequest(context.Background(), uint(callerID), 1, "Token")

	assert.NoError(t, err)
	repo.AssertCalled(t, "SetRequestStatus", uint(1), internal.Rejected)
}

func Test_RejectReservationRequest_RequestNotFound(t *testing.T) {
	svc, repo, _, _, _ := CreateTestRoomService()

	repo.On("FindRequestByID", uint(1)).Return(nil, errors.New("not found"))

	callerID := 2
	err := svc.RejectReservationRequest(context.Background(), uint(callerID), 1, "Token")

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
	err := svc.RejectReservationRequest(context.Background(), uint(callerID), 1, "Token")

	assert.ErrorIs(t, err, internal.ErrUnauthorized)
}

func Test_RejectReservationRequest_GuestNotFound(t *testing.T) {
	svc, repo, userClient, roomClient, _ := CreateTestRoomService()

	req := &internal.ReservationRequest{ID: 1, RoomID: 1}
	req.GuestID = 11
	roomVal := *DefaultRoom
	room := &roomVal
	room.HostID = 99
	guestVal := *DefaultUser_Guest
	guest := &guestVal
	guest.Id = 11
	guest.Deleted = true

	repo.On("FindRequestByID", req.ID).Return(req, nil)
	roomClient.On("FindById", context.Background(), req.RoomID).Return(room, nil)
	userClient.On("FindById", context.Background(), req.GuestID).Return(guest, nil)

	err := svc.RejectReservationRequest(context.Background(), room.HostID, req.ID, "Token")

	assert.Error(t, err)
}

func Test_RejectReservationRequest_GuestDbErr(t *testing.T) {
	svc, repo, userClient, roomClient, _ := CreateTestRoomService()

	req := &internal.ReservationRequest{ID: 1, RoomID: 1}
	req.GuestID = 11
	roomVal := *DefaultRoom
	room := &roomVal
	room.HostID = 99

	repo.On("FindRequestByID", req.ID).Return(req, nil)
	roomClient.On("FindById", context.Background(), req.RoomID).Return(room, nil)
	userClient.On("FindById", context.Background(), req.GuestID).Return(nil, errors.New("db create failed"))

	err := svc.RejectReservationRequest(context.Background(), room.HostID, req.ID, "Token")

	assert.Error(t, err)
}

func Test_RejectReservationRequest_SetStatusFails(t *testing.T) {
	svc, repo, userClient, roomClient, _ := CreateTestRoomService()

	req := &internal.ReservationRequest{ID: 1, RoomID: 1}
	req.GuestID = 11
	room := *DefaultRoom
	room.HostID = 2
	guestVal := *DefaultUser_Guest
	guest := &guestVal
	guest.Id = 11
	guest.Deleted = false

	repo.On("FindRequestByID", uint(1)).Return(req, nil)
	roomClient.On("FindById", context.Background(), uint(1)).Return(&room, nil)
	userClient.On("FindById", context.Background(), req.GuestID).Return(guest, nil)
	repo.On("SetRequestStatus", uint(1), internal.Rejected).Return(errors.New("db error"))

	err := svc.RejectReservationRequest(context.Background(), room.HostID, 1, "Token")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "db error")
}
