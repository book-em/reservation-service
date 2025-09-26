package internal

import (
	"bookem-reservation-service/client/roomclient"
	"bookem-reservation-service/client/userclient"
	"bookem-reservation-service/util"
	"context"
	"log"
	"time"
)

// AuthContext is used in cases where callerID is not enoug
type AuthContext struct {
	CallerID uint
	JWT      string
}

type Service interface {
	CreateRequest(context context.Context, authctx AuthContext, dto CreateReservationRequestDTO) (*ReservationRequest, error)

	// FindPendingRequestsByGuest answers the question:
	//
	// "which reservation requests have I created, that are not accepted or rejected?"
	FindPendingRequestsByGuest(context context.Context, callerID uint) ([]ReservationRequest, error)

	// FindPendingRequestsByRoom answers the question:
	//
	// "which reservation requests are there for this room, that are not accepted or rejected?"
	FindPendingRequestsByRoom(context context.Context, callerID uint, roomID uint) ([]ReservationRequest, error)

	// DeleteRequest removes a reservation request by the guest. This happens
	// when the guest changes his mind before the request has been processed
	// (accepted/rejected).
	DeleteRequest(context context.Context, callerID uint, requestID uint) error

	// AreThereReservationsOnDays checks if a room has reservations in the
	// specified date range.
	//
	// Note that this is referring to RESERVATIONS and not RESERVATION REQUESTS.
	AreThereReservationsOnDays(context context.Context, roomID uint, from, to time.Time) (bool, error)

	// GetGuestActiveReservations checks whether a guest has any active reservations
	// now or in the future.
	GetGuestActiveReservations(guestID uint) ([]Reservation, error)
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

func (s *service) CreateRequest(context context.Context, authctx AuthContext, dto CreateReservationRequestDTO) (*ReservationRequest, error) {
	callerID := authctx.CallerID
	jwt := authctx.JWT

	util.TEL.Info("user wants to create a reservation request", nil, "caller_id", authctx.CallerID)

	util.TEL.Push(context, "validate-room-and-user")
	defer util.TEL.Pop()

	util.TEL.Debug("check if user exists", nil, "id", callerID)
	user, err := s.userClient.FindById(util.TEL.Ctx(), callerID)
	if err != nil {
		util.TEL.Error("user does not exist", err, "id", callerID)
		return nil, ErrUnauthenticated
	}

	util.TEL.Debug("check if user is a guest", nil, "id", callerID)
	if user.Role != string(util.Guest) {
		util.TEL.Error("user has a bad role", nil, "role", user.Role)
		return nil, ErrUnauthorized
	}

	util.TEL.Debug("find room", "id", dto.RoomID)
	room, err := s.roomClient.FindById(util.TEL.Ctx(), dto.RoomID)
	if err != nil {
		util.TEL.Error("room not found", err, "id", dto.RoomID)
		return nil, ErrNotFound("room", dto.RoomID)
	}

	util.TEL.Debug("find room availability list", "room_id", room.ID)
	availList, err := s.roomClient.FindCurrentAvailabilityListOfRoom(util.TEL.Ctx(), room.ID)
	if err != nil {
		util.TEL.Error("room availability list of room not found", err, "room_id", dto.RoomID)
		return nil, ErrNotFound("room availability list", dto.RoomID)
	}

	util.TEL.Debug("find room price list", "room_id", room.ID)
	pricelist, err := s.roomClient.FindCurrentPricelistOfRoom(util.TEL.Ctx(), room.ID)
	if err != nil {
		util.TEL.Error("room price list of room not found", err, "room_id", dto.RoomID)
		return nil, ErrNotFound("room price list", dto.RoomID)
	}

	util.TEL.Push(context, "query-for-reservation")
	defer util.TEL.Pop()

	util.TEL.Debug("query room for reservation data")
	queryDTO := roomclient.RoomReservationQueryDTO{
		RoomID:     room.ID,
		DateFrom:   dto.DateFrom,
		DateTo:     dto.DateTo,
		GuestCount: dto.GuestCount,
	}
	queryResponse, err := s.roomClient.QueryForReservation(util.TEL.Ctx(), jwt, queryDTO)
	if err != nil {
		util.TEL.Error("could not query room for reservation", err, "room_id", dto.RoomID)
		return nil, ErrBadRequest
	}

	if !queryResponse.Available {
		util.TEL.Error("room is not available at this time", err)
		return nil, ErrBadRequest
	}

	util.TEL.Debug("calculate price")
	cost := queryResponse.TotalCost

	util.TEL.Push(context, "validate-reservation-request")
	defer util.TEL.Pop()

	util.TEL.Debug("validate fields")
	if dto.GuestCount < 1 {
		util.TEL.Error("guest count must be at least 1", err, "guest_count", dto.GuestCount)
		return nil, ErrBadRequestCustom("guest count must be at least 1")
	}

	if dto.DateFrom.After(dto.DateTo) {
		util.TEL.Error("dates are reversed", err, "from", dto.DateFrom, "to", dto.DateTo)
		return nil, ErrBadRequestCustom("dates are reversed")
	}

	util.TEL.Debug("prevent overlapping requests for the same room and same guest")
	existing, err := s.repo.FindPendingRequestsByGuestID(callerID)
	if err != nil {
		util.TEL.Error("could not find pending reservation requests of guest", err, "guest_id", callerID)
		return nil, err
	}
	for _, req := range existing {
		if req.RoomID == dto.RoomID {
			if util.AreDatesIntersecting(req.DateFrom, req.DateTo, dto.DateFrom, dto.DateTo) {
				util.TEL.Error("conflicting request of user for room", nil, "user_id", callerID, "room_id", dto.RoomID, "request_from", req.DateFrom, "request_to", req.DateTo, "existing_from", req.DateFrom, "existing_to", req.DateTo)
				return nil, ErrConflict
			}
		}
	}

	util.TEL.Debug("xheck if this room has a reservation for this date range", nil)

	has, err := s.AreThereReservationsOnDays(util.TEL.Ctx(), dto.RoomID, dto.DateFrom, dto.DateTo)
	if err != nil {
		util.TEL.Error("could not check for reservations for room", err, "room_id", dto.RoomID, "from", dto.DateFrom, "to", dto.DateTo)
		return nil, err
	}
	if has {
		util.TEL.Error("room has a reservation for this date range, cannot create a request", nil, "room_id", dto.RoomID, "from", dto.DateFrom, "to", dto.DateTo)
		return nil, ErrConflict
	}

	util.TEL.Push(context, "create-reservation-request-in-db")
	defer util.TEL.Pop()

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
		util.TEL.Error("failed creating a reservation request", err)
		return nil, err
	}
	return req, nil
}

