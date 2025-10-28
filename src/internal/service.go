package internal

import (
	"bookem-reservation-service/client/notificationclient"
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

	// GetActiveGuestReservations checks whether a guest has any active reservations
	// now or in the future.
	GetActiveGuestReservations(context context.Context, guestID uint) ([]Reservation, error)

	// GetActiveHostReservations checks whether a host has any active reservations
	// now or in the future.
	GetActiveHostReservations(context context.Context, hostID uint) ([]Reservation, error)

	// ExtractActiveReservations filters the provided list of reservations
	// and returns only the active ones.
	ExtractActiveReservations(reservations []Reservation) []Reservation

	// RejectReservationRequest changes the status of a pending reservation request
	// to rejected.
	RejectReservationRequest(context context.Context, hostID, requestID uint, jwt string) error

	// ApproveReservationRequest approves a reservation request for a room
	// and creates a corresponding reservation record.
	ApproveReservationRequest(context context.Context, hostID, requestID uint, jwt string) error

	CancelReservation(ctx context.Context, callerID uint, reservationID uint, jwt string) error

	GetGuestCancellationCount(context context.Context, guestID uint) (uint, error)

	CanUserRateHost(ctx context.Context, guestID, hostID uint) (bool, error)
	CanUserRateRoom(ctx context.Context, guestID, roomID uint) (bool, error)
}

type service struct {
	repo               Repository
	userClient         userclient.UserClient
	roomClient         roomclient.RoomClient
	notificationClient notificationclient.NotificationClient
}

func NewService(
	roomRepo Repository,
	userClient userclient.UserClient,
	roomClient roomclient.RoomClient,
	notificationClient notificationclient.NotificationClient,
) Service {
	return &service{roomRepo, userClient, roomClient, notificationClient}
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

	util.TEL.Debug("check if this room has a reservation for this date range", nil)

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

	if room.AutoApprove {
		util.TEL.Info("auto-approval is enabled, accepting reservation request automatically", "room_id", room.ID)
		if err := s.acceptReservationRequest(util.TEL.Ctx(), req, room, jwt); err != nil {
			util.TEL.Error("auto-approval process failed", err)
			return nil, err
		}
	}

	util.TEL.Info("reservation request created successfully", "request_id", req.ID)

	util.TEL.Push(context, "create-notification")
	defer util.TEL.Pop()

	createNotifDTO := notificationclient.CreateNotificationDTO{
		ReceiverID: room.HostID, // Host of the room receives the notification
		Type:       notificationclient.ReservationRequested,
		Subject:    callerID, // Guest ID (who made the request)
		Object:     room.ID,
	}

	if _, err := s.notificationClient.CreateNotification(util.TEL.Ctx(), jwt, createNotifDTO); err != nil {
		util.TEL.Error("failed to send notification to host", err, "host_id", room.HostID)
		// not returning error – notification failures shouldn’t break reservation creation
	}

	return req, nil
}

func (s *service) acceptReservationRequest(ctx context.Context, req *ReservationRequest, room *roomclient.RoomDTO, jwt string) error {
	util.TEL.Info("accept reservation request", "room_id", req.RoomID, "guest_id", req.GuestID)

	util.TEL.Debug("find current availability and price lists")
	availList, err := s.roomClient.FindCurrentAvailabilityListOfRoom(util.TEL.Ctx(), room.ID)
	if err != nil {
		util.TEL.Error("room availability list not found", err, "room_id", room.ID)
		return ErrNotFound("room availability list", room.ID)
	}

	pricelist, err := s.roomClient.FindCurrentPricelistOfRoom(util.TEL.Ctx(), room.ID)
	if err != nil {
		util.TEL.Error("room price list not found", err, "room_id", room.ID)
		return ErrNotFound("room price list", room.ID)
	}

	util.TEL.Debug("create reservation")
	res := &Reservation{
		RoomID:             req.RoomID,
		RoomAvailabilityID: availList.ID,
		RoomPriceID:        pricelist.ID,
		GuestID:            req.GuestID,
		DateFrom:           req.DateFrom,
		DateTo:             req.DateTo,
		GuestCount:         req.GuestCount,
		Cancelled:          false,
		Cost:               req.Cost,
	}
	if err := s.repo.CreateReservation(res); err != nil {
		util.TEL.Error("could not create reservation", err)
		return err
	}

	util.TEL.Debug("reject overlapping pending requests")
	if err := s.repo.RejectPendingRequestsInRange(req.RoomID, req.DateFrom, req.DateTo); err != nil {
		util.TEL.Warn("could not reject overlapping requests", "room_id", req.RoomID)
	}

	util.TEL.Debug("update current request to accepted")
	if err := s.repo.SetRequestStatus(req.ID, Accepted); err != nil {
		util.TEL.Error("failed updating request status to accepted", err)
		return err
	}
	req.Status = Accepted

	util.TEL.Info("reservation request accepted successfully", "request_id", req.ID)

	util.TEL.Push(ctx, "create-notification")
	defer util.TEL.Pop()

	createNotifDTO := notificationclient.CreateNotificationDTO{
		ReceiverID: req.GuestID, // Guest receives the notification
		Type:       notificationclient.ReservationAccepted,
		Subject:    room.HostID,
		Object:     room.ID,
	}

	if _, err := s.notificationClient.CreateNotification(util.TEL.Ctx(), jwt, createNotifDTO); err != nil {
		util.TEL.Error("failed to send notification to guest", err, "guest_id", req.GuestID)
		// not returning error – notification failures shouldn’t break reservation creation
	}

	return nil
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
		for _, res := range reservations {
			if !res.Cancelled {
				util.TEL.Debug("active reservation found", "room_id", roomID, "day", d)
				return true, nil
			}
		}
	}

	util.TEL.Debug("ok, room has no reservations on days", "room_id", roomID, "from", from, "to", to)
	return false, nil
}

