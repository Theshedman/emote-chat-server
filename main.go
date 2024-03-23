package main

import (
	"chat-server/auth"
	"chat-server/controller"
	"chat-server/rabbitmq"
	"chat-server/repository"
	"chat-server/websocket"
	"context"
	"errors"
	"github.com/joho/godotenv"
	"github.com/labstack/echo-contrib/echoprometheus"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"log"
	"net/http"
	"os"
)

func main() {
	// Instantiate echo server
	e := echo.New()

	// Middlewares
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.Gzip())
	e.Use(middleware.RemoveTrailingSlash())
	e.Use(echoprometheus.NewMiddleware("chatServer"))
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete},
	}))

	// Load .env file to system environment
	err := godotenv.Load()
	if err != nil {
		log.Println("no .env file present on the server project path")
	}

	// Initialize RabbitMQ connection
	rmq, err := rabbitmq.New()
	if err != nil {
		log.Fatal("Failed to initialize RabbitMQ:", err)
	}
	err = rmq.DeclareExchange(rabbitmq.ExchangeName)
	if err != nil {
		log.Fatal("failed to declare connection exchange: ", err)
	}

	// Initialize WebSocket handler
	socketHandler := websocket.New(rmq)
	go socketHandler.ConsumeMessages(rabbitmq.ExchangeName, rabbitmq.QueueName)

	// Initialize MongoDB
	repository.SetupDatabase()

	// Public route for metrics
	e.GET("/metrics", echoprometheus.NewHandler())

	// Auth route for signup and login
	// these routes are public and does not require authentication
	authRoute := e.Group("/auth")
	authRoute.POST("/signup", controller.Signup)
	authRoute.POST("/login", controller.Login)

	// Protected routes grouped according to their resources
	protectedRoute := e.Group("", echojwt.WithConfig(auth.JwtCustomConfig()))

	// Routes for the room resource
	roomRoute := protectedRoute.Group("/rooms")
	roomRoute.GET("", controller.GetRooms)
	roomRoute.POST("/:roomName/join", controller.Join)

	// Routes for the message resource
	msgRoute := protectedRoute.Group("/messages")
	msgRoute.GET("", controller.GetMessages)

	// Routes for websocket connection - chat
	wsRoute := protectedRoute.Group("/websocket")
	wsRoute.GET("", socketHandler.HandleConnection)

	// Start the webserver
	serverPort := os.Getenv("SERVER_PORT")
	if serverPort == "" {
		serverPort = "8080"
	}
	server := http.Server{
		Addr:    ":" + serverPort,
		Handler: e,
	}
	if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}

	// AFTER server shutdown
	// close database connection
	err = repository.Database.Client().Disconnect(context.TODO())
	if err != nil {
		log.Print(err) // Handle potential disconnect error
	}
}
