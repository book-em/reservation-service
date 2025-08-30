package test

import (
	"bookem-reservation-service/client/roomclient"
	"bookem-reservation-service/client/userclient"
	"bookem-reservation-service/internal"
	"time"

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

func (r *MockReservationRepo) CreateRequest(req *internal.ReservationRequest) error {
	args := r.Called(req)
	return args.Error(0)
}

func (r *MockReservationRepo) DeleteRequest(id uint) error {
	args := r.Called(id)
	return args.Error(0)
}

func (r *MockReservationRepo) FindRequestsByRoomIDUpcoming(roomID uint, now time.Time) ([]internal.ReservationRequest, error) {
	args := r.Called(roomID, now)
	return args.Get(0).([]internal.ReservationRequest), args.Error(1)
}

func (r *MockReservationRepo) SetRequestStatus(id uint, status internal.ReservationRequestStatus) error {
	args := r.Called(id, status)
	return args.Error(0)
}

func (r *MockReservationRepo) RejectPendingRequestsInRange(roomID uint, from, to time.Time) error {
	args := r.Called(roomID, from, to)
	return args.Error(0)
}

func (r *MockReservationRepo) FindPendingRequestsByRoomID(roomID uint) ([]internal.ReservationRequest, error) {
	args := r.Called(roomID)
	return args.Get(0).([]internal.ReservationRequest), args.Error(1)
}

func (r *MockReservationRepo) FindPendingRequestsByGuestID(guestID uint) ([]internal.ReservationRequest, error) {
	args := r.Called(guestID)
	return args.Get(0).([]internal.ReservationRequest), args.Error(1)
}

func (r *MockReservationRepo) CreateReservation(res *internal.Reservation) error {
	args := r.Called(res)
	return args.Error(0)
}

func (r *MockReservationRepo) CancelReservation(id uint) error {
	args := r.Called(id)
	return args.Error(0)
}

func (r *MockReservationRepo) FindCancelledReservationsByGuestID(guestID uint) ([]internal.Reservation, error) {
	args := r.Called(guestID)
	return args.Get(0).([]internal.Reservation), args.Error(1)
}

func (r *MockReservationRepo) FindReservationsByRoomIDForDay(roomID uint, day time.Time) ([]internal.Reservation, error) {
	args := r.Called(roomID, day)
	return args.Get(0).([]internal.Reservation), args.Error(1)
}

func (r *MockReservationRepo) FindReservationsByGuestID(guestID uint) ([]internal.Reservation, error) {
	args := r.Called(guestID)
	return args.Get(0).([]internal.Reservation), args.Error(1)
}

func (r *MockReservationRepo) CountGuestCancellations(guestID uint) (int64, error) {
	args := r.Called(guestID)
	return args.Get(0).(int64), args.Error(1)
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

func (r *MockRoomClient) FindCurrentAvailabilityListOfRoom(roomId uint) (*roomclient.RoomAvailabilityListDTO, error) {
	args := r.Called(roomId)
	list, _ := args.Get(0).(*roomclient.RoomAvailabilityListDTO)
	return list, args.Error(1)
}

func (r *MockRoomClient) FindCurrentPricelistOfRoom(roomId uint) (*roomclient.RoomPriceListDTO, error) {
	args := r.Called(roomId)
	list, _ := args.Get(0).(*roomclient.RoomPriceListDTO)
	return list, args.Error(1)
}

func (r *MockRoomClient) QueryForReservation(jwt string, dto roomclient.RoomReservationQueryDTO) (*roomclient.RoomReservationQueryResponseDTO, error) {
	args := r.Called(jwt, dto)
	resp, _ := args.Get(0).(*roomclient.RoomReservationQueryResponseDTO)
	return resp, args.Error(1)
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

var DefaultRoom = &roomclient.RoomDTO{
	ID:        1,
	HostID:    2,
	Name:      "Test Room",
	MinGuests: 1,
	MaxGuests: 4,
}

var DefaultAvailabilityList = &roomclient.RoomAvailabilityListDTO{
	ID:     1,
	RoomID: 1,
	Items:  []roomclient.RoomAvailabilityItemDTO{},
}

var DefaultPriceList = &roomclient.RoomPriceListDTO{
	ID:        1,
	RoomID:    1,
	BasePrice: 100,
	PerGuest:  true,
	Items:     []roomclient.RoomPriceItemDTO{},
}

var DefaultReservationQueryResponse = &roomclient.RoomReservationQueryResponseDTO{
	Available: true,
	TotalCost: 400,
}
