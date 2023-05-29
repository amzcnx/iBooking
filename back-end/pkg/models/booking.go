package models

import (
	"errors"
	"time"
)

type Booking struct {
	ID          int64 `gorm:"primaryKey" json:"id"`
	UserID      int64
	SeatID      int64
	RoomID      int64
	Duration    int8 // max 4h, unit hour
	CreatedAt   time.Time
	UpdatedAt   time.Time
	BookingTime time.Time // booking time, after it 15 min will auto cancel booking and free seat and notify student and record one default
	IsSigned    int8      // 0 represent waiting, 1 represent attend, 2 represent delay
}

func (b *Booking) Create() error {
	if db.NewRecord(b) {
		return errors.New("booking already exists")
	}
	room, err := GetRoomById(b.RoomID)
	if err != nil {
		return err
	}
	find := false
	for i := 0; i < len(room.Seats); i++ {
		if room.Seats[i].ID == b.SeatID {
			find = true
			if room.Seats[i].Status != 1 {
				return errors.New("seat is not free")
			}
			room.Seats[i].Status = 2
			break
		}
	}
	if !find {
		return errors.New("seat not found")
	}
	room.Free -= 1
	if err = UpdateRoom(room); err != nil {
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

func GetBookingByUserID(id int64) (*Booking, error) {
	var booking Booking
	if err := db.Model(&Booking{}).Where("user_id =?", id).Find(&booking).Error; err != nil {
		return nil, err
	}
	return &booking, nil
}

func DeleteBooking(id int64) error {
	var booking Booking
	if err := db.Model(&Booking{}).First(&booking).Error; err != nil {
		return err
	}
	// change seat status and room information
	room, err := GetRoomById(booking.RoomID)
	if err != nil {
		return err
	}
	find := false
	for i := 0; i < len(room.Seats); i++ {
		if room.Seats[i].ID == booking.SeatID {
			find = true
			room.Seats[i].Status = 1
			break
		}
	}
	if !find {
		return errors.New("seat not found")
	}
	room.Free += 1
	if err = UpdateRoom(room); err != nil {
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
