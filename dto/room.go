package dto

import (
	"chat-server/repository"
	"fmt"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Room struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Type         string   `json:"type"`
	Participants []string `json:"participants"`
}

func ToRoomListDto(roomModel []**repository.RoomModel) []Room {
	rooms := make([]Room, len(roomModel))

	var participants []string
	for i, roomReference := range roomModel {
		room := *roomReference

		rooms[i] = Room{
			ID:   room.ID.Hex(),
			Type: room.Type,
			Name: room.Name,
		}

		for _, v := range room.Participants {
			participants = append(participants, v.Hex())
		}

		rooms[i].Participants = participants
	}

	return rooms

}

func ToRoomDto(roomModel *repository.RoomModel) Room {
	room := *roomModel

	var participants []string
	for _, v := range room.Participants {
		participants = append(participants, v.Hex())
	}

	return Room{
		ID:           room.ID.Hex(),
		Name:         room.Name,
		Type:         room.Type,
		Participants: participants,
	}

}

func ToRoomModel(room Room) (*repository.RoomModel, error) {

	var participants []primitive.ObjectID

	for i, v := range room.Participants {
		objID, err := primitive.ObjectIDFromHex(v)
		if err != nil {
			return nil, fmt.Errorf("invalid roomId '%s': %w", v, err)
		}

		participants[i] = objID
	}

	return &repository.RoomModel{
		Name:         room.Name,
		Type:         room.Type,
		Participants: participants,
	}, nil
}
