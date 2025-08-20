package internal

import "gorm.io/gorm"

type Repository interface {
	Create(room *Reservation) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db}
}

func (r *repository) Create(room *Reservation) error {
	return r.db.Create(room).Error
}
