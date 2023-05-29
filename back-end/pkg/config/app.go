package config

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

const dbIp = "10.177.88.190"

var (
	db *gorm.DB
)

const (
	dsn = "root:aMz123@tcp(" + dbIp + ")/iBooking?charset=utf8mb4&parseTime=true&loc=Local"
)

// Connect to the database
func Connect() error {
	d, err := gorm.Open("mysql", dsn)
	if err != nil {
		return err
	}
	db = d
	return nil
}

// GetDB return database
func GetDB() *gorm.DB {
	return db
}
