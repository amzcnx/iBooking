package routes

import (
	docs "github.com/amzcnx/iBooking/docs" // main 文件中导入 docs 包
	"github.com/amzcnx/iBooking/pkg/controllers"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"net/http"
)

func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		origin := c.Request.Header.Get("Origin")
		if origin != "" {
			c.Header("Access-Control-Allow-Origin", "*") // 可将将 * 替换为指定的域名
			c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, UPDATE")
			c.Header("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept, Authorization")
			c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Cache-Control, Content-Language, Content-Type")
			c.Header("Access-Control-Allow-Credentials", "true")
		}
		if method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
		}
		c.Next()
	}
}

var RegisterBookingRoutes = func(router *gin.Engine) {

	router.Use(Cors())

	docs.SwaggerInfo.BasePath = ""
	// swagger
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// administrator management
	adminRouter := router.Group("/admin")
	{
		adminRouter.POST("/", controllers.CreateAdmin)
		adminRouter.POST("/login", controllers.AdminLogin)
	}

	// room management
	roomRouter := router.Group("/room")
	roomRouter.Use(controllers.UserAuthMiddleware.MiddlewareFunc())
	{
		roomRouter.GET("/", controllers.GetRoom)
		roomRouter.GET("/:roomID", controllers.GetRoomByID)
		auth := roomRouter.Group("/auth")
		auth.Use(controllers.AdminAuthMiddleware.MiddlewareFunc())
		{
			auth.GET("/", controllers.GetRoom)
			auth.GET("/:roomID", controllers.GetRoomByID)
			auth.POST("/createRoom", controllers.CreateRoom)
			auth.POST("/updateRoom", controllers.UpdateRoom)
			auth.POST("/deleteRoom", controllers.DeleteRoom)
		}
	}

	// seat management
	seatRouter := router.Group("/seat")
	seatRouter.Use(controllers.UserAuthMiddleware.MiddlewareFunc())
	{
		seatRouter.GET("/", controllers.GetSeat)
		seatRouter.GET("/:seatID", controllers.GetSeatByID)
		seatRouter.GET("/getSeatByRoomID/:roomID", controllers.GetSeatByRoomID)
		auth := seatRouter.Group("/auth")
		auth.Use(controllers.AdminAuthMiddleware.MiddlewareFunc())
		{
			auth.GET("/", controllers.GetSeat)
			auth.GET("/:seatID", controllers.GetSeatByID)
			auth.POST("/createSeat", controllers.CreateSeat)
			auth.POST("/updateSeat", controllers.UpdateSeat)
			auth.POST("/deleteSeat", controllers.DeleteSeat)
		}
	}

	// user management
	userRouter := router.Group("/user")
	{
		userRouter.POST("/", controllers.CreateUser)
		userRouter.POST("/login", controllers.UserLogin)
		auth := userRouter.Group("/auth")
		auth.Use(controllers.UserAuthMiddleware.MiddlewareFunc())
		{
			auth.POST("/logout", controllers.UserLogout)
			auth.POST("/refreshToken", controllers.UserRefreshToken)
			auth.POST("/updateUser", controllers.UpdateUser)
			auth.POST("/deleteUser", controllers.DeleteUser)
			auth.GET("/getUserByID/:userID", controllers.GetUserByID)
			auth.GET("/getUserByUsername/:username", controllers.GetUserByUsername)
			auth.POST("/password/", controllers.UpdatePassword)
		}
	}

	// booking management
	bookingRouter := router.Group("/booking")
	bookingRouter.Use(controllers.UserAuthMiddleware.MiddlewareFunc())
	{
		bookingRouter.POST("/", controllers.BookSeat)
		bookingRouter.GET("/getBookingByID/:bookingID", controllers.GetBookingByID)
		bookingRouter.GET("/getBookingByUserID/:userID", controllers.GetBookingByUserID)
		bookingRouter.POST("/updateBooking", controllers.UpdateBooking)             // update or attend
		bookingRouter.POST("/deleteBooking", controllers.DeleteBooking)             // cancel
		bookingRouter.GET("/bookingHistory/:userID", controllers.GetBookingHistory) // history
	}

	//// notification management
	//notificationRouter := router.Group("/notification")
	//{
	//	notificationRouter.POST("/", controllers.Notify)
	//}

	// default 404
	router.NoRoute(func(c *gin.Context) {
		c.HTML(http.StatusOK, "template/404.html", nil)
	})
}
