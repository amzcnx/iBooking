package middlewares

import (
	"github.com/amzcnx/iBooking/pkg/models"
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"time"
)

func GenerateAdminAuthMiddleware() (*jwt.GinJWTMiddleware, error) {
	authMiddleware, err := jwt.New(&jwt.GinJWTMiddleware{
		Realm:            "identity",
		SigningAlgorithm: "HS256",
		Key:              []byte("secret key"),
		Timeout:          24 * time.Hour,
		MaxRefresh:       time.Hour,
		IdentityKey:      identityKey,
		PayloadFunc: func(data interface{}) jwt.MapClaims {
			if v, ok := data.(*models.Administrator); ok {
				return jwt.MapClaims{
					identityKey: v.Username,
				}
			}
			return jwt.MapClaims{}
		},
		IdentityHandler: func(c *gin.Context) interface{} {
			claims := jwt.ExtractClaims(c)
			return &models.Administrator{
				Username: claims[identityKey].(string),
			}
		},
		Authenticator: adminAuthenticator,
		Authorizator: func(data interface{}, c *gin.Context) bool {
			// fmt.Println("admin")
			if v, ok := data.(*models.Administrator); ok {
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

func adminAuthenticator(c *gin.Context) (interface{}, error) {
	var loginVals models.Administrator
	if err := c.ShouldBind(&loginVals); err != nil {
		return "", jwt.ErrMissingLoginValues
	}
	username := loginVals.Username
	password := loginVals.Password
	admin, err := models.GetAdminByUsername(username)
	if err != nil {
		return "", err
	}
	if err = bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(password)); err != nil {
		return "", jwt.ErrFailedAuthentication
	}
	return admin, nil
}
