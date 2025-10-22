package internal

import (
	"bookem-reservation-service/util"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type Route struct{ handler Handler }

func NewRoute(handler Handler) *Route { return &Route{handler} }

func (r *Route) Route(rg *gin.RouterGroup) {
	rg.POST("/req", r.handler.createReservationRequest)
	rg.GET("/req/user", r.handler.findPendingRequestsByGuest)
	rg.GET("/req/room/:id", r.handler.findPendingRequestsByRoom)
	rg.DELETE("/req/:id", r.handler.deleteRequestByGuest)

	rg.GET("/room/:id/availability", r.handler.checkAvailability)

	rg.GET("/reservations/guest/active", r.handler.getActiveGuestReservations)
	rg.GET("/reservations/host/active", r.handler.getActiveHostReservations)

	rg.PUT("/req/:id/reject", r.handler.rejectReservationRequest)
	rg.PUT("/req/:id/approve", r.handler.approveReservationRequest)
}

type Handler struct{ service Service }

func NewHandler(s Service) Handler { return Handler{s} }

func (h *Handler) createReservationRequest(ctx *gin.Context) {
	util.TEL.Push(ctx.Request.Context(), "create-reservation-request-api")
	defer util.TEL.Pop()

	jwtString, err := util.GetJwtString(ctx)
	if err != nil {
		util.TEL.Error("failed fetching JWT", err)
		AbortError(ctx, ErrUnauthenticated)
		return
	}

	jwt, err := util.GetJwt(ctx)
	if err != nil {
		util.TEL.Error("failed fetching JWT", err)
		AbortError(ctx, ErrUnauthenticated)
		return
	}

	if jwt.Role != util.Guest {
		util.TEL.Error("user is not guest", nil, "role", jwt.Role)
		AbortError(ctx, ErrUnauthorized)
		return
	}

	var dto CreateReservationRequestDTO
	if err := ctx.ShouldBindJSON(&dto); err != nil {
		util.TEL.Error("failed binding JSON", err)
		AbortError(ctx, err)
		return
	}

	reservation, err := h.service.CreateRequest(util.TEL.Ctx(), AuthContext{CallerID: jwt.ID, JWT: jwtString}, dto)
	if err != nil {
		util.TEL.Error("failed creating reservation request", err)
		AbortError(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, reservation)
}

func (h *Handler) findPendingRequestsByGuest(ctx *gin.Context) {
	util.TEL.Push(ctx.Request.Context(), "find-pending-requests-by-guest-api")
	defer util.TEL.Pop()

	jwt, err := util.GetJwt(ctx)
	if err != nil {
		util.TEL.Error("failed fetching JWT", err)
		AbortError(ctx, ErrUnauthenticated)
		return
	}

	if jwt.Role != util.Guest {
		util.TEL.Error("user is not guest", nil, "role", jwt.Role)
		AbortError(ctx, ErrUnauthorized)
		return
	}

	requests, err := h.service.FindPendingRequestsByGuest(util.TEL.Ctx(), jwt.ID)
	if err != nil {
		util.TEL.Error("failed finding pending requests by guest", err)
		AbortError(ctx, err)
		return
	}

	util.TEL.Debug("building response")

	result := make([]ReservationRequestDTO, 0)
	for _, req := range requests {
		result = append(result, NewReservationRequestDTO(req))
	}

	ctx.JSON(http.StatusOK, result)
}

func (h *Handler) findPendingRequestsByRoom(ctx *gin.Context) {
	util.TEL.Push(ctx.Request.Context(), "find-pending-requests-by-room-api")
	defer util.TEL.Pop()

	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		util.TEL.Error("could not parse request param id into a number", err, "id", ctx.Param("id"))
		AbortError(ctx, ErrBadRequest)
		return
	}

	jwt, err := util.GetJwt(ctx)
	if err != nil {
		util.TEL.Error("failed fetching JWT", err)
		AbortError(ctx, ErrUnauthenticated)
		return
	}

	if jwt.Role != util.Host {
		util.TEL.Error("user is not host", nil, "role", jwt.Role)
		AbortError(ctx, ErrUnauthorized)
		return
	}

	requests, err := h.service.FindPendingRequestsByRoom(util.TEL.Ctx(), jwt.ID, uint(id))
	if err != nil {
		util.TEL.Error("failed finding pending requests by room", err)
		AbortError(ctx, err)
		return
	}

	util.TEL.Debug("building response")

	result := make([]ReservationRequestDTO, 0)
	for _, req := range requests {
		result = append(result, NewReservationRequestDTO(req))
	}

	ctx.JSON(http.StatusOK, result)
}

func (h *Handler) deleteRequestByGuest(ctx *gin.Context) {
	util.TEL.Push(ctx.Request.Context(), "delete-request-by-guest-api")
	defer util.TEL.Pop()

	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		util.TEL.Error("could not parse request param id into number", err, "id", ctx.Param("id"))
		AbortError(ctx, ErrBadRequest)
		return
	}

	jwt, err := util.GetJwt(ctx)
	if err != nil {
		util.TEL.Error("failed fetching JWT", err)
		AbortError(ctx, ErrUnauthenticated)
		return
	}

	if jwt.Role != util.Guest {
		util.TEL.Error("user is not host", nil, "role", jwt.Role)
		AbortError(ctx, ErrUnauthorized)
		return
	}

	err = h.service.DeleteRequest(util.TEL.Ctx(), jwt.ID, uint(id))
	if err != nil {
		util.TEL.Error("failed deleting request by guest", err)
		AbortError(ctx, err)
		return
	}

	ctx.JSON(http.StatusNoContent, nil)
}

