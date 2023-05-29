package controllers

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/araddon/dateparse"
	"github.com/gin-gonic/gin"
	"github.com/robfig/cron"

	"github.com/amzcnx/iBooking/pkg/models"
	"github.com/amzcnx/iBooking/pkg/utils"
)

// TODO:
// 1. 需要修改座位参数，按时间段座位可预约，而不是free属性
// 5. 暂时未考虑未来日期的预约

const (
	bookingIDIsRequiredErrorMessage = "bookingID is required"
)

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
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	if _, ok := json["user_id"]; !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "user_id is required",
		})
		return
	}
	if _, ok := json["seat_id"]; !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "seat_id is required",
		})
		return
	}
	if _, ok := json["room_id"]; !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "room_id is required",
		})
		return
	}
	fmt.Println(time.Now())
	userID := utils.Stoi(json["user_id"].(string), 64).(int64)

	userinfo, err := models.GetUserinfoByUserID(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	if userinfo.Email == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "user email is required",
		})
		return
	}

	// 同一用户不能重复预约
	_, err = models.GetBookingByUserID(userID)
	if err != nil {
		if err.Error() != "record not found" {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(), // not a user
			})
			return
		}
	}

	if _, ok := json["duration"]; !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "booking duration is required",
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

	if _, ok := json["booking_time"]; !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "booking time is required",
		})
		return
	}

	log.Println(json["booking_time"].(string))
	bookTime, err := dateparse.ParseLocal(json["booking_time"].(string))

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	var booking = models.Booking{
		ID:          utils.GetID(),
		UserID:      userID,
		SeatID:      utils.Stoi(json["seat_id"].(string), 64).(int64),
		RoomID:      utils.Stoi(json["room_id"].(string), 64).(int64), // no check
		Duration:    duration,                                         // booking duration, max 4 hour
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
	notifyTime := booking.BookingTime.Add(-1 * time.Minute) // 10 minutes ago
	h, m, s := notifyTime.Clock()

	// 预约时间之前15分钟未签到提醒，15分钟内不提醒
	spec := solveTime(s, m, h)
	c1 := cron.New()
	if err := c1.AddFunc(spec, func() {
		// 未签到
		bookingInfo, err := models.GetBookingByID(booking.ID)
		if err != nil {
			log.Println(err.Error())
			c1.Stop()
			return
		}
		if bookingInfo.IsSigned != 1 {
			log.Println("距离预约时间还差10分钟，发送提醒消息")
			err := NotifyByEmail(booking.UserID, "Your seat booking time is coming soon", "Your seat booking time will due in 15 minutes")
			if err != nil {
				log.Fatal(err)
			}
		}
		c1.Stop()
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	go c1.Start()

	// 预约时间之后10分钟未签到提醒
	notifyTime = notifyTime.Add(1 * time.Minute)
	h, m, s = notifyTime.Clock()
	spec = solveTime(s, m, h)

	c2 := cron.New()
	if err := c2.AddFunc(spec, func() {
		bookingInfo, err := models.GetBookingByID(booking.ID)
		if err != nil {
			log.Println(err.Error())
			c1.Stop()
			return
		}
		// 未签到
		if bookingInfo.IsSigned != 1 {
			log.Println("用户超过10分钟未签到，发送提醒消息")
			err := NotifyByEmail(booking.UserID, "Your appointment time has expired", "Your appointment time has expired by 10 minutes")
			if err != nil {
				log.Fatal(err)
			}
		}
		c2.Stop()
	}); err != nil {
		log.Println("c2 error")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	go c2.Start()

	// 预约时间之后15分钟未签到，自动取消预约，释放座位，提醒学生，记录一次违约
	//timeDifference += time.Minute * 1
	//s, m, h = int(timeDifference.Seconds())%60, int(timeDifference.Minutes())%60, int(timeDifference.Hours())%60
	notifyTime = notifyTime.Add(1 * time.Minute)
	h, m, s = notifyTime.Clock()
	spec = solveTime(s, m, h)

	c3 := cron.New()
	if err := c3.AddFunc(spec, func() {
		bookingInfo, err := models.GetBookingByID(booking.ID)
		if err != nil {
			log.Println(err.Error())
			c1.Stop()
			return
		}
		// 未签到
		if bookingInfo.IsSigned != 1 {
			log.Println("用户超时15分钟未签到，自动取消预约，释放座位，记录一次违约，发送提醒消息")
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
			if err := NotifyByEmail(booking.UserID, "Your appointment time has expired",
				"Your appointment time has expired more than 15 minutes, a default will be recorded"); err != nil {
				log.Fatal(err)
			}
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
	notifyTime = notifyTime.Add(time.Hour * time.Duration(duration))
	h, m, s = notifyTime.Clock()
	spec = solveTime(s, m, h)

	c4 := cron.New()
	if err := c4.AddFunc(spec, func() {
		bookingInfo, err := models.GetBookingByID(booking.ID)
		if err != nil {
			log.Println(err.Error())
			c1.Stop()
			return
		}
		// 签到后执行
		if bookingInfo.IsSigned == 1 {
			room, err := models.GetRoomById(booking.RoomID)
			if err != nil {
				log.Println(err)
				return
			}
			bk := models.BookingHistory{
				ID:           utils.GetID(),
				UserID:       booking.UserID,
				SeatID:       booking.SeatID,
				RoomID:       booking.RoomID,
				BookingTime:  booking.BookingTime,
				RoomLocation: room.Location,
				RoomNumber:   room.RoomNumber,
			}
			// create booking history and delete booking
			if err := bk.Create(); err != nil {
				log.Println(err)
				return
			}
		}
		c4.Stop()
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	go c4.Start()

	room, _ := models.GetRoomById(booking.RoomID)

	// return
	c.JSON(http.StatusOK, gin.H{
		"message":       "booking created successfully",
		"booking":       booking,
		"room_location": room.Location,
		"room_number":   room.RoomNumber,
	})
}

func solveTime(s int, m int, h int) string {
	spec := fmt.Sprintf("%d", s) + " " + fmt.Sprintf("%d", m) + " " + fmt.Sprintf("%d", h) + " * * *"
	return spec
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
		return
	}
	booking, err := models.GetBookingByUserID(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"data":  booking,
			"error": err.Error(),
		})
		return
	}
	room, err := models.GetRoomById(booking.RoomID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"data":  booking,
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "bookings retrieved successfully",
		"booking":       booking,
		"room_location": room.Location,
		"room_number":   room.RoomNumber,
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
//	@Router			/booking/getBookingByID/{booking_id} [get]
func GetBookingByID(c *gin.Context) {
	if c.Param("bookingID") == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": bookingIDIsRequiredErrorMessage,
		})
		return
	}
	bookingID := utils.Stoi(c.Param("bookingID"), 64).(int64)
	booking, err := models.GetBookingByID(bookingID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"data":  booking,
			"error": err.Error(),
		})
		return
	}
	room, err := models.GetRoomById(booking.RoomID)
	c.JSON(http.StatusOK, gin.H{
		"message":       "booking retrieved successfully",
		"booking":       booking,
		"room_location": room.Location,
		"room_number":   room.RoomNumber,
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
			"error": bookingIDIsRequiredErrorMessage,
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
//	@Router			/booking/updateBooking [post]
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
			"error": bookingIDIsRequiredErrorMessage,
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

	log.Println("用户签到")

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

// GetBookingHistory godoc
//
//	@Summary		get booking history
//	@Description	get booking history
//	@Tags			Booking
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			userID 	path	string	true	"booking history"
//
//	@Router			/booking/bookingHistory/{userID} [get]
func GetBookingHistory(c *gin.Context) {
	if c.Param("userID") == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "userID is required",
		})
		return
	}
	userID := utils.Stoi(c.Param("userID"), 64).(int64)
	bhs, err := models.GetBookingHistoryByUserID(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
			"data":  bhs,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "get booking history ok",
		"data":    bhs,
	})
}
