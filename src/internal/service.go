package internal

import (
	"bookem-reservation-service/client/roomclient"
	"bookem-reservation-service/client/userclient"
	"bookem-reservation-service/util"
	"context"
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

	util.TEL.Eventf("user %d wants to create a reservation request", nil, authctx.CallerID)

	util.TEL.Push(context, "validate-room-and-user")
	defer util.TEL.Pop()

	util.TEL.Eventf("check if user %d exists", nil, callerID)
	user, err := s.userClient.FindById(util.TEL.Ctx(), callerID)
	if err != nil {
		util.TEL.Eventf("user %d does not exist", err, callerID)
		return nil, ErrUnauthenticated
	}

	util.TEL.Eventf("check if user %d is a guest", nil, callerID)
	if user.Role != string(util.Guest) {
		util.TEL.Eventf("user has a bad role (%s)", nil, user.Role)
		return nil, ErrUnauthorized
	}

	util.TEL.Eventf("find room", nil)
	room, err := s.roomClient.FindById(util.TEL.Ctx(), dto.RoomID)
	if err != nil {
		util.TEL.Eventf("room not found %d", err, dto.RoomID)
		return nil, ErrNotFound("room", dto.RoomID)
	}

	util.TEL.Eventf("find room availability list", nil)
	availList, err := s.roomClient.FindCurrentAvailabilityListOfRoom(util.TEL.Ctx(), room.ID)
	if err != nil {
		util.TEL.Eventf("room availability list of room %d not found", err, dto.RoomID)
		return nil, ErrNotFound("room availability list", dto.RoomID)
	}

	util.TEL.Eventf("find room price list", nil)
	pricelist, err := s.roomClient.FindCurrentPricelistOfRoom(util.TEL.Ctx(), room.ID)
	if err != nil {
		util.TEL.Eventf("room price list of room %d not found", err, dto.RoomID)
		return nil, ErrNotFound("room price list", dto.RoomID)
	}

	util.TEL.Push(context, "query-for-reservation")
	defer util.TEL.Pop()

	util.TEL.Eventf("query room for reservation data", nil)
	queryDTO := roomclient.RoomReservationQueryDTO{
		RoomID:     room.ID,
		DateFrom:   dto.DateFrom,
		DateTo:     dto.DateTo,
		GuestCount: dto.GuestCount,
	}
	queryResponse, err := s.roomClient.QueryForReservation(util.TEL.Ctx(), jwt, queryDTO)
	if err != nil {
		util.TEL.Eventf("could not query room %d for reservation", err, dto.RoomID)
		return nil, ErrBadRequest
	}

	if !queryResponse.Available {
		util.TEL.Eventf("room is not available at this time", err)
		return nil, ErrBadRequest
	}

	util.TEL.Eventf("calculate price", nil)
	cost := queryResponse.TotalCost

	util.TEL.Push(context, "validate-reservation-request")
	defer util.TEL.Pop()

	util.TEL.Eventf("validate fields", nil)
	if dto.GuestCount < 1 {
		util.TEL.Eventf("guest count must be at least 1 (got %d)", err, dto.GuestCount)
		return nil, ErrBadRequestCustom("guest count must be at least 1")
	}

	if dto.DateFrom.After(dto.DateTo) {
		util.TEL.Eventf("dates are reversed (got from %d to %d)", err, dto.DateFrom.String(), dto.DateTo.String())
		return nil, ErrBadRequestCustom("dates are reversed")
	}

	util.TEL.Eventf("prevent overlapping requests for the same room and same guest", nil)
	existing, err := s.repo.FindPendingRequestsByGuestID(callerID)
	if err != nil {
		util.TEL.Eventf("could not find pending reservation requests of guest %d", err, callerID)
		return nil, err
	}
	for _, req := range existing {
		if req.RoomID == dto.RoomID {
			if util.AreDatesIntersecting(req.DateFrom, req.DateTo, dto.DateFrom, dto.DateTo) {
				util.TEL.Eventf("user already has request for room at [%s - %s], but he now wants [%s - %s] (intersection)", nil, req.DateFrom.String(), req.DateTo.String(), dto.DateFrom.String(), dto.DateTo.String())
				return nil, ErrConflict
			}
		}
	}

	util.TEL.Eventf("xheck if this room has a reservation for this date range", nil)

	has, err := s.AreThereReservationsOnDays(util.TEL.Ctx(), dto.RoomID, dto.DateFrom, dto.DateTo)
	if err != nil {
		util.TEL.Eventf("could not check for reservations for room %d for date range [%s - %s]", err, dto.RoomID, dto.DateFrom.String(), dto.DateTo.String())
		return nil, err
	}
	if has {
		util.TEL.Eventf("room has a reservation for this date range, cannot create a request", nil)
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
		util.TEL.Eventf("failed creating a reservation request", err)
		return nil, err
	}
	return req, nil
}

