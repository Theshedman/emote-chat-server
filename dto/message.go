package dto

import (
	"chat-server/repository"
	"fmt"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type MessageDto struct {
	RoomID  string `json:"roomId"`
	Content string `json:"content"`
}

type Message struct {
	ID        string    `json:"id"`
	RoomID    string    `json:"roomId"`
	SenderID  string    `json:"senderId"`
	Content   string    `json:"content"`
	Username  string    `json:"username"`
	Timestamp time.Time `json:"timestamp"`
}

func ToMessageDto(messageModel *repository.MessageModel) Message {
	message := *messageModel

	return Message{
		ID:        message.ID.Hex(),
		RoomID:    message.RoomID.Hex(),
		SenderID:  message.SenderID.Hex(),
		Content:   message.Content,
		Username:  message.Username,
		Timestamp: message.Timestamp,
	}

}

func ToMessageModel(message MessageDto) (*repository.MessageModel, error) {
	roomID, err := primitive.ObjectIDFromHex(message.RoomID)
	if err != nil {
		return nil, fmt.Errorf("invalid roomId '%s': %w", message.RoomID, err)
	}

	return &repository.MessageModel{
		Content: message.Content,
		RoomID:  roomID,
	}, nil
}