func (s *service) GetActiveGuestReservations(context context.Context, guestID uint) ([]Reservation, error) {
	log.Print("GetActiveGuestReservations [1] User must be guest")

	_, err := s.userClient.FindById(context, guestID)
	if err != nil {
		return nil, ErrNotFound("user", guestID)
	}

	log.Print("GetActiveGuestReservations [2]  Find all reservations")

	allReservations, err := s.repo.FindReservationsByGuestID(guestID)
	if err != nil {
		return nil, err
	}

	log.Print("GetActiveGuestReservations [3] Return")

	activeReservations := s.ExtractActiveReservations(allReservations)

	return activeReservations, nil
}

func (s *service) GetActiveHostReservations(context context.Context, hostID uint) ([]Reservation, error) {

	log.Print("GetActiveHostReservations [1] Fetch host rooms")

	rooms, err := s.roomClient.FindByHostId(context, hostID)
	if err != nil {
		log.Printf("%s", err.Error())
		return nil, ErrNotFound("rooms of host", hostID)
	}

	log.Print("GetActiveHostReservations [2] Find all reservations from all rooms")

	var allReservations []Reservation
	for _, room := range rooms {
		roomReservations, err := s.repo.FindReservationsByRoomID(room.ID)
		if err != nil {
			return nil, err
		}
		allReservations = append(allReservations, roomReservations...)
	}

	log.Print("GetActiveHostReservations [3] Extract active reservations")

	activeReservations := s.ExtractActiveReservations(allReservations)

	log.Print("GetActiveHostReservations [4] Return")

	return activeReservations, nil
}

func (s *service) ExtractActiveReservations(reservations []Reservation) []Reservation {
	var activeReservations []Reservation
	for _, reservation := range reservations {
		if reservation.DateTo.After(time.Now()) && reservation.Cancelled == false {
			activeReservations = append(activeReservations, reservation)
		}
	}
	return activeReservations
}

func (s *service) RejectReservationRequest(ctx context.Context, hostID, requestID uint, jwt string) error {
	util.TEL.Push(ctx, "reject-reservation-request-service")
	defer util.TEL.Pop()

	req, err := s.repo.FindRequestByID(requestID)
	if err != nil {
		util.TEL.Error("could not find reservation requst", err, "request_id", requestID)
		return err
	}

	room, err := s.roomClient.FindById(util.TEL.Ctx(), req.RoomID)
	if err != nil {
		util.TEL.Error("room not found", err, "id", req.RoomID)
		return err
	}

	if room.HostID != hostID {
		util.TEL.Error("bad host for room", nil, "host_id", room.HostID, "room_id", room.ID)
		return ErrUnauthorized
	}

	if err := s.repo.SetRequestStatus(requestID, Rejected); err != nil {
		util.TEL.Error("could not change status to rejected", err, "request_id", requestID)
		return err
	}

	util.TEL.Info("reservation request rejected", "request_id", requestID)

	util.TEL.Push(ctx, "create-notification")
	defer util.TEL.Pop()

	createNotifDTO := notificationclient.CreateNotificationDTO{
		ReceiverID: req.GuestID, // Guest receives the notification
		Type:       notificationclient.ReservationDeclined,
		Subject:    room.HostID,
		Object:     room.ID,
	}

	if _, err := s.notificationClient.CreateNotification(util.TEL.Ctx(), jwt, createNotifDTO); err != nil {
		util.TEL.Error("failed to send notification to guest", err, "guest_id", req.GuestID)
		// not returning error – notification failures shouldn’t break reservation creation
	}

	return nil
}

func (s *service) ApproveReservationRequest(ctx context.Context, hostID, requestID uint, jwt string) error {
	util.TEL.Push(ctx, "approve-reservation-request-service")
	defer util.TEL.Pop()

	req, err := s.repo.FindRequestByID(requestID)
	if err != nil {
		util.TEL.Error("could not find reservation requst", err, "request_id", requestID)
		return err
	}

	room, err := s.roomClient.FindById(util.TEL.Ctx(), req.RoomID)
	if err != nil {
		util.TEL.Error("room not found", err, "id", req.RoomID)
		return err
	}

	if room.HostID != hostID {
		util.TEL.Error("bad host for room", nil, "host_id", room.HostID, "room_id", room.ID)
		return ErrUnauthorized
	}

	if err := s.acceptReservationRequest(util.TEL.Ctx(), req, room, jwt); err != nil {
		util.TEL.Error("could not change status to accepted", err, "request_id", requestID)
		return err
	}

	util.TEL.Info("reservation request approved successfully", "request_id", requestID)
	return nil
}

