package internal

import (
	"time"

	"gorm.io/gorm"
)

type Repository interface {
	// ReservationRequest methods
	CreateRequest(req *ReservationRequest) error
	DeleteRequest(id uint) error
	FindRequestsByRoomIDUpcoming(roomID uint, now time.Time) ([]ReservationRequest, error)
	SetRequestStatus(id uint, status ReservationRequestStatus) error
	RejectPendingRequestsInRange(roomID uint, from, to time.Time) error
	FindPendingRequestsByRoomID(roomID uint) ([]ReservationRequest, error)
	FindPendingRequestsByGuestID(guestID uint) ([]ReservationRequest, error)
	FindRequestByID(id uint) (*ReservationRequest, error)

	// Reservation methods
	CreateReservation(res *Reservation) error
	CancelReservation(id uint) error
	FindCancelledReservationsByGuestID(guestID uint) ([]Reservation, error)
	FindReservationsByRoomIDForDay(roomID uint, day time.Time) ([]Reservation, error)
	FindReservationsByGuestID(guestID uint) ([]Reservation, error)
	CountGuestCancellations(guestID uint) (int64, error)
	FindReservationsByRoomID(roomID uint) ([]Reservation, error)
	FindReservationById(id uint) (*Reservation, error)
	HasGuestPastReservationInRooms(guestID uint, roomIDs []uint, now time.Time) (bool, error)

}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db}
}

func (r *repository) CreateRequest(req *ReservationRequest) error {
	return r.db.Create(req).Error
}

func (r *repository) DeleteRequest(id uint) error {
	return r.db.Delete(&ReservationRequest{}, id).Error
}

func (r *repository) FindRequestsByRoomIDUpcoming(roomID uint, now time.Time) ([]ReservationRequest, error) {
	var requests []ReservationRequest
	err := r.db.Where("room_id = ? AND date_from > ?", roomID, now).Find(&requests).Error
	return requests, err
}

func (r *repository) SetRequestStatus(id uint, status ReservationRequestStatus) error {
	return r.db.Model(&ReservationRequest{}).Where("id = ?", id).Update("status", status).Error
}

func (r *repository) RejectPendingRequestsInRange(roomID uint, from, to time.Time) error {
	return r.db.Model(&ReservationRequest{}).
		Where("room_id = ? AND status = ? AND date_to >= ? AND date_from <= ?", roomID, Pending, from, to).
		Update("status", Rejected).Error
}

func (r *repository) CreateReservation(res *Reservation) error {
	return r.db.Create(res).Error
}

func (r *repository) CancelReservation(id uint) error {
	return r.db.Model(&Reservation{}).Where("id = ?", id).Update("cancelled", true).Error
}

func (r *repository) FindCancelledReservationsByGuestID(guestID uint) ([]Reservation, error) {
	var reservations []Reservation
	err := r.db.Where("guest_id = ? AND cancelled = ?", guestID, true).Find(&reservations).Error
	return reservations, err
}

func (r *repository) FindReservationsByRoomIDForDay(roomID uint, day time.Time) ([]Reservation, error) {
	var reservations []Reservation
	err := r.db.Where("room_id = ? AND date_to >= ? AND date_from <= ?", roomID, day, day).Find(&reservations).Error
	return reservations, err
}

func (r *repository) FindPendingRequestsByRoomID(roomID uint) ([]ReservationRequest, error) {
	var requests []ReservationRequest
	err := r.db.Where("room_id = ? AND status = ?", roomID, Pending).Find(&requests).Error
	return requests, err
}

func (r *repository) CountGuestCancellations(guestID uint) (int64, error) {
	var count int64
	err := r.db.Model(&Reservation{}).Where("guest_id = ? AND cancelled = ?", guestID, true).Count(&count).Error
	return count, err
}

func (r *repository) FindPendingRequestsByGuestID(guestID uint) ([]ReservationRequest, error) {
	var requests []ReservationRequest
	err := r.db.Where("guest_id = ? AND status = ?", guestID, Pending).Find(&requests).Error
	return requests, err
}

func (r *repository) FindReservationsByGuestID(guestID uint) ([]Reservation, error) {
	var reservations []Reservation
	err := r.db.Where("guest_id = ?", guestID).Find(&reservations).Error
	return reservations, err
}

func (r *repository) FindReservationsByRoomID(roomID uint) ([]Reservation, error) {
	var reservations []Reservation
	err := r.db.Where("room_id = ?", roomID).Find(&reservations).Error
	return reservations, err
}

func (r *repository) FindRequestByID(id uint) (*ReservationRequest, error) {
	var req ReservationRequest
	err := r.db.First(&req, id).Error
	if err != nil {
		return nil, err
	}
	return &req, nil
}

func (r *repository) FindReservationById(id uint) (*Reservation, error) {
	var reservation Reservation
	err := r.db.Where("id = ?", id).First(&reservation).Error
	if err != nil {
		return nil, err
	}
	return &reservation, nil
}

func (r *repository) HasGuestPastReservationInRooms(guestID uint, roomIDs []uint, now time.Time) (bool, error) {
	if len(roomIDs) == 0 {
		return false, nil
	}
	var count int64
	err := r.db.Model(&Reservation{}).
		Where("guest_id = ? AND cancelled = ? AND date_to < ? AND room_id IN ?", guestID, false, now, roomIDs).
		Count(&count).Error
	return count > 0, err
}

