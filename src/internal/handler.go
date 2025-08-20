package internal

import (
	"bookem-reservation-service/client/userclient"
	"bookem-reservation-service/util"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Route struct{ handler Handler }

func NewRoute(handler Handler) *Route { return &Route{handler} }

func (r *Route) Route(rg *gin.RouterGroup) {
	rg.POST("/new", r.handler.createReservation)
}

type Handler struct{ service Service }

func NewHandler(s Service) Handler { return Handler{s} }

func (h *Handler) createReservation(ctx *gin.Context) {
	jwt, err := util.GetJwt(ctx)
	if err != nil {
		AbortError(ctx, ErrUnauthenticated)
		return
	}

	if jwt.Role != userclient.Guest {
		AbortError(ctx, ErrUnauthorized)
		return
	}

	var dto ReservationDTO
	if err := ctx.ShouldBindJSON(&dto); err != nil {
		AbortError(ctx, err)
		return
	}

	reservation, err := h.service.Create(jwt.ID, dto)
	if err != nil {
		AbortError(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, reservation)
}
