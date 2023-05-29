package middlewares

import (
	"time"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"github.com/amzcnx/iBooking/pkg/models"
)

var identityKey = "username"

func GenerateUserAuthMiddleware() (*jwt.GinJWTMiddleware, error) {
	authMiddleware, err := jwt.New(&jwt.GinJWTMiddleware{
		Realm:            "identity",
		SigningAlgorithm: "HS256",
		Key:              []byte("secret key"),
		Timeout:          24 * time.Hour,
		MaxRefresh:       time.Hour,
		IdentityKey:      identityKey,
		PayloadFunc: func(data interface{}) jwt.MapClaims {
			if v, ok := data.(*models.User); ok {
				return jwt.MapClaims{
					identityKey: v.Username,
				}
			}
			return jwt.MapClaims{}
		},
		IdentityHandler: func(c *gin.Context) interface{} {
			claims := jwt.ExtractClaims(c)
			return &models.User{
				Username: claims[identityKey].(string),
			}
		},
		Authenticator: userAuthenticator,
		Authorizator: func(data interface{}, c *gin.Context) bool {
			// fmt.Println("user")
			if v, ok := data.(*models.User); ok {
				if _, err := models.GetUserByUsername(v.Username); err == nil {
					return true
				}
				if _, err := models.GetAdminByUsername(v.Username); err == nil {
					return true
				}
			}
			return false
		},
		Unauthorized: func(c *gin.Context, code int, message string) {
			c.JSON(code, gin.H{
				"error": message,
			})
		},
		TokenLookup:   "header:Authorization,query:token,cookie:jwt",
		TokenHeadName: "Bearer",
		TimeFunc:      time.Now,
	})
	if err != nil {
		return nil, err
	}
	return authMiddleware, nil
}

// userAuthenticator process login request
func userAuthenticator(c *gin.Context) (interface{}, error) {
	var loginVals models.User
	if err := c.ShouldBind(&loginVals); err != nil {
		return "", jwt.ErrMissingLoginValues
	}
	userName := loginVals.Username
	password := loginVals.Password
	user, err := models.GetUserByUsername(userName)
	if err != nil {
		return "", err
	}
	if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", jwt.ErrFailedAuthentication
	}
	return user, nil
}