func (s *service) FindPendingRequestsByGuest(context context.Context, callerID uint) ([]ReservationRequest, error) {
	util.TEL.Eventf("user %s wants to see his pending reservation requests", nil, callerID)

	util.TEL.Eventf("check if user %d exists", nil, callerID)
	user, err := s.userClient.FindById(util.TEL.Ctx(), callerID)
	if err != nil {
		util.TEL.Eventf("user %d does not exist", err, callerID)
		return nil, ErrNotFound("user", callerID)
	}

	util.TEL.Eventf("check if user %d is a guest", nil, callerID)
	if user.Role != string(util.Guest) {
		util.TEL.Eventf("user has a bad role (%s)", nil, user.Role)
		return nil, ErrUnauthorized
	}

	util.TEL.Push(context, "find-pending-reservation-requests-by-guest-in-db")
	defer util.TEL.Pop()
	return s.repo.FindPendingRequestsByGuestID(callerID)
}

func (s *service) FindPendingRequestsByRoom(context context.Context, callerID uint, roomID uint) ([]ReservationRequest, error) {
	util.TEL.Eventf("user %s wants to see pending reservation requests for room %d", nil, callerID, roomID)

	util.TEL.Eventf("check if user %d exists", nil, callerID)
	user, err := s.userClient.FindById(util.TEL.Ctx(), callerID)
	if err != nil {
		util.TEL.Eventf("user %d does not exist", err, callerID)
		return nil, ErrNotFound("user", callerID)
	}

	util.TEL.Eventf("check if user %d is a host", nil, callerID)
	if user.Role != string(util.Host) {
		util.TEL.Eventf("user has a bad role (%s)", nil, user.Role)
		return nil, ErrUnauthorized
	}

	util.TEL.Eventf("find room %d", nil, roomID)

	room, err := s.roomClient.FindById(util.TEL.Ctx(), roomID)
	if err != nil {
		util.TEL.Eventf("room %d does not exist", err, callerID)
		return nil, ErrNotFound("room", roomID)
	}

	util.TEL.Eventf("user must be owner of the room", nil)

	if room.HostID != callerID {
		util.TEL.Eventf("user is not owner of the room (want %d, owner is %d)", nil, callerID, room.HostID)

		return nil, ErrUnauthorized
	}

	util.TEL.Push(context, "find-pending-reservation-requests-by-room-in-db")
	defer util.TEL.Pop()
	return s.repo.FindPendingRequestsByRoomID(roomID)
}

func (s *service) DeleteRequest(context context.Context, callerID uint, requestID uint) error {
	util.TEL.Eventf("user %s wants delete reservation request %d", nil, callerID, requestID)

	util.TEL.Eventf("check if user %d exists", nil, callerID)
	user, err := s.userClient.FindById(util.TEL.Ctx(), callerID)
	if err != nil {
		util.TEL.Eventf("user %d does not exist", err, callerID)
		return ErrNotFound("user", callerID)
	}

	util.TEL.Eventf("check if user %d is a guest", nil, callerID)
	if user.Role != string(util.Guest) {
		util.TEL.Eventf("user has a bad role (%s)", nil, user.Role)
		return ErrUnauthorized
	}

	util.TEL.Eventf("find all reservation requests by user in db", nil)
	requests, err := s.repo.FindPendingRequestsByGuestID(callerID)
	if err != nil {
		util.TEL.Eventf("could not find reservation requests of user %d", err, callerID)
		return err
	}
	found := false
	util.TEL.Eventf("find reservation request by id", nil)
	var request *ReservationRequest
	for _, req := range requests {
		if req.ID == requestID {
			found = true
			request = &req
			break
		}
	}

	if !found {
		util.TEL.Eventf("could not find reservation request by id %d of user %d", err, requestID, callerID)
		return ErrNotFound("reservation request", requestID)
	}

	if request.Status != Pending {
		util.TEL.Eventf("request isn't pending (status is %s)", nil, request.Status)
		return ErrBadRequestCustom("cannot cancel a handled request")
	}

	util.TEL.Push(context, "delete-request-in-db")
	defer util.TEL.Pop()

	return s.repo.DeleteRequest(requestID)
}

func (s *service) AreThereReservationsOnDays(context context.Context, roomID uint, from, to time.Time) (bool, error) {
	util.TEL.Eventf("checking if room %d has reservations on days [%s - %s] ", nil, roomID, from.String(), to.String())

	for d := from; !d.After(to); d = d.AddDate(0, 0, 1) {
		util.TEL.Eventf("checking if room %d has a reservation on day %s ", nil, roomID, d.String())

		reservations, err := s.repo.FindReservationsByRoomIDForDay(roomID, d)
		if err != nil {
			util.TEL.Eventf("could not find reservations for room %d on day %s", err, roomID, d.String())
			return false, err
		}
		if len(reservations) > 0 {
			util.TEL.Eventf("reservation for room %d found on day %s", nil, roomID, d.String())
			return true, nil
		}
	}

	util.TEL.Eventf("ok, no room %d has no reservations on days [%s - %s] ", nil, roomID, from.String(), to.String())
	return false, nil
}