func (s *service) FindPendingRequestsByGuest(context context.Context, callerID uint) ([]ReservationRequest, error) {
	util.TEL.Info("user wants to see his pending reservation requests", "caller_id", callerID)

	util.TEL.Debug("check if user exists", "id", callerID)
	user, err := s.userClient.FindById(util.TEL.Ctx(), callerID)
	if err != nil {
		util.TEL.Debug("user does not exist", err, "id", callerID)
		return nil, ErrNotFound("user", callerID)
	}

	util.TEL.Debug("check if user is a guest", "id", callerID)
	if user.Role != string(util.Guest) {
		util.TEL.Debug("user has a bad role", "role", user.Role)
		return nil, ErrUnauthorized
	}

	util.TEL.Push(context, "find-pending-reservation-requests-by-guest-in-db")
	defer util.TEL.Pop()
	return s.repo.FindPendingRequestsByGuestID(callerID)
}

func (s *service) FindPendingRequestsByRoom(context context.Context, callerID uint, roomID uint) ([]ReservationRequest, error) {
	util.TEL.Info("user wants to see pending reservation requests for room", "caller_id", callerID, "room_id", roomID)

	util.TEL.Debug("check if user exists", "id", callerID)
	user, err := s.userClient.FindById(util.TEL.Ctx(), callerID)
	if err != nil {
		util.TEL.Error("user does not exist", err, "id", callerID)
		return nil, ErrNotFound("user", callerID)
	}

	util.TEL.Debug("check if user is a host", "id", callerID)
	if user.Role != string(util.Host) {
		util.TEL.Error("user has a bad role", nil, "role", user.Role)
		return nil, ErrUnauthorized
	}

	util.TEL.Debug("find room", "id", roomID)
	room, err := s.roomClient.FindById(util.TEL.Ctx(), roomID)
	if err != nil {
		util.TEL.Error("room does not exist", err, "id", roomID)
		return nil, ErrNotFound("room", roomID)
	}

	util.TEL.Debug("check if user is owner of the room", "user_id", callerID, "room_id", room.ID)
	if room.HostID != callerID {
		util.TEL.Error("user is not owner of the room", nil, "user_id", callerID, "host_id", room.HostID)

		return nil, ErrUnauthorized
	}

	util.TEL.Push(context, "find-pending-reservation-requests-by-room-in-db")
	defer util.TEL.Pop()
	return s.repo.FindPendingRequestsByRoomID(roomID)
}

