package test

import (
	"bookem-reservation-service/client/roomclient"
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_CanUserRateHost_True(t *testing.T) {
	svc, repo, _, roomCli, _ := CreateTestRoomService()

	guestID := uint(10)
	hostID := uint(20)

	hostRooms := []roomclient.RoomDTO{
		{ID: 101, HostID: hostID},
		{ID: 102, HostID: hostID},
	}
	roomCli.On("FindByHostId", mock.Anything, hostID).Return(hostRooms, nil)

	repo.On("HasGuestPastReservationInRooms", guestID, []uint{101, 102}, mock.AnythingOfType("time.Time")).
		Return(true, nil)

	ok, err := svc.CanUserRateHost(context.Background(), guestID, hostID)

	assert.NoError(t, err)
	assert.True(t, ok)
	roomCli.AssertCalled(t, "FindByHostId", mock.Anything, hostID)
	repo.AssertCalled(t, "HasGuestPastReservationInRooms", guestID, []uint{101, 102}, mock.AnythingOfType("time.Time"))
}

func Test_CanUserRateHost_False_NoPastStay(t *testing.T) {
	svc, repo, _, roomCli, _ := CreateTestRoomService()

	guestID := uint(11)
	hostID := uint(21)

	roomCli.On("FindByHostId", mock.Anything, hostID).
		Return([]roomclient.RoomDTO{{ID: 201, HostID: hostID}}, nil)

	repo.On("HasGuestPastReservationInRooms", guestID, []uint{201}, mock.AnythingOfType("time.Time")).
		Return(false, nil)

	ok, err := svc.CanUserRateHost(context.Background(), guestID, hostID)

	assert.NoError(t, err)
	assert.False(t, ok)
}

func Test_CanUserRateHost_False_HostHasNoRooms(t *testing.T) {
	svc, _, _, roomCli, _ := CreateTestRoomService()

	guestID := uint(12)
	hostID := uint(22)

	roomCli.On("FindByHostId", mock.Anything, hostID).
		Return([]roomclient.RoomDTO{}, nil)

	ok, err := svc.CanUserRateHost(context.Background(), guestID, hostID)

	assert.NoError(t, err)
	assert.False(t, ok)
}

func Test_CanUserRateHost_RoomClientError(t *testing.T) {
	svc, _, _, roomCli, _ := CreateTestRoomService()

	guestID := uint(13)
	hostID := uint(23)

	roomCli.On("FindByHostId", mock.Anything, hostID).
		Return(nil, errors.New("room service down"))

	ok, err := svc.CanUserRateHost(context.Background(), guestID, hostID)

	assert.Error(t, err)
	assert.False(t, ok)
}

func Test_CanUserRateHost_RepoError(t *testing.T) {
	svc, repo, _, roomCli, _ := CreateTestRoomService()

	guestID := uint(14)
	hostID := uint(24)

	roomCli.On("FindByHostId", mock.Anything, hostID).
		Return([]roomclient.RoomDTO{{ID: 301, HostID: hostID}}, nil)

	repo.On("HasGuestPastReservationInRooms", guestID, []uint{301}, mock.AnythingOfType("time.Time")).
		Return(false, errors.New("db error"))

	ok, err := svc.CanUserRateHost(context.Background(), guestID, hostID)

	assert.Error(t, err)
	assert.False(t, ok)
}

func Test_CanUserRateRoom_True(t *testing.T) {
	svc, repo, _, _, _ := CreateTestRoomService()

	guestID := uint(15)
	roomID := uint(401)

	repo.On("HasGuestPastReservationInRooms", guestID, []uint{roomID}, mock.AnythingOfType("time.Time")).
		Return(true, nil)

	ok, err := svc.CanUserRateRoom(context.Background(), guestID, roomID)

	assert.NoError(t, err)
	assert.True(t, ok)
	repo.AssertCalled(t, "HasGuestPastReservationInRooms", guestID, []uint{roomID}, mock.AnythingOfType("time.Time"))
}

func Test_CanUserRateRoom_False(t *testing.T) {
	svc, repo, _, _, _ := CreateTestRoomService()

	guestID := uint(16)
	roomID := uint(402)

	repo.On("HasGuestPastReservationInRooms", guestID, []uint{roomID}, mock.AnythingOfType("time.Time")).
		Return(false, nil)

	ok, err := svc.CanUserRateRoom(context.Background(), guestID, roomID)

	assert.NoError(t, err)
	assert.False(t, ok)
}

func Test_CanUserRateRoom_RepoError(t *testing.T) {
	svc, repo, _, _, _ := CreateTestRoomService()

	guestID := uint(17)
	roomID := uint(403)

	repo.On("HasGuestPastReservationInRooms", guestID, []uint{roomID}, mock.AnythingOfType("time.Time")).
		Return(false, errors.New("db error"))

	ok, err := svc.CanUserRateRoom(context.Background(), guestID, roomID)

	assert.Error(t, err)
	assert.False(t, ok)
}