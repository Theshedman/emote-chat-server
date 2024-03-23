package repository

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type RoomModel struct {
	ID           primitive.ObjectID   `bson:"_id"`
	Name         string               `bson:"name"`
	Participants []primitive.ObjectID `bson:"participants"`
	CreatedAt    time.Time            `bson:"created_at"`
}

func NewRoom() *Model[*RoomModel] {
	userCollection := Database.Collection(Rooms)

	return newModel[*RoomModel](userCollection)
}

func (rm *RoomModel) GetID() primitive.ObjectID {
	return rm.ID
}

func (rm *RoomModel) SetID(id primitive.ObjectID) {
	rm.ID = id
}

func (rm *RoomModel) SetTimestamp() {
	rm.CreatedAt = time.Now()
}
