package websocket

import (
	"chat-server/auth"
	"chat-server/dto"
	"chat-server/rabbitmq"
	"chat-server/repository"
	"context"
	"encoding/json"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log"
	"time"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Client struct {
	Conn     *websocket.Conn
	Send     chan []byte // Channel for sending messages to the client
	UserID   primitive.ObjectID
	Username string
	Rooms    map[string]bool
	RabbitMQ *rabbitmq.RabbitMQ // RabbitMQ instance
	Handler  *SocketHandler     // Reference to the SocketHandler
}

type SocketHandler struct {
	clients  map[string]*Client // Client map (using user IDs as keys)
	rabbitMQ *rabbitmq.RabbitMQ
}

func New(rabbitMQ *rabbitmq.RabbitMQ) *SocketHandler {
	return &SocketHandler{
		clients:  make(map[string]*Client),
		rabbitMQ: rabbitMQ,
	}
}

func (sh *SocketHandler) HandleConnection(c echo.Context) error {
	conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		log.Println("Error upgrading to WebSocket: ", err)
		return err
	}

	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(*auth.JwtCustomClaims)
	username := claims.UserName
	userID, err := primitive.ObjectIDFromHex(claims.Subject)
	if err != nil {
		badRequest := echo.ErrBadRequest
		badRequest.Message = "invalid userID"

		return badRequest
	}

	client := &Client{
		Conn:     conn,
		Send:     make(chan []byte),
		UserID:   userID,
		Username: username,
		Rooms:    make(map[string]bool),
		RabbitMQ: sh.rabbitMQ, // Inject RabbitMQ instance
		Handler:  sh,
	}
	sh.clients[userID.Hex()] = client

	go client.readLoop()
	go client.writeLoop()

	return nil
}

func (c *Client) readLoop() {
	defer func() {
		// Clean up: Remove from client map, Close connection, Leave Rooms...
		delete(c.Handler.clients, c.UserID.Hex()) // Remove client from map

		for roomID := range c.Rooms {
			leaveRoom(roomID, c) // Leave each room the client is in
		}

		err := c.Conn.Close()
		if err != nil {
			fmt.Println("Error closing WebSocket connection:", err)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var cx context.Context
	var terminate context.CancelFunc
	for {
		var msg dto.MessageDto
		err := c.Conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				fmt.Println("Client disconnected gracefully")
			} else {
				fmt.Println("Client disconnected due to error:", err)
			}
			break
		}

		msgModel, err := dto.ToMessageModel(msg)
		if err != nil {
			log.Println(err)
			break
		}

		cx, terminate = context.WithTimeout(context.Background(), 5*time.Second)
		msgRepository := repository.NewMessage()

		msgModel.SenderID = c.UserID
		msgModel.Username = c.Username
		newMsg, err := msgRepository.Create(cx, msgModel)
		if err != nil {
			log.Println("failed to persist message: " + err.Error())
		}

		msgModelJson, err := json.Marshal(*newMsg)
		if err != nil {
			log.Println("failed to marshal message to json")
			break
		}

		err = c.RabbitMQ.Publish(ctx, rabbitmq.ExchangeName, msg.RoomID, msgModelJson)
		if err != nil {
			log.Println(err)
			break
		}
	}
	defer terminate()
}

func (c *Client) writeLoop() {
	for message := range c.Send {
		err := c.Conn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			fmt.Println("Error writing message to WebSocket:", err)
			break
		}
	}
}

func (sh *SocketHandler) ConsumeMessages(exchange, queueName string) {
	deliveries, err := sh.rabbitMQ.Consume(exchange, queueName)
	if err != nil {
		log.Printf("Error consuming messages from RabbitMQ: %v\n", err)
	}

	for delivery := range deliveries {
		for _, client := range sh.clients {
			select {
			case client.Send <- delivery.Body:
			default:
				fmt.Println("Client's message buffer is full. Skipping message.")
			}
		}
	}
}

func leaveRoom(roomID string, client *Client) {
	// Check if the client is in the specified room
	_, isInRoom := client.Rooms[roomID]
	if !isInRoom {
		// Client is not in the specified room, nothing to do
		return
	}

	// Remove the room from the client's rooms
	delete(client.Rooms, roomID)
}