package main

import (
	"context"
	"fmt"
	"github.com/amzcnx/iBooking/pkg/routes"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"
)

const (
	port = 8800
)

//	@title			iBooking
//	@version		1.0
//	@description	iBooking back-end api.
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	API Support
//	@contact.url	http://www.swagger.io/support
//	@contact.email	support@swagger.io

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

//	@host	10.177.88.252:8800

// @securityDefinitions.apiKey	ApiKeyAuth
// @in							header
// @name						Authorization

const ipAddr = "localhost"

func main() {
	router := gin.Default()

	routes.RegisterBookingRoutes(router)

	fmt.Printf("Starting listening port:%d\n", port)

	srv := &http.Server{
		Addr:    ipAddr + ":" + strconv.Itoa(port),
		Handler: router,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen:%s\n", err)
		}
	}()

	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Println("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}
	log.Println("Server exiting")
}
