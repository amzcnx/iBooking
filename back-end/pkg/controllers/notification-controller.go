package controllers

import (
	"errors"
	"fmt"
	"github.com/amzcnx/iBooking/pkg/models"
	"github.com/jordan-wright/email"
	"net/smtp"
)

func NotifyByEmail(userID int64, subject string, message string) error {
	userinfo, err := models.GetUserinfoByUserID(userID)
	if err != nil {
		return err
	}
	emailAddress := userinfo.Email
	if emailAddress == "" {
		return errors.New("user haven't set email address")
	}
	// send notification by email
	e := email.NewEmail()
	e.From = "iBooking <1514000750@qq.com>"
	e.To = []string{emailAddress}
	e.Subject = "iBooking " + subject
	e.Text = []byte(message)
	fmt.Println(e)
	err = e.Send("smtp.qq.com:25", smtp.PlainAuth("", "1514000750@qq.com", "eptxbsxbdzryjgah", "smtp.qq.com"))
	return err
}
