package internal

import (
	"bookem-reservation-service/client/userclient"
	"bookem-reservation-service/util"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Route struct{ handler Handler }

func NewRoute(handler Handler) *Route { return &Route{handler} }

func (r *Route) Route(rg *gin.RouterGroup) {
	rg.POST("/req", r.handler.createReservationRequest)
	rg.GET("/req/user", r.handler.findPendingRequestsByGuest)
	rg.GET("/req/room/:id", r.handler.findPendingRequestsByRoom)
	rg.DELETE("/req/:id", r.handler.deleteRequestByGuest)
}

type Handler struct{ service Service }

func NewHandler(s Service) Handler { return Handler{s} }

func (h *Handler) createReservationRequest(ctx *gin.Context) {
	jwt, err := util.GetJwt(ctx)
	if err != nil {
		AbortError(ctx, ErrUnauthenticated)
		return
	}

	if jwt.Role != userclient.Guest {
		AbortError(ctx, ErrUnauthorized)
		return
	}

	var dto CreateReservationRequestDTO
	if err := ctx.ShouldBindJSON(&dto); err != nil {
		AbortError(ctx, err)
		return
	}

	reservation, err := h.service.CreateRequest(jwt.ID, dto)
	if err != nil {
		AbortError(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, reservation)
}

func (h *Handler) findPendingRequestsByGuest(ctx *gin.Context) {
	log.Printf("findPendingRequestsByGuest called")

	jwt, err := util.GetJwt(ctx)
	if err != nil {
		AbortError(ctx, ErrUnauthenticated)
		return
	}

	if jwt.Role != userclient.Guest {
		AbortError(ctx, ErrUnauthorized)
		return
	}

	requests, err := h.service.FindPendingRequestsByGuest(jwt.ID)
	if err != nil {
		AbortError(ctx, err)
		return
	}

	result := make([]ReservationRequestDTO, 0)
	for _, req := range requests {
		result = append(result, NewReservationRequestDTO(req))
	}

	ctx.JSON(http.StatusOK, result)
}

func (h *Handler) findPendingRequestsByRoom(ctx *gin.Context) {
	log.Printf("findPendingRequestsByRoom called")

	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		log.Printf("Could not parse ID %s: %s", ctx.Param("id"), err.Error())
		AbortError(ctx, ErrBadRequest)
		return
	}

	jwt, err := util.GetJwt(ctx)
	if err != nil {
		AbortError(ctx, ErrUnauthenticated)
		return
	}

	if jwt.Role != userclient.Host {
		AbortError(ctx, ErrUnauthorized)
		return
	}

	requests, err := h.service.FindPendingRequestsByRoom(jwt.ID, uint(id))
	if err != nil {
		AbortError(ctx, err)
		return
	}

	result := make([]ReservationRequestDTO, 0)
	for _, req := range requests {
		result = append(result, NewReservationRequestDTO(req))
	}

	ctx.JSON(http.StatusOK, result)
}

func (h *Handler) deleteRequestByGuest(ctx *gin.Context) {
	log.Printf("deleteRequestByGuest called")

	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		log.Printf("Could not parse ID %s: %s", ctx.Param("id"), err.Error())
		AbortError(ctx, ErrBadRequest)
		return
	}

	jwt, err := util.GetJwt(ctx)
	if err != nil {
		AbortError(ctx, ErrUnauthenticated)
		return
	}

	if jwt.Role != userclient.Guest {
		AbortError(ctx, ErrUnauthorized)
		return
	}

	err = h.service.DeleteRequest(jwt.ID, uint(id))
	if err != nil {
		AbortError(ctx, err)
		return
	}

	ctx.JSON(http.StatusNoContent, nil)
}
