package websocket

import (
	"chat-server/auth"
	"chat-server/dto"
	"chat-server/rabbitmq"
	"chat-server/repository"
	"context"
	"encoding/json"
	"fmt"
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
		return fmt.Errorf("error upgrading to webSocket: %w", err)
	}

	principal := auth.GetPrincipal(c)
	userID, err := primitive.ObjectIDFromHex(principal.ID)
	if err != nil {
		badRequest := echo.ErrBadRequest
		badRequest.Message = "invalid userID"

		return badRequest
	}

	client := &Client{
		Conn:     conn,
		Send:     make(chan []byte),
		UserID:   userID,
		Username: principal.Username,
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
		// Clean up: Remove from a client map, Close connection, Leave Rooms...
		delete(c.Handler.clients, c.UserID.Hex()) // Remove a client from the map

		for roomID := range c.Rooms {
			leaveRoom(roomID, c) // Leave each room the client is in
		}

		err := c.Conn.Close()
		if err != nil {
			fmt.Println("Error closing WebSocket connection: ", err)
		}
	}()

	// Infinitely read messages from the clients
	// and publish them to a message broker (RabbitMQ).
	// This makes room for scalability as it prevents message loss
	// even when there's network/server unavailability, and which,
	// in turn, reduces the loads on the server
	var ctx context.Context
	var cancel context.CancelFunc
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
			log.Println("could not parse incoming message: ", err)
		}

		// Persist message to DB
		msgRepository := repository.NewMessage()
		ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
		msgModel.SenderID = c.UserID
		msgModel.Username = c.Username

		newMsg, err := msgRepository.Create(ctx, msgModel)
		if err != nil {
			log.Println("failed to persist message to DB: ", err)

			break
		}

		msgModelJson, err := json.Marshal(dto.ToMessageDto(*newMsg))
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
	defer cancel()
}

func (c *Client) writeLoop() {
	// Infinitely read messages from the queue,
	// persist them to the DB and then send them to the clients.
	for message := range c.Send {
		// Send messages to clients
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