func (s *service) DeleteRequest(context context.Context, callerID uint, requestID uint) error {
	util.TEL.Info("user wants delete reservation request", "caller_id", callerID, "request_id", requestID)

	util.TEL.Debug("check if user exists", "id", callerID)
	user, err := s.userClient.FindById(util.TEL.Ctx(), callerID)
	if err != nil {
		util.TEL.Error("user does not exist", err, "id", callerID)
		return ErrNotFound("user", callerID)
	}

	util.TEL.Debug("check if user is a guest", "id", callerID)
	if user.Role != string(util.Guest) {
		util.TEL.Error("user has a bad role", nil, "role", user.Role)
		return ErrUnauthorized
	}

	util.TEL.Debug("find all reservation requests by user in db", "user_id", callerID)
	requests, err := s.repo.FindPendingRequestsByGuestID(callerID)
	if err != nil {
		util.TEL.Error("could not find reservation requests of user", err, "user_id", callerID)
		return err
	}
	found := false
	util.TEL.Debug("find reservation request by id", nil, "request_id", requestID)
	var request *ReservationRequest
	for _, req := range requests {
		if req.ID == requestID {
			found = true
			request = &req
			break
		}
	}

	if !found {
		util.TEL.Error("could not find reservation request of user", err, "request_id", requestID, "user_id", callerID)
		return ErrNotFound("reservation request", requestID)
	}

	if request.Status != Pending {
		util.TEL.Error("request isn't pending", nil, "request_status", request.Status)
		return ErrBadRequestCustom("cannot cancel a handled request")
	}

	util.TEL.Push(context, "delete-request-in-db")
	defer util.TEL.Pop()

	return s.repo.DeleteRequest(requestID)
}

func (s *service) AreThereReservationsOnDays(context context.Context, roomID uint, from, to time.Time) (bool, error) {
	util.TEL.Info("checking if room has reservations on days", "room_id", roomID, "from", from, "to", to)

	for d := from; !d.After(to); d = d.AddDate(0, 0, 1) {
		util.TEL.Debug("checking if room has a reservation on a single day ", "room_id", roomID, "day", d)

		reservations, err := s.repo.FindReservationsByRoomIDForDay(roomID, d)
		if err != nil {
			util.TEL.Error("could not find reservations for room on a single day", err, "room_id", roomID, "day", d)
			return false, err
		}
		if len(reservations) > 0 {
			util.TEL.Debug("reservation for room found on day", "room_id", roomID, "day", d)
			return true, nil
		}
	}

	util.TEL.Debug("ok, room has no reservations on days", "room_id", roomID, "from", from, "to", to)
	return false, nil
}

func (s *service) GetGuestActiveReservations(guestID uint) ([]Reservation, error) {
	log.Print("GetGuestActiveReservations [1] User must be guest")

	_, err := s.userClient.FindById(guestID)
	if err != nil {
		return nil, ErrNotFound("user", guestID)
	}

	log.Print("GetGuestActiveReservations [2] Return")

	allReservations, err := s.repo.FindReservationsByGuestID(guestID)
	if err != nil {
		return nil, err
	}

	var activeReservations []Reservation
	for _, reservation := range allReservations {
		if reservation.GuestID == guestID && reservation.DateTo.After(time.Now()) && reservation.Cancelled == false {
			activeReservations = append(activeReservations, reservation)
		}
	}

	return activeReservations, nil
}
