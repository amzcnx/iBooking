package models

import (
	"time"
)

type User struct {
	ID        int64  `gorm:"primaryKey" json:"id"`
	Username  string `gorm:"unique" form:"username" json:"username" binding:"required"`
	Password  string `form:"password" json:"password" binding:"required"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Create is used to Sign up
func (u *User) Create() error {
	db.NewRecord(u)
	return db.Create(u).Error
}

func GetAllUser() ([]User, error) {
	var users []User
	if err := db.Model(&users).Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func GetUserByID(id int64) (*User, error) {
	var user User
	if err := db.Model(&User{}).Where("id =?", id).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func GetUserByUsername(username string) (*User, error) {
	var user User
	if err := db.Model(&User{}).Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func DeleteUser(id int64) error {
	return db.Model(&User{}).Where("id = ?", id).Delete(&User{}).Error
}

func UpdateUser(u *User) error {
	return db.Model(&User{}).Where("id = ?", u.ID).Update(u).Error
}
