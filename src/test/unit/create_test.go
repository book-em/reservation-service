package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	mock "github.com/stretchr/testify/mock"
)

func Test_Create_Stub(t *testing.T) {
	svc, mockRepo, _, _ := CreateTestRoomService()

	dto := DefaultReservationDTO

	mockRepo.On("Create", mock.AnythingOfType("*internal.Reservation")).Return(nil)

	roomGot, err := svc.Create(DefaultUser_Host.Id, dto)

	assert.Error(t, err)
	assert.Nil(t, roomGot)
}
