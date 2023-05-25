package models

import (
	"errors"
	"fmt"
	"time"
)

type Booking struct {
	ID          int64 `gorm:"primaryKey" json:"id"`
	UserID      int64
	SeatID      int64
	Duration    int8 // max 4h, unit hour
	CreatedAt   time.Time
	UpdatedAt   time.Time
	BookingTime time.Time // booking time, after it 15 min will auto cancel booking and free seat and notify student and record one default
	IsSigned    int8      // 0 represent waiting, 1 represent attend, 2 represent delay
}

func (b *Booking) Create() error {
	if ok := db.NewRecord(b); ok == true {
		return errors.New("booking already exists")
	}
	seat, err := GetSeatByID(b.SeatID)
	if err != nil {
		return err
	}
	// seat is not free
	if seat.Status != 1 {
		return errors.New("seat is not free")
	}
	room, err := GetRoomById(seat.RoomID)
	if err != nil {
		return err
	}
	room.Free -= 1
	if err = UpdateRoom(room); err != nil {
		return err
	}
	seat.Status = 2
	if err = UpdateSeat(seat); err != nil {
		return err
	}

	return db.Create(b).Error
}

func GetBookingByID(id int64) (*Booking, error) {
	var booking Booking
	if err := db.Model(&Booking{}).Where("id =?", id).First(&booking).Error; err != nil {
		return nil, err
	}
	return &booking, nil
}

func GetBookingByUserID(id int64) ([]Booking, error) {
	var bookings []Booking
	if err := db.Model(&Booking{}).Where("user_id =?", id).Find(&bookings).Error; err != nil {
		return nil, err
	}
	return bookings, nil
}

func DeleteBooking(id int64) error {
	var booking Booking
	if err := db.Model(&Booking{}).First(&booking).Error; err != nil {
		return err
	}
	seat, err := GetSeatByID(booking.SeatID)
	if err != nil {
		return err
	}
	// change seat status and room information
	seat.Status = 1
	room, err := GetRoomById(seat.RoomID)
	if err != nil {
		return err
	}
	room.Free += 1
	fmt.Println(room.Free)
	if err = UpdateRoom(room); err != nil {
		return err
	}
	if err = UpdateSeat(seat); err != nil {
		return err
	}
	// delete booking record
	return db.Model(&Booking{}).Where("id = ?", id).Delete(&Booking{}).Error
}

func UpdateBooking(id int64, booking *Booking) error {
	if booking.IsSigned == 1 { // user is using this seat, then change the seat status
		seat, err := GetSeatByID(booking.SeatID)
		if err != nil {
			return err
		}
		seat.Status = 3
		if err = UpdateSeat(seat); err != nil {
			return err
		}
	}
	return db.Model(&Booking{}).Where("id =?", id).Updates(*booking).Error
}
