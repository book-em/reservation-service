package test

import (
	"bookem-reservation-service/client/roomclient"
	"bookem-reservation-service/internal"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_CreateRequest_Success(t *testing.T) {
	svc, repo, userClient, roomClient := CreateTestRoomService()

	dto := internal.CreateReservationRequestDTO{
		RoomID:     1,
		DateFrom:   time.Now().AddDate(0, 0, 1),
		DateTo:     time.Now().AddDate(0, 0, 3),
		GuestCount: 2,
	}

	repo.On("FindPendingRequestsByGuestID", uint(1)).Return([]internal.ReservationRequest{}, nil)
	repo.On("FindReservationsByRoomIDForDay", mock.Anything, mock.Anything).Return([]internal.Reservation{}, nil)
	repo.On("CreateRequest", mock.AnythingOfType("*internal.ReservationRequest")).Return(nil)

	userClient.On("FindById", uint(1)).Return(DefaultUser_Guest, nil)
	roomClient.On("FindById", uint(1)).Return(DefaultRoom, nil)
	roomClient.On("FindCurrentAvailabilityListOfRoom", uint(1)).Return(DefaultAvailabilityList, nil)
	roomClient.On("FindCurrentPricelistOfRoom", uint(1)).Return(DefaultPriceList, nil)
	roomClient.On("QueryForReservation", mock.Anything, mock.Anything).Return(DefaultReservationQueryResponse, nil)

	auth := internal.AuthContext{CallerID: 1, JWT: "token"}
	req, err := svc.CreateRequest(auth, dto)

	assert.NoError(t, err)
	assert.NotNil(t, req)
	repo.AssertNumberOfCalls(t, "CreateRequest", 1)
}

func Test_CreateRequest_Unauthenticated(t *testing.T) {
	svc, _, userClient, _ := CreateTestRoomService()

	userClient.On("FindById", uint(1)).Return(nil, errors.New("not found"))

	auth := internal.AuthContext{CallerID: 1, JWT: "token"}
	dto := internal.CreateReservationRequestDTO{RoomID: 1}

	req, err := svc.CreateRequest(auth, dto)

	assert.ErrorIs(t, err, internal.ErrUnauthenticated)
	assert.Nil(t, req)
}

func Test_CreateRequest_UnauthorizedRole(t *testing.T) {
	svc, _, userClient, _ := CreateTestRoomService()

	userClient.On("FindById", uint(1)).Return(DefaultUser_Host, nil)

	auth := internal.AuthContext{CallerID: 1, JWT: "token"}
	dto := internal.CreateReservationRequestDTO{RoomID: 1}

	req, err := svc.CreateRequest(auth, dto)

	assert.ErrorIs(t, err, internal.ErrUnauthorized)
	assert.Nil(t, req)
}

func Test_CreateRequest_RoomNotFound(t *testing.T) {
	svc, _, userClient, roomClient := CreateTestRoomService()

	userClient.On("FindById", uint(1)).Return(DefaultUser_Guest, nil)
	roomClient.On("FindById", uint(1)).Return(nil, errors.New("not found"))

	auth := internal.AuthContext{CallerID: 1, JWT: "token"}
	dto := internal.CreateReservationRequestDTO{RoomID: 1}

	req, err := svc.CreateRequest(auth, dto)

	assert.ErrorContains(t, err, "room")
	assert.Nil(t, req)
}

func Test_CreateRequest_AvailabilityListNotFound(t *testing.T) {
	svc, _, userClient, roomClient := CreateTestRoomService()

	userClient.On("FindById", uint(1)).Return(DefaultUser_Guest, nil)
	roomClient.On("FindById", uint(1)).Return(DefaultRoom, nil)
	roomClient.On("FindCurrentAvailabilityListOfRoom", uint(1)).Return(nil, errors.New("not found"))

	auth := internal.AuthContext{CallerID: 1, JWT: "token"}
	dto := internal.CreateReservationRequestDTO{RoomID: 1}

	req, err := svc.CreateRequest(auth, dto)

	assert.ErrorContains(t, err, "room availability list")
	assert.Nil(t, req)
}

func Test_CreateRequest_PriceListNotFound(t *testing.T) {
	svc, _, userClient, roomClient := CreateTestRoomService()

	userClient.On("FindById", uint(1)).Return(DefaultUser_Guest, nil)
	roomClient.On("FindById", uint(1)).Return(DefaultRoom, nil)
	roomClient.On("FindCurrentAvailabilityListOfRoom", uint(1)).Return(DefaultAvailabilityList, nil)
	roomClient.On("FindCurrentPricelistOfRoom", uint(1)).Return(nil, errors.New("not found"))

	auth := internal.AuthContext{CallerID: 1, JWT: "token"}
	dto := internal.CreateReservationRequestDTO{RoomID: 1}

	req, err := svc.CreateRequest(auth, dto)

	assert.ErrorContains(t, err, "room price list")
	assert.Nil(t, req)
}

func Test_CreateRequest_RoomNotAvailable(t *testing.T) {
	svc, _, userClient, roomClient := CreateTestRoomService()

	unavailable := &roomclient.RoomReservationQueryResponseDTO{Available: false, TotalCost: 0}

	userClient.On("FindById", uint(1)).Return(DefaultUser_Guest, nil)
	roomClient.On("FindById", uint(1)).Return(DefaultRoom, nil)
	roomClient.On("FindCurrentAvailabilityListOfRoom", uint(1)).Return(DefaultAvailabilityList, nil)
	roomClient.On("FindCurrentPricelistOfRoom", uint(1)).Return(DefaultPriceList, nil)
	roomClient.On("QueryForReservation", mock.Anything, mock.Anything).Return(unavailable, nil)

	auth := internal.AuthContext{CallerID: 1, JWT: "token"}
	dto := internal.CreateReservationRequestDTO{
		RoomID:     1,
		DateFrom:   time.Now(),
		DateTo:     time.Now().AddDate(0, 0, 1),
		GuestCount: 2,
	}

	req, err := svc.CreateRequest(auth, dto)

	assert.ErrorIs(t, err, internal.ErrBadRequest)
	assert.Nil(t, req)
}

func Test_CreateRequest_InvalidGuestCount(t *testing.T) {
	svc, _, userClient, roomClient := CreateTestRoomService()

	userClient.On("FindById", uint(1)).Return(DefaultUser_Guest, nil)
	roomClient.On("FindById", uint(1)).Return(DefaultRoom, nil)
	roomClient.On("FindCurrentAvailabilityListOfRoom", uint(1)).Return(DefaultAvailabilityList, nil)
	roomClient.On("FindCurrentPricelistOfRoom", uint(1)).Return(DefaultPriceList, nil)
	roomClient.On("QueryForReservation", mock.Anything, mock.Anything).Return(DefaultReservationQueryResponse, nil)

	auth := internal.AuthContext{CallerID: 1, JWT: "token"}
	dto := internal.CreateReservationRequestDTO{
		RoomID:     1,
		DateFrom:   time.Now(),
		DateTo:     time.Now().AddDate(0, 0, 1),
		GuestCount: 0,
	}

	req, err := svc.CreateRequest(auth, dto)

	assert.ErrorContains(t, err, "guest count")
	assert.Nil(t, req)
}

func Test_CreateRequest_ReversedDates(t *testing.T) {
	svc, _, userClient, roomClient := CreateTestRoomService()

	userClient.On("FindById", uint(1)).Return(DefaultUser_Guest, nil)
	roomClient.On("FindById", uint(1)).Return(DefaultRoom, nil)
	roomClient.On("FindCurrentAvailabilityListOfRoom", uint(1)).Return(DefaultAvailabilityList, nil)
	roomClient.On("FindCurrentPricelistOfRoom", uint(1)).Return(DefaultPriceList, nil)
	roomClient.On("QueryForReservation", mock.Anything, mock.Anything).Return(DefaultReservationQueryResponse, nil)

	auth := internal.AuthContext{CallerID: 1, JWT: "token"}
	dto := internal.CreateReservationRequestDTO{
		RoomID:     1,
		DateFrom:   time.Now().AddDate(0, 0, 2),
		DateTo:     time.Now(),
		GuestCount: 2,
	}

	req, err := svc.CreateRequest(auth, dto)

	assert.ErrorContains(t, err, "dates are reversed")
	assert.Nil(t, req)
}

func Test_CreateRequest_ConflictDueToExistingRequestForUser(t *testing.T) {
	svc, repo, userClient, roomClient := CreateTestRoomService()

	existing := []internal.ReservationRequest{{RoomID: 1}}
	repo.On("FindPendingRequestsByGuestID", uint(1)).Return(existing, nil)

	userClient.On("FindById", uint(1)).Return(DefaultUser_Guest, nil)
	roomClient.On("FindById", uint(1)).Return(DefaultRoom, nil)
	roomClient.On("FindCurrentAvailabilityListOfRoom", uint(1)).Return(DefaultAvailabilityList, nil)
	roomClient.On("FindCurrentPricelistOfRoom", uint(1)).Return(DefaultPriceList, nil)
	roomClient.On("QueryForReservation", mock.Anything, mock.Anything).Return(DefaultReservationQueryResponse, nil)

	auth := internal.AuthContext{CallerID: 1, JWT: "token"}
	dto := internal.CreateReservationRequestDTO{
		RoomID:     1,
		DateFrom:   time.Now().AddDate(0, 0, 1),
		DateTo:     time.Now().AddDate(0, 0, 2),
		GuestCount: 2,
	}

	req, err := svc.CreateRequest(auth, dto)

	assert.ErrorIs(t, err, internal.ErrConflict)
	assert.Nil(t, req)
}

func Test_CreateRequest_ConflictDueToReservationOverlap(t *testing.T) {
	svc, repo, userClient, roomClient := CreateTestRoomService()

	userClient.On("FindById", uint(1)).Return(DefaultUser_Guest, nil)
	roomClient.On("FindById", uint(1)).Return(DefaultRoom, nil)
	roomClient.On("FindCurrentAvailabilityListOfRoom", uint(1)).Return(DefaultAvailabilityList, nil)
	roomClient.On("FindCurrentPricelistOfRoom", uint(1)).Return(DefaultPriceList, nil)
	roomClient.On("QueryForReservation", mock.Anything, mock.Anything).Return(DefaultReservationQueryResponse, nil)

	repo.On("FindPendingRequestsByGuestID", uint(1)).Return([]internal.ReservationRequest{}, nil)
	repo.On("FindReservationsByRoomIDForDay", mock.Anything, mock.Anything).Return([]internal.Reservation{{ID: 99}}, nil)

	auth := internal.AuthContext{CallerID: 1, JWT: "token"}
	dto := internal.CreateReservationRequestDTO{
		RoomID:     1,
		DateFrom:   time.Now(),
		DateTo:     time.Now().AddDate(0, 0, 1),
		GuestCount: 2,
	}

	req, err := svc.CreateRequest(auth, dto)

	assert.ErrorIs(t, err, internal.ErrConflict)
	assert.Nil(t, req)
}

func Test_CreateRequest_CreateFails(t *testing.T) {
	svc, repo, userClient, roomClient := CreateTestRoomService()

	userClient.On("FindById", uint(1)).Return(DefaultUser_Guest, nil)
	roomClient.On("FindById", uint(1)).Return(DefaultRoom, nil)
	roomClient.On("FindCurrentAvailabilityListOfRoom", uint(1)).Return(DefaultAvailabilityList, nil)
	roomClient.On("FindCurrentPricelistOfRoom", uint(1)).Return(DefaultPriceList, nil)
	roomClient.On("QueryForReservation", mock.Anything, mock.Anything).Return(DefaultReservationQueryResponse, nil)

	repo.On("FindPendingRequestsByGuestID", uint(1)).Return([]internal.ReservationRequest{}, nil)
	repo.On("FindReservationsByRoomIDForDay", mock.Anything, mock.Anything).Return([]internal.Reservation{}, nil)
	repo.On("CreateRequest", mock.AnythingOfType("*internal.ReservationRequest")).Return(errors.New("db error"))

	auth := internal.AuthContext{CallerID: 1, JWT: "token"}
	dto := internal.CreateReservationRequestDTO{
		RoomID:     1,
		DateFrom:   time.Now(),
		DateTo:     time.Now().AddDate(0, 0, 1),
		GuestCount: 2,
	}

	req, err := svc.CreateRequest(auth, dto)

	assert.ErrorContains(t, err, "db error")
	assert.Nil(t, req)
}
