package models

import (
	"errors"
	"time"
)

// TODO:索引添加

type BookingHistory struct {
	ID           int64 `gorm:"primary_key"`
	UserID       int64
	SeatID       int64
	RoomID       int64
	RoomLocation string
	RoomNumber   string
	BookingTime  time.Time
	CreatedAt    time.Time
}

func (b *BookingHistory) Create() error {
	if db.NewRecord(b) {
		return errors.New("booking history already exists")
	}
	if err := db.Create(b).Error; err != nil {
		return err
	}
	return DeleteBooking(b.ID)
}

func GetBookingHistoryByUserID(id int64) ([]BookingHistory, error) {
	var bks []BookingHistory
	err := db.Model(&BookingHistory{}).Where("user_id = ?", id).Find(&bks).Error
	return bks, err
}
