package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"github.com/amzcnx/iBooking/pkg/controllers/middlewares"
	"github.com/amzcnx/iBooking/pkg/models"
	"github.com/amzcnx/iBooking/pkg/utils"
)

// CreatAdmin godoc
//
// @Summary		create admin
// @Description	create admin
// @Tags			Admin
// @Accept			json
// @Produce		json
// @Param			admin	body	models.Administrator	true	"admin 's username and password"
//
// @Router			/admin/ [post]
func CreateAdmin(c *gin.Context) {
	json := make(map[string]interface{})
	if err := c.BindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	if json["username"] == nil || json["password"] == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "username or password is nil",
		})
		return
	}
	// encrypt the password, using password + salt(a string of random numbers) and then hash
	hash, err := bcrypt.GenerateFromPassword([]byte(json["password"].(string)), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	admin := models.Administrator{
		ID:       utils.GetID(),
		Username: json["username"].(string),
		Password: string(hash),
	}

	if err := admin.Create(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "create admin success",
		"data":    admin,
	})
}

var AdminAuthMiddleware, adminGinJWTMiddleErr = middlewares.GenerateAdminAuthMiddleware()

// AdminLogin godoc
//
// @Summary		Admin Login
// @Description	admin login
// @Tags			Admin
// @Accept			json
// @Produce		json
// @Param			admin	body	models.Administrator	true	"Admin login with username and password"
//
// @Router			/admin/login/ [post]
func AdminLogin(c *gin.Context) {
	if adminGinJWTMiddleErr != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"error": adminGinJWTMiddleErr.Error(),
		})
	}
	AdminAuthMiddleware.LoginHandler(c)
}
