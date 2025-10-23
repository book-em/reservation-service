package test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetGuestCancellationCount_Success(t *testing.T) {
	svc, mockRepo, _, _ := CreateTestRoomService()

	guestID := uint(1)

	mockRepo.On("CountGuestCancellations", guestID).Return(int64(3), nil)

	count, err := svc.GetGuestCancellationCount(context.Background(), guestID)

	assert.NoError(t, err)
	assert.Equal(t, uint(3), count)
	mockRepo.AssertCalled(t, "CountGuestCancellations", guestID)
}

func TestGetGuestCancellationCount_DBError(t *testing.T) {
	svc, mockRepo, _, _ := CreateTestRoomService()

	guestID := uint(2)

	mockRepo.On("CountGuestCancellations", guestID).Return(int64(0), errors.New("db error"))

	count, err := svc.GetGuestCancellationCount(context.Background(), guestID)

	assert.Error(t, err)
	assert.Equal(t, uint(0), count)
	assert.Contains(t, err.Error(), "db error")
}

func TestGetGuestCancellationCount_NoCancellations(t *testing.T) {
	svc, mockRepo, _, _ := CreateTestRoomService()

	guestID := uint(3)

	mockRepo.On("CountGuestCancellations", guestID).Return(int64(0), nil)

	count, err := svc.GetGuestCancellationCount(context.Background(), guestID)

	assert.NoError(t, err)
	assert.Equal(t, uint(0), count)
	mockRepo.AssertCalled(t, "CountGuestCancellations", guestID)
}
