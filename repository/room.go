package repository

import (
	"context"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log"
	"time"
)

const (
	PrivateChatRoom = "private"
	GroupChatRoom   = "group"
)

type RoomModel struct {
	ID           primitive.ObjectID   `bson:"_id"`
	Name         string               `bson:"name"`
	Type         string               `bson:"type"`
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

func (m *Model[T]) JoinPrivateChatRoom(currentUserId primitive.ObjectID, targetUserId primitive.ObjectID) (*RoomModel, error) {

	// Search for existing private room
	filter := bson.M{
		"type":         PrivateChatRoom,
		"participants": bson.M{"$size": 2, "$all": bson.A{currentUserId, targetUserId}},
	}

	roomRepo := NewRoom()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Validate only the targetUserId as the currentUserId is decoded from auth token
	userRepo := NewUser()
	targetUser, err := userRepo.FindById(ctx, targetUserId.Hex())
	if err != nil {
		badRequestError := echo.ErrBadRequest
		badRequestError.Message = "invalid targetParticipantId received"

		return nil, badRequestError
	}
	if (*targetUser).ID == currentUserId {
		badRequestError := echo.ErrBadRequest
		badRequestError.Message = "targetParticipantId cannot be the same as the current user id"

		return nil, badRequestError
	}

	existingRoom, err := roomRepo.FindOne(ctx, filter)
	if err != nil {
		// Room not found, create a new one
		newChatRoom, err := roomRepo.Create(ctx, &RoomModel{
			Type:         PrivateChatRoom,
			Participants: []primitive.ObjectID{currentUserId, targetUserId},
		})
		if err != nil {
			serverError := echo.ErrInternalServerError
			serverError.Message = "failed to join room: " + err.Error()

			return nil, serverError
		}

		return *newChatRoom, nil
	} else {
		return *existingRoom, nil
	}
}

func (m *Model[T]) JoinGroupChatRoom(currentUserId primitive.ObjectID, roomName string) (*RoomModel, error) {

	roomRepo := NewRoom()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	existingRoomReference, err := roomRepo.FindOne(ctx, bson.M{"name": roomName})
	if err != nil {
		log.Printf("room '%s' does not exist. creating one...", roomName)
	}

	if existingRoomReference != nil {
		existingRoom := *existingRoomReference

		// a user should not join a room twice
		for _, v := range existingRoom.Participants {
			if v == currentUserId {
				return existingRoom, nil
			}
		}

		existingRoom.Participants = append(existingRoom.Participants, currentUserId)
		updatedRoom, err := roomRepo.Update(ctx, existingRoom.ID.Hex(), existingRoom)
		if err != nil {
			serverError := echo.ErrInternalServerError
			serverError.Message = "failed to join room: " + err.Error()

			return nil, serverError
		}

		return *updatedRoom, nil
	}

	var participants []primitive.ObjectID
	participants = make([]primitive.ObjectID, 1)
	participants[0] = currentUserId

	newRoom, err := roomRepo.Create(ctx, &RoomModel{
		Name:         roomName,
		Type:         GroupChatRoom,
		Participants: participants,
	})

	return *newRoom, nil
}
