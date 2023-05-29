package models

import (
	"github.com/lib/pq"
	"time"
)

type Seat struct {
	ID        int64         `gorm:"primaryKey" json:"id"`
	RoomID    int64         `json:"room_id"`
	X         int8          `json:"x"`
	Y         int8          `json:"y"`
	Status    int8          `json:"status"`        // 0 represent not usable, 1 represent free, 2 represent reserved, 3 represent occupied, 4 represent under repair
	S         pq.ByteaArray `gorm:"type:byte[][]"` // test status  3*24, 最多预约3天内座位
	Plug      bool          `json:"plug"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (s *Seat) Create() error {
	room, err := GetRoomById(s.RoomID)
	if err != nil {
		return err
	}
	room.Seats = append(room.Seats, *s)
	room.Total += 1
	room.Free += 1
	return UpdateRoom(room)
}

func GetAllSeats() ([]Seat, error) {
	var seats []Seat
	if err := db.Model(&Seat{}).Find(&seats).Error; err != nil {
		return nil, err
	}
	return seats, nil
}

func GetSeatByID(id int64) (*Seat, error) {
	var seat Seat
	if err := db.Model(&Seat{}).Where("id=?", id).First(&seat).Error; err != nil {
		return nil, err
	}
	return &seat, nil
}

func DeleteSeat(id int64) error {
	seat, err := GetSeatByID(id)
	if err != nil {
		return err
	}
	room, err := GetRoomById(seat.RoomID)
	if err != nil {
		return err
	}
	room.Total -= 1
	room.Free -= 1
	if err = UpdateRoom(room); err != nil {
		return err
	}
	return db.Model(&Seat{}).Where("id=?", id).Delete(&Seat{}).Error
}

func DeleteSeatByRoomID(id int64) error {
	return db.Model(&Seat{}).Where("room_id = ?", id).Delete(&Seat{}).Error
}

func UpdateSeat(s *Seat) error {
	return db.Model(&Seat{}).Where("id = ?", s.ID).Update(s).Error
}

func GetSeatByRoomID(id int64) ([]Seat, error) {
	var seats []Seat
	if err := db.Model(&Seat{}).Where("room_id = ?", id).Find(&seats).Error; err != nil {
		return nil, err
	}
	return seats, nil
}
