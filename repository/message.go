package repository

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type MessageModel struct {
	ID        primitive.ObjectID `bson:"_id"`
	RoomID    primitive.ObjectID `bson:"room_id"`
	SenderID  primitive.ObjectID `bson:"sender_id"`
	Content   string             `bson:"content"`
	Username  string             `bson:"username"`
	Timestamp time.Time          `bson:"timestamp"`
}

func NewMessage() *Model[*MessageModel] {

	userCollection := Database.Collection("messages")

	return newModel[*MessageModel](userCollection)
}

func (mm *MessageModel) GetID() primitive.ObjectID {
	return mm.ID
}

func (mm *MessageModel) SetID(id primitive.ObjectID) {
	mm.ID = id
}

func (mm *MessageModel) SetTimestamp() {
	mm.Timestamp = time.Now()
}
