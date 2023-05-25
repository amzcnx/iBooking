package controllers

import (
	"fmt"
	"github.com/amzcnx/iBooking/pkg/models"
	"github.com/amzcnx/iBooking/pkg/utils"
	"github.com/araddon/dateparse"
	"github.com/gin-gonic/gin"
	"github.com/robfig/cron"
	"log"
	"net/http"
	"time"
)

var isSignedFlag = false

// TODO:
// 1. 需要修改座位参数，按时间段座位可预约，而不是free属性
// 2. 座位使用结束后，将预约记录添加至预约历史记录

// BookSeat godoc
//
//	@Summary		book seat
//	@Description	book seat
//	@Tags			Booking
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			id	body	string	true	"Book a seat by giving seat_id and user_id"
//
//	@Router			/booking/ [post]
func BookSeat(c *gin.Context) {
	var json map[string]interface{}
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if json["user_id"] == "" || json["seat_id"] == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "user_id or seat_id is required",
		})
		return
	}

	duration := utils.Stoi(json["duration"].(string), 8).(int8)
	if duration < 0 || duration > 4 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "wrong duration",
		})
		return
	}

	bookTime, err := dateparse.ParseLocal(json["booking_time"].(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "parsing booking time error",
		})
		return
	}

	var booking = models.Booking{
		ID:          utils.GetID(),
		UserID:      utils.Stoi(json["user_id"].(string), 64).(int64),
		SeatID:      utils.Stoi(json["seat_id"].(string), 64).(int64),
		Duration:    duration, // booking duration, max 4 hour
		IsSigned:    0,
		BookingTime: bookTime, // no examine
	}

	if err := booking.Create(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	// Each appointment is based on hourly hours, and the maximum one-time appointment is 4 hours (system parameters are adjustable)
	timeDifference := booking.BookingTime.Sub(time.Now()) - time.Minute*15

	s, m, h := int(timeDifference.Seconds()), int(timeDifference.Minutes()), int(timeDifference.Hours())

	// 预约时间之前15分钟未签到提醒，15分钟内不提醒
	if s > 0 || m > 0 || h > 0 {
		spec := solveTime(s, m, h)
		c1 := cron.New()
		if err := c1.AddFunc(spec, func() {
			if isSignedFlag {
				// 签到后取消
				c1.Stop()
				return
			}
			err := NotifyByEmail(booking.UserID, "Your seat booking time is coming soon", "Your seat booking time will due in 15 minutes")
			if err != nil {
				log.Fatal(err)
			}
			c1.Stop()
		}); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		go c1.Start()
	}
	// 预约时间之后10分钟未签到提醒
	timeDifference += time.Minute * 25
	s, m, h = int(timeDifference.Seconds()), int(timeDifference.Minutes()), int(timeDifference.Hours())
	spec := solveTime(s, m, h)

	c2 := cron.New()
	if err := c2.AddFunc(spec, func() {
		if isSignedFlag {
			// 签到后取消
			c2.Stop()
			return
		}
		err := NotifyByEmail(booking.UserID, "Your appointment time has expired", "Your appointment time has expired by 10 minutes")
		if err != nil {
			log.Fatal(err)
		}
		c2.Stop()
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	go c2.Start()
	// 预约时间之后15分钟未签到，自动取消预约，释放座位，提醒学生，记录一次违约
	timeDifference += time.Minute * 5
	s, m, h = int(timeDifference.Seconds()), int(timeDifference.Minutes()), int(timeDifference.Hours())
	spec = solveTime(s, m, h)

	c3 := cron.New()
	if err := c3.AddFunc(spec, func() {
		if isSignedFlag {
			// 签到后取消
			c3.Stop()
			return
		}
		// 自动取消预约
		if err := models.DeleteBooking(booking.ID); err != nil {
			if err != nil {
				log.Fatal(err)
			}
			return
		}
		// 记录一次违约
		if err := recordDefault(booking.UserID); err != nil {
			if err != nil {
				log.Fatal(err)
			}
			return
		}
		// 提醒学生
		err := NotifyByEmail(booking.UserID, "Your appointment time has expired",
			"Your appointment time has expired more than 15 minutes, a default will be recorded")
		if err != nil {
			log.Fatal(err)
		}
		c3.Stop()
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	go c3.Start()

	// 预约结束后释放座位，将预约存入历史记录
	timeDifference -= 15 * time.Minute
	timeDifference += time.Duration(duration) * time.Hour

	s, m, h = int(timeDifference.Seconds()), int(timeDifference.Minutes()), int(timeDifference.Hours())
	spec = solveTime(s, m, h)

	c4 := cron.New()
	if err := c4.AddFunc(spec, func() {
		if isSignedFlag {
			// 签到后执行

		}
		c4.Stop()
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	go c4.Start()

	// return
	c.JSON(http.StatusOK, gin.H{
		"message": "booking created successfully",
		"data":    booking,
	})
}

func solveTime(s int, m int, h int) (spec string) {
	if s > 0 {
		spec += fmt.Sprintf("%d", s)
	} else {
		spec += "?"
	}
	spec += " "
	if m > 0 {
		spec += fmt.Sprintf("%d", m)
	} else {
		spec += "?"
	}
	spec += " "
	if h > 0 {
		spec += fmt.Sprintf("%d", h)
	} else {
		spec += "?"
	}
	spec += " ? ? ?"
	return
}

// GetBookingByUserID godoc
//
//	@Summary		get booking by user ID
//	@Description	get booking by user ID
//	@Tags			Booking
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			user_id	path	string	true	"user id"
//
//	@Router			/booking/getBookingByUserID/{user_id} [get]
func GetBookingByUserID(c *gin.Context) {
	if c.Param("userID") == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "userID is required",
		})
		return
	}
	userID := utils.Stoi(c.Param("userID"), 64).(int64)
	if _, err := models.GetUserByID(userID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
	}
	bookings, err := models.GetBookingByUserID(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "bookings retrieved successfully",
		"data":    bookings,
	})
}

// GetBookingByID godoc
//
//	@Summary		get booking by ID
//	@Description	get booking by ID
//	@Tags			Booking
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			booking_id	path	string	true	"booking id"
//
//	@Router			/booking/getBookingByUserID/{booking_id} [get]
func GetBookingByID(c *gin.Context) {
	if c.Param("bookingID") == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "bookingID is required",
		})
		return
	}
	bookingID := utils.Stoi(c.Param("bookingID"), 64).(int64)
	booking, err := models.GetBookingByID(bookingID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "booking retrieved successfully",
		"data":    booking,
	})
}

