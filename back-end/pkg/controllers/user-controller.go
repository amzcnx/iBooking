package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"github.com/amzcnx/iBooking/pkg/controllers/middlewares"
	"github.com/amzcnx/iBooking/pkg/models"
	"github.com/amzcnx/iBooking/pkg/utils"
)

const (
	userIDRequiredErrorMessage string = "userID is required"
)

func UpdateUserinfo(userinfo *models.UserInfo, info map[string]interface{}) {
	if info["email"] != nil {
		userinfo.Email = info["email"].(string)
	}
	if info["gender"] != nil {
		userinfo.Gender = info["gender"].(string)
	}
	if info["number_defaults"] != nil {
		userinfo.NumberDefaults = utils.Stoi(info["number_defaults"].(string), 32).(int32)
	}
	if info["accept_notification"] != nil {
		userinfo.AcceptNotification = info["accept_notification"].(bool)
	}
}

// CreateUser godoc
//
// @Summary		create user
// @Description	create user
// @Tags			User
// @Accept			json
// @Produce		json
// @Param			user	body	models.User	true	"user 's username and password"
//
// @Router			/user/ [post]
func CreateUser(c *gin.Context) {
	json := make(map[string]interface{})
	if err := c.BindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	if json["username"] == nil || json["password"] == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "username or password is required",
		})
		return
	}
	pwd, err := utils.Encrypt(json["password"].(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	user := models.User{
		ID:       utils.GetID(),
		Username: json["username"].(string),
		Password: pwd,
	}

	// log.Printf("password: %v\n", string(hash))
	// TODO:test
	userinfo := models.UserInfo{
		UserID:   user.ID,
		Username: user.Username,
	}

	if json["userinfo"] != nil {
		info := json["userinfo"].(map[string]interface{})
		UpdateUserinfo(&userinfo, info)
		// log.Printf("userinfo: %v\n", userinfo)
	}

	if err := user.Create(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	if err := userinfo.Create(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message":  "create user success",
		"user":     user,
		"userinfo": userinfo,
	})
}

var UserAuthMiddleware, userGinJWTMiddleErr = middlewares.GenerateUserAuthMiddleware()

// UserLogin godoc
//
// @Summary	 	user login
// @Description	user login
// @Tags			User
// @Accept			json
// @Produce		json
// @Param			user	body	models.User	true	"user 's username and password"
//
// @Router			/user/login [post]
func UserLogin(c *gin.Context) {
	if userGinJWTMiddleErr != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": userGinJWTMiddleErr.Error(),
		})
		return
	}

	UserAuthMiddleware.LoginHandler(c)
}

func UserLogout(c *gin.Context) {
	if userGinJWTMiddleErr != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": userGinJWTMiddleErr.Error(),
		})
		return
	}
	UserAuthMiddleware.LogoutHandler(c)
}

func UserRefreshToken(c *gin.Context) {
	if userGinJWTMiddleErr != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": userGinJWTMiddleErr.Error(),
		})
		return
	}
	token, expire, err := UserAuthMiddleware.RefreshToken(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
	}
	c.JSON(http.StatusOK, gin.H{
		"message":   "refresh token ok",
		"token":     token,
		"expire at": expire,
	})
}

// DeleteUser godoc
//
// @Summary	 	delete user
// @Description	delete user
// @Tags			User
// @Accept			json
// @Produce		json
// @Param			user_id 	body	string	true	"user id"
//
// @Router			/user/auth/deleteUser [post]
func DeleteUser(c *gin.Context) {
	json := make(map[string]string)
	if err := c.BindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	if json["user_id"] == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": userIDRequiredErrorMessage,
		})
		return
	}
	userID := utils.Stoi(json["user_id"], 64).(int64)
	if err := models.DeleteUserInfo(userID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	if err := models.DeleteUser(userID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "delete user OK",
	})
}

// UpdateUser godoc
//
// @Summary	 	update user
// @Description	update user
// @Tags			User
// @Accept			json
// @Produce		json
// @Param			admin	body	models.User	true	"user information"
//
// @Router			/user/auth/updateUser [post]
func UpdateUser(c *gin.Context) {
	json := make(map[string]interface{})
	if err := c.BindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	if json["id"] == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": userIDRequiredErrorMessage,
		})
		return
	}
	userID := utils.Stoi(json["user_id"].(string), 64).(int64)
	userinfo, err := models.GetUserinfoByUserID(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	UpdateUserinfo(userinfo, json)

	if err := models.UpdateUserInfo(userinfo); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "update user OK",
		"data":    userinfo,
	})
}

// GetUserByID godoc
//
// @Summary	 		get user by id
// @Description		get user by id
// @Tags			User
// @Accept			json
// @Produce			json
// @Param			user_id	path	string	true	"user id"
//
// @Router			/user/auth/getUserByID/{user_id} [get]
func GetUserByID(c *gin.Context) {
	if c.Param("userID") == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": userIDRequiredErrorMessage,
		})
		return
	}
	userID := utils.Stoi(c.Param("userID"), 64).(int64)
	user, err := models.GetUserByID(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	userinfo, err := models.GetUserinfoByUserID(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
			"data":  userinfo,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message":  "get user OK",
		"user":     user,
		"userinfo": userinfo,
	})
}

// GetUserByUsername godoc
//
// @Summary	 		get user by username
// @Description		get user by username
// @Tags			User
// @Accept			json
// @Produce			json
// @Param			username	path	string	true	"username"
//
// @Router			/user/auth/getUserByUsername/{username} [get]
func GetUserByUsername(c *gin.Context) {
	if c.Param("username") == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "username is required",
		})
		return
	}
	username := c.Param("username")
	user, err := models.GetUserByUsername(username)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	userinfo, err := models.GetUserinfoByUsername(username)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
			"data":  userinfo,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message":  "get user OK",
		"user":     user,
		"userinfo": userinfo,
	})
}

// UpdatePassword godoc
//
// @Summary	 		update password
// @Description		update password
// @Tags			User
// @Accept			json
// @Produce			json
// @Param			userinfo	body	string	true	"userID and password"
//
// @Router			/user/auth/password [post]
func UpdatePassword(c *gin.Context) {
	json := make(map[string]interface{})
	if err := c.BindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	if json["user_id"] == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "user id is required",
		})
		return
	}
	userID := utils.Stoi(json["user_id"].(string), 64).(int64)
	user, err := models.GetUserByID(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	if json["password"] == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "password is required",
		})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(json["password"].(string))); err == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "same password",
		})
		return
	}
	pwd, err := utils.Encrypt(json["password"].(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	user.Password = pwd
	if err := models.UpdateUser(user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "update password OK",
	})
}

func recordDefault(id int64) error {
	userinfo, err := models.GetUserinfoByUserID(id)
	if err != nil {
		return err
	}
	userinfo.NumberDefaults += 1
	return models.UpdateUserInfo(userinfo)
}
