package test

import (
	"bookem-reservation-service/internal"
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_ApproveReservationRequest_Success(t *testing.T) {
	svc, repo, _, roomClient := CreateTestRoomService()

	req := &internal.ReservationRequest{ID: 1, RoomID: 1, GuestID: 1, GuestCount: 2}
	room := *DefaultRoom
	room.HostID = 2

	repo.On("FindRequestByID", uint(1)).Return(req, nil)
	roomClient.On("FindById", context.Background(), uint(1)).Return(&room, nil)
	roomClient.On("FindCurrentAvailabilityListOfRoom", context.Background(), uint(1)).Return(DefaultAvailabilityList, nil)
	roomClient.On("FindCurrentPricelistOfRoom", context.Background(), uint(1)).Return(DefaultPriceList, nil)
	roomClient.On("QueryForReservation", mock.Anything, mock.Anything, mock.Anything).Return(DefaultReservationQueryResponse, nil)
	repo.On("CreateReservation", mock.AnythingOfType("*internal.Reservation")).Return(nil)
	repo.On("SetRequestStatus", uint(1), internal.Accepted).Return(nil)
	repo.On("RejectPendingRequestsInRange", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	callerID := 2
	err := svc.ApproveReservationRequest(context.Background(), uint(callerID), 1)

	assert.NoError(t, err)
	repo.AssertCalled(t, "SetRequestStatus", uint(1), internal.Accepted)
}

func Test_ApproveReservationRequest_RequestNotFound(t *testing.T) {
	svc, repo, _, _ := CreateTestRoomService()

	repo.On("FindRequestByID", uint(1)).Return(nil, errors.New("not found"))

	callerID := 2
	err := svc.ApproveReservationRequest(context.Background(), uint(callerID), 1)

	assert.Error(t, err)
}

func Test_ApproveReservationRequest_UnauthorizedHost(t *testing.T) {
	svc, repo, _, roomClient := CreateTestRoomService()

	req := &internal.ReservationRequest{ID: 1, RoomID: 1}
	room := *DefaultRoom
	room.HostID = 99

	repo.On("FindRequestByID", uint(1)).Return(req, nil)
	roomClient.On("FindById", context.Background(), uint(1)).Return(&room, nil)

	callerID := 2
	err := svc.ApproveReservationRequest(context.Background(), uint(callerID), 1)

	assert.ErrorIs(t, err, internal.ErrUnauthorized)
}

func Test_ApproveReservationRequest_CreateReservationFails(t *testing.T) {
	svc, repo, _, roomClient := CreateTestRoomService()

	req := &internal.ReservationRequest{ID: 1, RoomID: 1, GuestID: 1, GuestCount: 2}
	room := *DefaultRoom
	room.HostID = 2

	repo.On("FindRequestByID", uint(1)).Return(req, nil)
	roomClient.On("FindById", context.Background(), uint(1)).Return(&room, nil)
	roomClient.On("FindCurrentAvailabilityListOfRoom", context.Background(), uint(1)).Return(DefaultAvailabilityList, nil)
	roomClient.On("FindCurrentPricelistOfRoom", context.Background(), uint(1)).Return(DefaultPriceList, nil)

	repo.On("CreateReservation", mock.AnythingOfType("*internal.Reservation")).Return(errors.New("db create failed"))

	callerID := 2
	err := svc.ApproveReservationRequest(context.Background(), uint(callerID), 1)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "db create failed")
}
