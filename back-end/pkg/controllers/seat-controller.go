package controllers

import (
	"bytes"
	"github.com/amzcnx/iBooking/pkg/models"
	"github.com/amzcnx/iBooking/pkg/utils"
	"github.com/gin-gonic/gin"
	"net/http"
)

// CreateSeat godoc
//
//	@Summary		create seat
//	@Description	create seat
//	@Tags			Seat
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			admin	body	models.Seat	 true	"Create Seat by giving seat information"
//
//	@Router			/seat/auth/createSeat [post]
func CreateSeat(c *gin.Context) {
	json := make(map[string]interface{})
	if err := c.BindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	var errorMsg bytes.Buffer
	var Seats []models.Seat

	// check if the room exists
	if json["room_id"] == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "roomID is required",
		})
		return
	}
	roomID := utils.Stoi(json["room_id"].(string), 64).(int64)

	if _, err := models.GetRoomById(roomID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	for _, v := range json["seats"].(map[string]interface{}) {
		m := v.(map[string]interface{})

		seat := &models.Seat{
			ID:     utils.GetID(),
			RoomID: roomID,
			X:      utils.Stoi(m["x"].(string), 8).(int8),
			Y:      utils.Stoi(m["y"].(string), 8).(int8),
			Status: utils.Stoi(m["status"].(string), 8).(int8),
			Plug:   m["plug"].(bool),
		}
		if err := seat.Create(); err != nil {
			errorMsg.WriteString(err.Error() + "\n")
			continue
		}
		Seats = append(Seats, *seat)
	}
	if errorMsg.Len() > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": errorMsg.String(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"create seats": Seats,
	})
}

// UpdateSeat godoc
//
//	@Summary		update seat
//	@Description	update seat
//	@Tags			Seat
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			admin	body	models.Seat	 true	"Update Seat by giving seat information"
//
//	@Router			/seat/auth/updateSeat [post]
func UpdateSeat(c *gin.Context) {
	json := make(map[string]map[string]interface{})
	if err := c.BindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	// log.Printf("%v\n", &json)
	var errorMsg bytes.Buffer
	var Seats []models.Seat

	for _, v := range json {
		seatID := utils.Stoi(v["seat_id"].(string), 64).(int64)
		seat, err := models.GetSeatByID(seatID)
		if err != nil {
			errorMsg.WriteString(err.Error() + "\n")
			continue
		}

		if v["room_id"] != nil {
			roomID := utils.Stoi(v["room_id"].(string), 64).(int64)
			if _, err := models.GetRoomById(roomID); err != nil {
				errorMsg.WriteString(err.Error() + "\n")
				continue
			}
			seat.RoomID = roomID
		}
		if v["x"] != nil {
			seat.X = utils.Stoi(v["x"].(string), 8).(int8)
		}
		if v["y"] != nil {
			seat.Y = utils.Stoi(v["y"].(string), 8).(int8)
		}
		if v["status"] != nil {
			seat.Status = utils.Stoi(v["status"].(string), 8).(int8)
		}
		if v["plug"] != nil {
			seat.Plug = v["plug"].(bool)
		}

		if err := models.UpdateSeat(seat); err != nil {
			errorMsg.WriteString(err.Error() + "\n")
			continue
		}
		Seats = append(Seats, *seat)
	}

	if errorMsg.Len() > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": errorMsg.String(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"seats": Seats,
	})
}

// GetSeat godoc
//
//	@Summary		get seat
//	@Description	get all seats
//	@Tags			Seat
//	@Produce		json
//	@Security		ApiKeyAuth
//
//	@Router			/seat/ [get]
func GetSeat(c *gin.Context) {
	seats, err := models.GetAllSeats()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"seats": seats,
	})
}

// DeleteSeat godoc
//
//	@Summary		delete seat
//	@Description	delete seat by id
//	@Tags			Seat
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			seat_id 	body	string	 true	"Delete Seat by giving seat id"
//
//	@Router			/seat/auth/deleteSeat [post]
func DeleteSeat(c *gin.Context) {
	json := make(map[string]interface{})
	if err := c.BindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	if json["seat_id"] == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "seatID is required",
		})
		return
	}
	seatID := utils.Stoi(json["seat_id"].(string), 64).(int64)
	if err := models.DeleteSeat(seatID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "delete seat OK",
	})
}

// GetSeatByID godoc
//
//	@Summary		get seat by id
//	@Description	get seat by id
//	@Tags			Seat
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			seat_id 	path	string	 true	"Get Seat by giving seat id"
//
//	@Router			/seat/{seat_id} [post]
func GetSeatByID(c *gin.Context) {
	if c.Param("seatID") == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "seatID is required",
		})
		return
	}
	seatID := utils.Stoi(c.Param("seatID"), 64).(int64)
	seat, err := models.GetSeatByID(seatID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"seat": seat,
	})
}

// GetSeatByRoomID godoc
//
//	@Summary		get seat by room id
//	@Description	get seat by room id
//	@Tags			Seat
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			room_id 	body	string	 true	"Get Seats by giving room id"
//
//	@Router			/seat/getSeatByRoomID/{room_id} [post]
func GetSeatByRoomID(c *gin.Context) {
	if c.Param("roomID") == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "roomID is required",
		})
		return
	}
	roomID := utils.Stoi(c.Param("roomID"), 64).(int64)
	seats, err := models.GetSeatByRoomID(roomID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "get seats ok",
		"data":    seats,
	})
}
