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
}

type Handler struct{ service Service }

func NewHandler(s Service) Handler { return Handler{s} }

func (h *Handler) createReservationRequest(ctx *gin.Context) {
	util.TEL.Push(ctx.Request.Context(), "create-reservation-request-api")
	defer util.TEL.Pop()

	jwtString, err := util.GetJwtString(ctx)
	if err != nil {
		util.TEL.Event("failed fetching JWT", err)
		AbortError(ctx, ErrUnauthenticated)
		return
	}

	jwt, err := util.GetJwt(ctx)
	if err != nil {
		util.TEL.Event("failed fetching JWT", err)
		AbortError(ctx, ErrUnauthenticated)
		return
	}

	if jwt.Role != util.Guest {
		util.TEL.Eventf("user is not guest (role=%s)", nil, jwt.Role)
		AbortError(ctx, ErrUnauthorized)
		return
	}

	var dto CreateReservationRequestDTO
	if err := ctx.ShouldBindJSON(&dto); err != nil {
		util.TEL.Event("failed binding JSON", err)
		AbortError(ctx, err)
		return
	}

	reservation, err := h.service.CreateRequest(util.TEL.Ctx(), AuthContext{CallerID: jwt.ID, JWT: jwtString}, dto)
	if err != nil {
		util.TEL.Event("failed creating reservation request", err)
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
		util.TEL.Event("failed fetching JWT", err)
		AbortError(ctx, ErrUnauthenticated)
		return
	}

	if jwt.Role != util.Guest {
		util.TEL.Eventf("user is not guest (role=%s)", nil, jwt.Role)
		AbortError(ctx, ErrUnauthorized)
		return
	}

	requests, err := h.service.FindPendingRequestsByGuest(util.TEL.Ctx(), jwt.ID)
	if err != nil {
		util.TEL.Event("failed finding pending requests by guest", err)
		AbortError(ctx, err)
		return
	}

	util.TEL.Event("building response", err)

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
		util.TEL.Eventf("could not parse request param id %s", err, ctx.Param("id"))
		AbortError(ctx, ErrBadRequest)
		return
	}

	jwt, err := util.GetJwt(ctx)
	if err != nil {
		util.TEL.Event("failed fetching JWT", err)
		AbortError(ctx, ErrUnauthenticated)
		return
	}

	if jwt.Role != util.Host {
		util.TEL.Eventf("user is not host (role=%s)", nil, jwt.Role)
		AbortError(ctx, ErrUnauthorized)
		return
	}

	requests, err := h.service.FindPendingRequestsByRoom(util.TEL.Ctx(), jwt.ID, uint(id))
	if err != nil {
		util.TEL.Event("failed finding pending requests by room", err)
		AbortError(ctx, err)
		return
	}

	util.TEL.Event("building response", err)

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
		util.TEL.Eventf("could not parse request param id %s", err, ctx.Param("id"))
		AbortError(ctx, ErrBadRequest)
		return
	}

	jwt, err := util.GetJwt(ctx)
	if err != nil {
		util.TEL.Event("failed fetching JWT", err)
		AbortError(ctx, ErrUnauthenticated)
		return
	}

	if jwt.Role != util.Guest {
		util.TEL.Eventf("user is not host (role=%s)", nil, jwt.Role)
		AbortError(ctx, ErrUnauthorized)
		return
	}

	err = h.service.DeleteRequest(util.TEL.Ctx(), jwt.ID, uint(id))
	if err != nil {
		util.TEL.Event("failed deleting request by guest", err)
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
		util.TEL.Eventf("could not parse request param id %s", err, ctx.Param("id"))
		AbortError(ctx, ErrBadRequest)
		return
	}

	fromStr := ctx.Query("from")
	toStr := ctx.Query("to")

	from, err := time.Parse("2006-01-02", fromStr)
	if err != nil {
		util.TEL.Eventf("invalid 'from' date format (should be YYYY-MM-DD, got %s)", err, from)
		AbortError(ctx, ErrBadRequestCustom("invalid 'from' date format"))
		return
	}

	to, err := time.Parse("2006-01-02", toStr)
	if err != nil {
		util.TEL.Eventf("invalid 'to' date format (should be YYYY-MM-DD, got %s)", err, from)
		AbortError(ctx, ErrBadRequestCustom("invalid 'to' date format"))
		return
	}

	available, err := h.service.AreThereReservationsOnDays(util.TEL.Ctx(), uint(roomID), from, to)
	if err != nil {
		util.TEL.Event("failed check if room has reservation in a date range", err)
		AbortError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"available": !available})
}
