package models

import (
	"time"
)

type Room struct {
	ID         int64  `gorm:"primaryKey" json:"id"`
	RoomNumber string `json:"room_number"`
	Location   string `json:"location"`
	OpenTime   int8   `json:"open_time"`  // unit hour, 0-24
	CloseTime  int8   `json:"close_time"` // unit hour, 0-24, not earlier than open_time
	Overnight  bool   `json:"overnight"`
	Seats      []Seat `json:"seats"`
	Total      int16  `json:"total"`
	Free       int16  `json:"free"`
	Usable     bool   `json:"usable"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (r *Room) Create() error {
	db.NewRecord(r)
	return db.Create(r).Error
}

func GetAllRooms() ([]Room, error) {
	var rooms []Room
	if err := db.Model(&Room{}).Preload("Seats").Find(&rooms).Error; err != nil {
		return nil, err
	}
	return rooms, nil
}

func GetRoomById(id int64) (*Room, error) {
	var room Room
	if err := db.Model(&Room{}).Where("ID=?", id).Preload("Seats").Find(&room).Error; err != nil {
		return nil, err
	}
	return &room, nil
}

func DeleteRoom(id int64) error {
	return db.Model(&Room{}).Where("ID=?", id).Delete(&Room{}).Error
}

func UpdateRoom(r *Room) error {
	return db.Model(&Room{}).Where("ID = ?", r.ID).Update(r).Error
}
