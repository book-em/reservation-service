package test

import (
	"bookem-reservation-service/client/roomclient"
	"bookem-reservation-service/client/userclient"
	"bookem-reservation-service/internal"

	mock "github.com/stretchr/testify/mock"
)

func CreateTestRoomService() (
	internal.Service,
	*MockReservationRepo,
	*MockUserClient,
	*MockRoomClient,
) {
	mockRepo := new(MockReservationRepo)
	mockUserClient := new(MockUserClient)
	mockRoomClient := new(MockRoomClient)

	svc := internal.NewService(mockRepo, mockUserClient, mockRoomClient)
	return svc, mockRepo, mockUserClient, mockRoomClient
}

// ----------------------------------------------- Mock Reservation repo

type MockReservationRepo struct {
	mock.Mock
}

func (r *MockReservationRepo) Create(room *internal.Reservation) error {
	args := r.Called(room)
	return args.Error(0)
}

// ----------------------------------------------- Mock user client

type MockUserClient struct {
	mock.Mock
}

func (r *MockUserClient) FindById(id uint) (*userclient.UserDTO, error) {
	args := r.Called(id)
	user, _ := args.Get(0).(*userclient.UserDTO)
	return user, args.Error(1)
}

// ----------------------------------------------- Mock room client

type MockRoomClient struct {
	mock.Mock
}

func (r *MockRoomClient) FindById(id uint) (*roomclient.RoomDTO, error) {
	args := r.Called(id)
	room, _ := args.Get(0).(*roomclient.RoomDTO)
	return room, args.Error(1)
}

// ----------------------------------------------- Mock data

var DefaultReservation = &internal.Reservation{
	ID:      0,
	GuestID: 1,
	RoomID:  1,
}

var DefaultReservationDTO = internal.ReservationDTO{
	ID:      0,
	GuestID: 1,
	RoomID:  1,
}

var DefaultUser_Guest = &userclient.UserDTO{
	Id:       1,
	Username: "guser",
	Email:    "gemail@mail.com",
	Name:     "gname",
	Surname:  "gsurname",
	Role:     "guest",
	Address:  "gAddress 123",
}

var DefaultUser_Host = &userclient.UserDTO{
	Id:       2,
	Username: "huser",
	Email:    "hemail@mail.com",
	Name:     "hname",
	Surname:  "hsurname",
	Role:     "host",
	Address:  "hAddress 123",
}