func (s *service) CancelReservation(ctx context.Context, callerID uint, reservationID uint, jwt string) error {
	util.TEL.Info("user wants to cancel reservation", "caller_id", callerID, "reservation_id", reservationID)

	user, err := s.userClient.FindById(util.TEL.Ctx(), callerID)
	if err != nil {
		util.TEL.Error("user not found", err, "user_id", callerID)
		return ErrUnauthenticated
	}

	if user.Role != string(util.Guest) {
		util.TEL.Error("user is not a guest", nil, "role", user.Role)
		return ErrUnauthorized
	}

	reservation, err := s.repo.FindReservationById(reservationID)
	if err != nil {
		util.TEL.Error("reservation not found", err, "reservation_id", reservationID)
		return ErrNotFound("reservation", reservationID)
	}

	if reservation.GuestID != callerID {
		util.TEL.Error("reservation does not belong to this guest", nil, "reservation_guest_id", reservation.GuestID, "caller_id", callerID)
		return ErrUnauthorized
	}

	if reservation.Cancelled {
		util.TEL.Error("reservation already cancelled", nil, "reservation_id", reservationID)
		return ErrBadRequestCustom("reservation already cancelled")
	}

	if !time.Now().Before(reservation.DateFrom) {
		util.TEL.Error("cannot cancel reservation that already started", nil, "date_from", reservation.DateFrom)
		return ErrBadRequestCustom("cannot cancel reservation that already started")
	}

	util.TEL.Push(ctx, "cancel-reservation-in-db")
	defer util.TEL.Pop()

	err = s.repo.CancelReservation(reservationID)
	if err != nil {
		util.TEL.Error("could not cancel reservation in database", err, "reservation_id", reservationID)
		return err
	}

	util.TEL.Info("reservation cancelled successfully", "reservation_id", reservationID)

	util.TEL.Push(ctx, "create-notification")
	defer util.TEL.Pop()

	room, err := s.roomClient.FindById(util.TEL.Ctx(), reservation.RoomID)
	if err != nil {
		util.TEL.Error("room not found", err, "id", reservation.RoomID)
		return err
	}

	createNotifDTO := notificationclient.CreateNotificationDTO{
		ReceiverID: room.HostID,
		Type:       notificationclient.ReservationCancelled,
		Subject:    reservation.GuestID,
		Object:     reservation.RoomID,
	}

	if _, err := s.notificationClient.CreateNotification(util.TEL.Ctx(), jwt, createNotifDTO); err != nil {
		util.TEL.Error("failed to send notification to host", err, "host_id", room.HostID)
		// not returning error – notification failures shouldn’t break reservation creation
	}

	return nil
}

func (s *service) GetGuestCancellationCount(ctx context.Context, guestID uint) (uint, error) {
	util.TEL.Push(ctx, "get-guest-cancellation-count")
	defer util.TEL.Pop()

	util.TEL.Info("count guest cancellations for user", "guest_id", guestID)

	count, err := s.repo.CountGuestCancellations(guestID)
	if err != nil {
		util.TEL.Error("could not count guest cancellations", err, "guest_id", guestID)
		return 0, err
	}

	util.TEL.Debug("guest cancellation count calculated", "guest_id", guestID, "count", count)
	return uint(count), nil
}

func (s *service) CanUserRateHost(ctx context.Context, guestID, hostID uint) (bool, error) {
	util.TEL.Push(ctx, "eligibility-can-user-rate-host")
	defer util.TEL.Pop()

	rooms, err := s.roomClient.FindByHostId(util.TEL.Ctx(), hostID)
	if err != nil {
		util.TEL.Error("failed to fetch rooms by host", err, "host_id", hostID)
		return false, err
	}
	if len(rooms) == 0 {
		util.TEL.Info("host has no rooms; guest cannot have stayed", "host_id", hostID)
		return false, nil
	}

	roomIDs := make([]uint, 0, len(rooms))
	for _, r := range rooms {
		roomIDs = append(roomIDs, r.ID)
	}

	ok, err := s.repo.HasGuestPastReservationInRooms(guestID, roomIDs, time.Now().UTC())
	if err != nil {
		util.TEL.Error("repo eligibility check failed", err)
		return false, err
	}
	return ok, nil
}

func (s *service) CanUserRateRoom(ctx context.Context, guestID, roomID uint) (bool, error) {
	util.TEL.Push(ctx, "eligibility-can-user-rate-room")
	defer util.TEL.Pop()

	ok, err := s.repo.HasGuestPastReservationInRooms(guestID, []uint{roomID}, time.Now().UTC())
	if err != nil {
		util.TEL.Error("repo eligibility check failed", err, "guest_id", guestID, "room_id", roomID)
		return false, err
	}
	return ok, nil
}

