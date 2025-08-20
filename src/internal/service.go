package internal

import (
	"bookem-reservation-service/client/roomclient"
	"bookem-reservation-service/client/userclient"
)

type Service interface {
	Create(callerID uint, dto ReservationDTO) (*Reservation, error)
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

func (s *service) Create(callerID uint, dto ReservationDTO) (*Reservation, error) {
	return nil, ErrNotFound("user", callerID)
}
