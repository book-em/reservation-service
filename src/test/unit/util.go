package test

import (
	"bookem-reservation-service/client/notificationclient"
	"bookem-reservation-service/client/roomclient"
	"bookem-reservation-service/client/userclient"
	"bookem-reservation-service/internal"
	"context"
	"time"

	mock "github.com/stretchr/testify/mock"
)

func CreateTestRoomService() (
	internal.Service,
	*MockReservationRepo,
	*MockUserClient,
	*MockRoomClient,
	*MockNotificationClient,
) {
	mockRepo := new(MockReservationRepo)
	mockUserClient := new(MockUserClient)
	mockRoomClient := new(MockRoomClient)
	mockNotificationClient := new(MockNotificationClient)

	svc := internal.NewService(mockRepo, mockUserClient, mockRoomClient, mockNotificationClient)
	return svc, mockRepo, mockUserClient, mockRoomClient, mockNotificationClient
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

func (r *MockReservationRepo) FindReservationsByRoomID(roomID uint) ([]internal.Reservation, error) {
	args := r.Called(roomID)
	return args.Get(0).([]internal.Reservation), args.Error(1)
}

func (m *MockReservationRepo) FindRequestByID(id uint) (*internal.ReservationRequest, error) {
	args := m.Called(id)
	if req, ok := args.Get(0).(*internal.ReservationRequest); ok {
		return req, args.Error(1)
	}
	return nil, args.Error(1)
}

func (r *MockReservationRepo) HasGuestPastReservationInRooms(guestID uint, roomIDs []uint, now time.Time) (bool, error) {
    args := r.Called(guestID, roomIDs, now)
    return args.Bool(0), args.Error(1)
}


// ----------------------------------------------- Mock user client

type MockUserClient struct {
	mock.Mock
}

func (r *MockUserClient) FindById(context context.Context, id uint) (*userclient.UserDTO, error) {
	args := r.Called(context, id)
	user, _ := args.Get(0).(*userclient.UserDTO)
	return user, args.Error(1)
}

// ----------------------------------------------- Mock room client

type MockRoomClient struct {
	mock.Mock
}

func (r *MockRoomClient) FindById(context context.Context, id uint) (*roomclient.RoomDTO, error) {
	args := r.Called(context, id)
	room, _ := args.Get(0).(*roomclient.RoomDTO)
	return room, args.Error(1)
}

func (r *MockRoomClient) FindCurrentAvailabilityListOfRoom(context context.Context, roomId uint) (*roomclient.RoomAvailabilityListDTO, error) {
	args := r.Called(context, roomId)
	list, _ := args.Get(0).(*roomclient.RoomAvailabilityListDTO)
	return list, args.Error(1)
}

func (r *MockRoomClient) FindCurrentPricelistOfRoom(context context.Context, roomId uint) (*roomclient.RoomPriceListDTO, error) {
	args := r.Called(context, roomId)
	list, _ := args.Get(0).(*roomclient.RoomPriceListDTO)
	return list, args.Error(1)
}

func (r *MockRoomClient) QueryForReservation(context context.Context, jwt string, dto roomclient.RoomReservationQueryDTO) (*roomclient.RoomReservationQueryResponseDTO, error) {
	args := r.Called(context, jwt, dto)
	resp, _ := args.Get(0).(*roomclient.RoomReservationQueryResponseDTO)
	return resp, args.Error(1)
}

func (r *MockRoomClient) FindByHostId(context context.Context, id uint) ([]roomclient.RoomDTO, error) {
	args := r.Called(context, id)
	room, _ := args.Get(0).([]roomclient.RoomDTO)
	return room, args.Error(1)
}

func (r *MockReservationRepo) FindReservationById(id uint) (*internal.Reservation, error) {
	args := r.Called(id)
	if res, ok := args.Get(0).(*internal.Reservation); ok {
		return res, args.Error(1)
	}
	return nil, args.Error(1)
}

// ----------------------------------------------- Mock notification client

type MockNotificationClient struct {
	mock.Mock
}

func (m *MockNotificationClient) CreateNotification(ctx context.Context, jwt string, dto notificationclient.CreateNotificationDTO) (*notificationclient.NotificationDTO, error) {
	args := m.Called(ctx, jwt, dto)
	notification, _ := args.Get(0).(*notificationclient.NotificationDTO)
	return notification, args.Error(1)
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