// DeleteBooking godoc
//
//	@Summary		delete booking
//	@Description	delete booking by ID
//	@Tags			Booking
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			booking_id	body	string	true	"booking id"
//
//	@Router			/booking/deleteBooking [post]
func DeleteBooking(c *gin.Context) {
	json := make(map[string]interface{})
	if err := c.BindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	if json["booking_id"] == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "bookingID is required",
		})
		return
	}
	bookingID := utils.Stoi(json["booking_id"].(string), 64).(int64)
	if err := models.DeleteBooking(bookingID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "booking deleted successfully",
	})
}

// UpdateBooking godoc
//
//	@Summary		update booking
//	@Description	update booking , change isSigned
//	@Tags			Booking
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			booking 	body	models.Booking	true	"booking information"
//
//	@Router			/booking/deleteBooking [post]
func UpdateBooking(c *gin.Context) {
	json := make(map[string]interface{})
	if err := c.BindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	if json["booking_id"] == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "bookingID is required",
		})
		return
	}
	bookingID := utils.Stoi(json["booking_id"].(string), 64).(int64)
	booking, err := models.GetBookingByID(bookingID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	if json["is_signed"] == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "change information is required",
		})
		return
	}
	isSigned := utils.Stoi(json["is_signed"].(string), 8).(int8)
	if booking.IsSigned == isSigned {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "not changed",
		})
		return
	}
	booking.IsSigned = isSigned

	isSignedFlag = true

	if err := models.UpdateBooking(bookingID, booking); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "booking updated successfully",
		"data":    booking,
	})
}
