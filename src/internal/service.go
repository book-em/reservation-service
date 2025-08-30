package internal

import (
	"bookem-reservation-service/client/roomclient"
	"bookem-reservation-service/client/userclient"
	"log"
	"time"
)

// AuthContext is used in cases where callerID is not enoug
type AuthContext struct {
	CallerID uint
	JWT      string
}

type Service interface {
	CreateRequest(authctx AuthContext, dto CreateReservationRequestDTO) (*ReservationRequest, error)

	// FindPendingRequestsByGuest answers the question:
	//
	// "which reservation requests have I created, that are not accepted or rejected?"
	FindPendingRequestsByGuest(callerID uint) ([]ReservationRequest, error)

	// FindPendingRequestsByRoom answers the question:
	//
	// "which reservation requests are there for this room, that are not accepted or rejected?"
	FindPendingRequestsByRoom(callerID uint, roomID uint) ([]ReservationRequest, error)

	// DeleteRequest removes a reservation request by the guest. This happens
	// when the guest changes his mind before the request has been processed
	// (accepted/rejected).
	DeleteRequest(callerID uint, requestID uint) error

	// AreThereNoReservationsOnDays checks if a room has no reservations in the
	// specified date range.
	//
	// Note that this is referring to RESERVATIONS and not RESERVATION REQUESTS.
	AreThereNoReservationsOnDays(roomID uint, from, to time.Time) (bool, error)
}

type service struct {
	repo       Repository
	userClient userclient.UserClient
	roomClient roomclient.RoomClient
}

func NewService(
	roomRepo Repository,
	userClient userclient.UserClient,
	roomClient roomclient.RoomClient) Service {
	return &service{roomRepo, userClient, roomClient}
}

func (s *service) CreateRequest(authctx AuthContext, dto CreateReservationRequestDTO) (*ReservationRequest, error) {
	callerID := authctx.CallerID
	jwt := authctx.JWT

	// [1] Find user

	user, err := s.userClient.FindById(callerID)
	if err != nil {
		return nil, ErrUnauthenticated
	}

	// [2] User must be a Guest

	if user.Role != string(userclient.Guest) {
		return nil, ErrUnauthorized
	}

	// [3] Find room

	room, err := s.roomClient.FindById(dto.RoomID)
	if err != nil {
		return nil, ErrNotFound("room", dto.RoomID)
	}

	// [4] Find room availability list

	availList, err := s.roomClient.FindCurrentAvailabilityListOfRoom(room.ID)
	if err != nil {
		return nil, ErrNotFound("room availability list", dto.RoomID)
	}

	// [5] Find room price list

	pricelist, err := s.roomClient.FindCurrentPricelistOfRoom(room.ID)
	if err != nil {
		return nil, ErrNotFound("room price list", dto.RoomID)
	}

	// [6] Query room for reservation data.

	queryDTO := roomclient.RoomReservationQueryDTO{
		RoomID:     room.ID,
		DateFrom:   dto.DateFrom,
		DateTo:     dto.DateTo,
		GuestCount: dto.GuestCount,
	}
	queryResponse, err := s.roomClient.QueryForReservation(jwt, queryDTO)
	if err != nil {
		return nil, ErrBadRequest
	}

	if !queryResponse.Available {
		log.Print("Room is not available at this time range")
		return nil, ErrBadRequest
	}

	// [7] Calculate price

	cost := queryResponse.TotalCost

	// [8] Validate fields

	if dto.GuestCount < 1 {
		return nil, ErrBadRequestCustom("guest count must be at least 1")
	}

	if dto.DateFrom.After(dto.DateTo) {
		return nil, ErrBadRequestCustom("dates are reversed")
	}

	// [9] Allow only 1 request per guest per room

	existing, err := s.repo.FindPendingRequestsByGuestID(callerID)
	if err != nil {
		return nil, err
	}
	for _, req := range existing {
		if req.RoomID == dto.RoomID {
			return nil, ErrConflict
		}
	}

	// [10] Check if an existing reservation exists for this time range

	ok, err := s.AreThereNoReservationsOnDays(dto.RoomID, dto.DateFrom, dto.DateTo)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrConflict
	}

	// [11] Create request

	req := &ReservationRequest{
		RoomID:             dto.RoomID,
		DateFrom:           dto.DateFrom,
		DateTo:             dto.DateTo,
		GuestCount:         dto.GuestCount,
		GuestID:            callerID,
		Status:             Pending,
		RoomAvailabilityID: availList.ID,
		RoomPriceID:        pricelist.ID,
		Cost:               cost,
	}
	if err := s.repo.CreateRequest(req); err != nil {
		return nil, err
	}
	return req, nil
}

func (s *service) FindPendingRequestsByGuest(callerID uint) ([]ReservationRequest, error) {
	// [1] Find user

	user, err := s.userClient.FindById(callerID)
	if err != nil {
		return nil, ErrNotFound("user", callerID)
	}

	// [2] User must be guest

	if user.Role != string(userclient.Guest) {
		return nil, ErrUnauthorized
	}

	// [3] Return

	return s.repo.FindPendingRequestsByGuestID(callerID)
}

func (s *service) FindPendingRequestsByRoom(callerID uint, roomID uint) ([]ReservationRequest, error) {
	// [1] Find user

	user, err := s.userClient.FindById(callerID)
	if err != nil {
		return nil, ErrNotFound("user", callerID)
	}

	// [2] User must be host

	if user.Role != string(userclient.Host) {
		return nil, ErrUnauthorized
	}

	// [3] Find room

	room, err := s.roomClient.FindById(roomID)
	if err != nil {
		return nil, ErrNotFound("room", roomID)
	}

	// [4] Host must be the owner of this room

	if room.HostID != callerID {
		return nil, ErrUnauthorized
	}

	// [5] Return

	return s.repo.FindPendingRequestsByRoomID(roomID)
}

func (s *service) DeleteRequest(callerID uint, requestID uint) error {
	// [1] Find user

	user, err := s.userClient.FindById(callerID)
	if err != nil {
		return ErrNotFound("user", callerID)
	}

	// [2] User must be guest

	if user.Role != string(userclient.Guest) {
		return ErrUnauthorized
	}

	// [3] Find request

	requests, err := s.repo.FindPendingRequestsByGuestID(callerID)
	if err != nil {
		return err
	}
	found := false
	var request *ReservationRequest
	for _, req := range requests {
		if req.ID == requestID {
			found = true
			request = &req
			break
		}
	}
	if !found {
		return ErrNotFound("reservation request", requestID)
	}

	// [4] Request must be pending

	if request.Status != Pending {
		return ErrBadRequestCustom("cannot cancel a handled request")
	}

	// [5] Delete

	return s.repo.DeleteRequest(requestID)
}

func (s *service) AreThereNoReservationsOnDays(roomID uint, from, to time.Time) (bool, error) {
	for d := from; !d.After(to); d = d.AddDate(0, 0, 1) {
		reservations, err := s.repo.FindReservationsByRoomIDForDay(roomID, d)
		if err != nil {
			return false, err
		}
		if len(reservations) > 0 {
			return false, nil
		}
	}
	return true, nil
}
