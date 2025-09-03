package internal

import (
	"bookem-reservation-service/client/roomclient"
	"bookem-reservation-service/client/userclient"
	"bookem-reservation-service/util"
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

	// AreThereReservationsOnDays checks if a room has reservations in the
	// specified date range.
	//
	// Note that this is referring to RESERVATIONS and not RESERVATION REQUESTS.
	AreThereReservationsOnDays(roomID uint, from, to time.Time) (bool, error)
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

	log.Print("CreateRequest [1] Find user")

	user, err := s.userClient.FindById(callerID)
	if err != nil {
		return nil, ErrUnauthenticated
	}

	log.Print("CreateRequest [2] User must be a Guest")

	if user.Role != string(userclient.Guest) {
		return nil, ErrUnauthorized
	}

	log.Print("CreateRequest [3] Find room")

	room, err := s.roomClient.FindById(dto.RoomID)
	if err != nil {
		return nil, ErrNotFound("room", dto.RoomID)
	}

	log.Print("CreateRequest [4] Find room availability list")

	availList, err := s.roomClient.FindCurrentAvailabilityListOfRoom(room.ID)
	if err != nil {
		return nil, ErrNotFound("room availability list", dto.RoomID)
	}

	log.Print("CreateRequest [5] Find room price list")

	pricelist, err := s.roomClient.FindCurrentPricelistOfRoom(room.ID)
	if err != nil {
		return nil, ErrNotFound("room price list", dto.RoomID)
	}

	log.Print("CreateRequest [6] Query room for reservation data")

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

	log.Print("CreateRequest [7] Calculate price")

	cost := queryResponse.TotalCost

	log.Print("CreateRequest [8] Validate fields")

	if dto.GuestCount < 1 {
		return nil, ErrBadRequestCustom("guest count must be at least 1")
	}

	if dto.DateFrom.After(dto.DateTo) {
		return nil, ErrBadRequestCustom("dates are reversed")
	}

	log.Print("CreateRequest [9] Prevent overlapping requests for the same room and same guest")

	existing, err := s.repo.FindPendingRequestsByGuestID(callerID)
	if err != nil {
		return nil, err
	}
	for _, req := range existing {
		if req.RoomID == dto.RoomID {
			if util.AreDatesIntersecting(req.DateFrom, req.DateTo, dto.DateFrom, dto.DateTo) {
				return nil, ErrConflict
			}
		}
	}

	log.Print("CreateRequest [10] Check if this room has a reservation for this date range")

	has, err := s.AreThereReservationsOnDays(dto.RoomID, dto.DateFrom, dto.DateTo)
	if err != nil {
		return nil, err
	}
	if has {
		return nil, ErrConflict
	}

	log.Print("CreateRequest [11] Create request")

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
	log.Print("FindPendingRequestsByGuest [1] Find user")

	user, err := s.userClient.FindById(callerID)
	if err != nil {
		return nil, ErrNotFound("user", callerID)
	}

	log.Print("FindPendingRequestsByGuest [2] User must be guest")

	if user.Role != string(userclient.Guest) {
		return nil, ErrUnauthorized
	}

	log.Print("FindPendingRequestsByGuest [3] Return")

	return s.repo.FindPendingRequestsByGuestID(callerID)
}

func (s *service) FindPendingRequestsByRoom(callerID uint, roomID uint) ([]ReservationRequest, error) {
	log.Print("FindPendingRequestsByRoom [1] User must be guest")

	user, err := s.userClient.FindById(callerID)
	if err != nil {
		return nil, ErrNotFound("user", callerID)
	}

	log.Print("FindPendingRequestsByRoom [2] User must be host")

	if user.Role != string(userclient.Host) {
		return nil, ErrUnauthorized
	}

	log.Print("FindPendingRequestsByRoom [3] Find room")

	room, err := s.roomClient.FindById(roomID)
	if err != nil {
		return nil, ErrNotFound("room", roomID)
	}

	log.Print("FindPendingRequestsByRoom [4] Host must be the owner of this room")

	if room.HostID != callerID {
		return nil, ErrUnauthorized
	}

	log.Print("FindPendingRequestsByRoom [5] Return")

	return s.repo.FindPendingRequestsByRoomID(roomID)
}

func (s *service) DeleteRequest(callerID uint, requestID uint) error {
	log.Print("DeleteRequest [1] Find user")

	user, err := s.userClient.FindById(callerID)
	if err != nil {
		return ErrNotFound("user", callerID)
	}

	log.Print("DeleteRequest [2] User must be guest")

	if user.Role != string(userclient.Guest) {
		return ErrUnauthorized
	}

	log.Print("DeleteRequest [3] Find request")

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

	log.Print("DeleteRequest [4] Request must be pending")

	if request.Status != Pending {
		return ErrBadRequestCustom("cannot cancel a handled request")
	}

	log.Print("DeleteRequest [5] Delete")

	return s.repo.DeleteRequest(requestID)
}

func (s *service) AreThereReservationsOnDays(roomID uint, from, to time.Time) (bool, error) {
	log.Printf("AreThereReservationsOnDays [1] Checking if room %d has a reservation from %s to %s", roomID, from.String(), to.String())

	for d := from; !d.After(to); d = d.AddDate(0, 0, 1) {
		log.Printf("AreThereReservationsOnDays [1.x] Checking if room %d has a reservation on day %s", roomID, d.String())
		reservations, err := s.repo.FindReservationsByRoomIDForDay(roomID, d)
		if err != nil {
			log.Printf("AreThereReservationsOnDays [1.x] Error %s", err.Error())
			return false, err
		}
		if len(reservations) > 0 {
			log.Printf("AreThereReservationsOnDays [1.x] Reservation found on day %s", d.String())
			return true, nil
		}
	}

	log.Printf("AreThereReservationsOnDays [2] OK, no reservations found for room %d", roomID)

	return false, nil
}
