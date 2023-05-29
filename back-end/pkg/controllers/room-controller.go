package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/amzcnx/iBooking/pkg/models"
	"github.com/amzcnx/iBooking/pkg/utils"
)

// CreateRoom godoc
//
//	@Summary		create room
//	@Description	create room
//	@Tags			Room
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			admin	body	models.Room	true	"Create Room by giving room information"
//
//	@Router			/room/auth/createRoom [post]
func CreateRoom(c *gin.Context) {
	json := make(map[string]interface{})
	if err := c.BindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	// log.Printf("%v\n", &json)

	// check time
	openTime, closeTime := utils.Stoi(json["open_time"].(string), 8).(int8), utils.Stoi(json["close_time"].(string), 8).(int8)
	if !check(openTime, closeTime) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "wrong time setting",
		})
		return
	}

	room := &models.Room{
		ID:         utils.GetID(),
		RoomNumber: json["room_number"].(string),
		Location:   json["location"].(string),
		OpenTime:   openTime,
		CloseTime:  closeTime,
		Overnight:  json["overnight"].(bool),
		// new room should have no seat, waiting for admin to create
		Total:  0,    //utils.Stoi(json["total"].(string), 16).(int16),
		Free:   0,    //utils.Stoi(json["free"].(string), 16).(int16),
		Usable: true, // default room is usable
	}

	if err := room.Create(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"room": room,
	})
}

// GetRoom godoc
//
//	@Summary		get all room
//	@Description	get all room information
//	@Tags			Room
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//
//	@Router			/room/ [get]
func GetRoom(c *gin.Context) {
	rooms, err := models.GetAllRooms()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"data":  rooms,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"rooms": rooms,
	})
}

// GetRoomByID godoc
//
//	@Summary		get room
//	@Description	get room by id
//	@Tags			Room
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			admin	path	string	true	"Create Room by giving room information"
//
//	@Router			/room/auth/ [get]
func GetRoomByID(c *gin.Context) {
	if c.Param("roomID") == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "empty id"})
	}
	roomID := utils.Stoi(c.Param("roomID"), 64).(int64)
	room, err := models.GetRoomById(roomID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"data":  room,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"room": room,
	})
}

// DeleteRoom godoc
//
//	@Summary		delete room
//	@Description	delete room by id
//	@Tags			Room
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			room_id	 path	string	 true	"Delete Room by giving room id"
//
//	@Router			/room/auth/deleteRoom [post]
func DeleteRoom(c *gin.Context) {
	json := make(map[string]interface{})
	if err := c.BindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	// log.Printf("update room:%v\n", &json)
	if json["room_id"] == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "roomID is required",
		})
		return
	}
	roomID := utils.Stoi(json["room_id"].(string), 64).(int64)
	// delete the seat in the room
	if err := models.DeleteSeatByRoomID(roomID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	if err := models.DeleteRoom(roomID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "delete OK",
	})
}

// UpdateRoom godoc
//
//	@Summary		update room
//	@Description	update room
//	@Tags			Room
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			room	 body	models.Room	 true	"Update Room by giving room id and room details"
//
//	@Router			/room/auth/updateRoom [post]
func UpdateRoom(c *gin.Context) {
	json := make(map[string]interface{})
	if err := c.BindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	// log.Printf("update room:%v\n", &json)
	if json["room_id"] == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "roomID is required",
		})
		return
	}

	roomID := utils.Stoi(json["room_id"].(string), 64).(int64)
	room, err := models.GetRoomById(roomID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	if json["room_number"] != nil {
		room.RoomNumber = json["room_number"].(string)
	}
	if json["location"] != nil {
		room.Location = json["location"].(string)
	}
	// using UpdateSeat to update,not here
	//if json["seat"] != nil {
	//	updateSeats(json["seat"].(map[string]interface{}))
	//}
	if json["total"] != nil {
		room.Total = utils.Stoi(json["total"].(string), 16).(int16)
	}
	if json["free"] != nil {
		room.Free = utils.Stoi(json["free"].(string), 16).(int16)
	}
	// study room is usable or not
	if json["usable"] != nil {
		room.Usable = json["usable"].(bool)
	}
	if json["overnight"] != nil {
		room.Overnight = json["overnight"].(bool)
	}
	// time modify
	openTime, closeTime := room.OpenTime, room.CloseTime
	if json["open_time"] != nil {
		openTime = utils.Stoi(json["open_time"].(string), 8).(int8)
	}
	if json["close_time"] != nil {
		closeTime = utils.Stoi(json["close_time"].(string), 8).(int8)
	}
	if !check(openTime, closeTime) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "wrong time setting",
		})
		return
	}

	if err := models.UpdateRoom(room); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message":    "update room ok",
		"updateRoom": room,
	})
}

// check if time is available
func check(x int8, y int8) bool {
	if x < 0 || x > 24 || y < 0 || y > 24 || x > y {
		return false
	}
	return true
}
