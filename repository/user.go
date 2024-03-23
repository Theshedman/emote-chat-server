package repository

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type UserModel struct {
	ID        primitive.ObjectID `bson:"_id"`
	FirstName string             `bson:"first_name"`
	LastName  string             `bson:"last_name"`
	Username  string             `bson:"username"`
	Password  string             `bson:"password"`
	CreatedAt time.Time          `bson:"created_at"`
}

func NewUser() *Model[*UserModel] {
	userCollection := Database.Collection("users")

	return newModel[*UserModel](userCollection)
}

func (um *UserModel) GetID() primitive.ObjectID {
	return um.ID
}

func (um *UserModel) SetID(id primitive.ObjectID) {
	um.ID = id
}

func (um *UserModel) SetTimestamp() {
	um.CreatedAt = time.Now()
}