func (h *Handler) checkAvailability(ctx *gin.Context) {
	util.TEL.Push(ctx.Request.Context(), "check-availability-api")
	defer util.TEL.Pop()

	roomID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		util.TEL.Error("could not parse request param id into number", err, "id", ctx.Param("id"))
		AbortError(ctx, ErrBadRequest)
		return
	}

	fromStr := ctx.Query("from")
	toStr := ctx.Query("to")

	from, err := time.Parse("2006-01-02", fromStr)
	if err != nil {
		util.TEL.Error("invalid 'from' date format (should be YYYY-MM-DD)", err, "date", from)
		AbortError(ctx, ErrBadRequestCustom("invalid 'from' date format"))
		return
	}

	to, err := time.Parse("2006-01-02", toStr)
	if err != nil {
		util.TEL.Error("invalid 'to' date format (should be YYYY-MM-DD)", err, "date", from)
		AbortError(ctx, ErrBadRequestCustom("invalid 'to' date format"))
		return
	}

	available, err := h.service.AreThereReservationsOnDays(util.TEL.Ctx(), uint(roomID), from, to)
	if err != nil {
		util.TEL.Error("failed check if room has reservation in a date range", err)
		AbortError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"available": !available})
}

func (h *Handler) getActiveGuestReservations(ctx *gin.Context) {
	util.TEL.Push(ctx, "get-active-reservations-for-guest")
	defer util.TEL.Pop()

	jwt, err := util.GetJwt(ctx)
	if err != nil {
		util.TEL.Error("failed fetching JWT", err)
		AbortError(ctx, ErrUnauthenticated)
		return
	}

	if jwt.Role != util.Guest {
		util.TEL.Error("user is not host", nil, "role", jwt.Role)
		AbortError(ctx, ErrUnauthorized)
		return
	}

	reservations, err := h.service.GetActiveGuestReservations(ctx, jwt.ID)
	if err != nil {
		util.TEL.Error("could not get active guest reservations", err)
		AbortError(ctx, err)
		return
	}

	result := make([]ReservationDTO, 0)
	for _, res := range reservations {
		result = append(result, NewReservationDTO(res))
	}

	ctx.JSON(http.StatusOK, result)
}

func (h *Handler) getActiveHostReservations(ctx *gin.Context) {
	util.TEL.Push(ctx.Request.Context(), "get-active-host-reservations-api")
	defer util.TEL.Pop()

	jwt, err := util.GetJwt(ctx)
	if err != nil {
		util.TEL.Error("failed fetching JWT", err)
		AbortError(ctx, ErrUnauthenticated)
		return
	}

	if jwt.Role != util.Host {
		util.TEL.Error("user is not host", nil, "role", jwt.Role)
		AbortError(ctx, ErrUnauthorized)
		return
	}

	reservations, err := h.service.GetActiveHostReservations(ctx, jwt.ID)
	if err != nil {
		util.TEL.Error("could not get active host reservations", err)
		AbortError(ctx, err)
		return
	}

	result := make([]ReservationDTO, 0)
	for _, res := range reservations {
		result = append(result, NewReservationDTO(res))
	}

	ctx.JSON(http.StatusOK, result)
}

func (h *Handler) rejectReservationRequest(ctx *gin.Context) {
	util.TEL.Push(ctx, "reject-reservation-request")
	defer util.TEL.Pop()

	jwt, err := util.GetJwt(ctx)
	if err != nil {
		util.TEL.Error("failed fetching JWT", err)
		AbortError(ctx, ErrUnauthenticated)
		return
	}

	if jwt.Role != util.Host {
		util.TEL.Error("user is not host", nil, "role", jwt.Role)
		AbortError(ctx, ErrUnauthorized)
		return
	}

	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		util.TEL.Error("could not parse request param id into a number", err, "id", ctx.Param("id"))
		AbortError(ctx, ErrBadRequest)
		return
	}

	err = h.service.RejectReservationRequest(util.TEL.Ctx(), jwt.ID, uint(id))
	if err != nil {
		util.TEL.Error("could not reject reservation request", err)
		AbortError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "reservation request rejected successfully"})
}

func (h *Handler) approveReservationRequest(ctx *gin.Context) {
	util.TEL.Push(ctx, "approve-reservation-request")
	defer util.TEL.Pop()

	jwt, err := util.GetJwt(ctx)
	if err != nil {
		util.TEL.Error("failed fetching JWT", err)
		AbortError(ctx, ErrUnauthenticated)
		return
	}

	if jwt.Role != util.Host {
		util.TEL.Error("user is not host", nil, "role", jwt.Role)
		AbortError(ctx, ErrUnauthorized)
		return
	}

	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		util.TEL.Error("could not parse request param id into a number", err, "id", ctx.Param("id"))
		AbortError(ctx, ErrBadRequest)
		return
	}

	err = h.service.ApproveReservationRequest(util.TEL.Ctx(), jwt.ID, uint(id))
	if err != nil {
		util.TEL.Error("could not accept reservation request", err)
		AbortError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "reservation request approved successfully"})
}
