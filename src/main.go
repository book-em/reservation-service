package main

import (
	"bookem-reservation-service/client/roomclient"
	"bookem-reservation-service/client/userclient"
	internal "bookem-reservation-service/internal"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

var (
	server *gin.Engine
	dB     *gorm.DB
	rawDB  *sql.DB
)

func syncDatabase() {
	dB.AutoMigrate(&internal.Reservation{})
	dB.AutoMigrate(&internal.ReservationRequest{})
}

func connectToDb() {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	dbURL := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname,
	)

	db, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to open DB: %v", err)
	}

	dB = db
	rawDB, _ = db.DB()

	log.Printf("Connected to DB!")
}

func main() {
	connectToDb()
	defer rawDB.Close()
	syncDatabase()

	server = gin.Default()

	server.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://localhost", "http://bookem.local"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Authorization", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	server.GET("/healthz", func(ctx *gin.Context) {
		err := rawDB.Ping()
		if err != nil {
			ctx.JSON(http.StatusServiceUnavailable, "Database not reachable")
			return
		}
		ctx.JSON(http.StatusOK, nil)
	})

	userClient := userclient.NewUserClient()
	roomClient := roomclient.NewRoomClient()

	reservationRepo := internal.NewRepository(dB)

	service := internal.NewService(reservationRepo, userClient, roomClient)
	handler := internal.NewHandler(service)
	route := *internal.NewRoute(handler)

	rg := server.Group("/api")
	route.Route(rg)

	server.Run()
}
