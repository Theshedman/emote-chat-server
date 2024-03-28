package controller

import (
	"chat-server/auth"
	"chat-server/dto"
	"chat-server/repository"
	"context"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log"
	"net/http"
	"strconv"
	"time"
)

func JoinRoom(c echo.Context) error {

	principal := auth.GetPrincipal(c)
	participant, err := primitive.ObjectIDFromHex(principal.ID)
	if err != nil {
		badRequest := echo.ErrBadRequest
		badRequest.Message = err

		return badRequest
	}

	roomName := c.Param("roomName")
	room, err := JoinARoom(roomName, participant)
	if err != nil {
		serverError := echo.ErrInternalServerError
		serverError.Message = "failed to join room: " + err.Error()

		return serverError
	}

	return c.JSON(http.StatusOK, echo.Map{
		"data": dto.ToRoomDto(*room),
	})
}

func JoinARoom(roomName string, participant primitive.ObjectID) (**repository.RoomModel, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	roomRepository := repository.NewRoom()

	existingRoom, err := roomRepository.FindOne(ctx, bson.M{"name": roomName})
	if err != nil {
		log.Printf("room '%s' does not exist. creating one...", roomName)
	}

	if existingRoom != nil {
		// a user should not join a room twice
		for _, v := range (*existingRoom).Participants {
			if v == participant {
				return existingRoom, nil
			}
		}

		(*existingRoom).Participants = append((*existingRoom).Participants, participant)
		updatedRoom, err := roomRepository.Update(ctx, (*existingRoom).ID.Hex(), *existingRoom)
		if err != nil {
			serverError := echo.ErrInternalServerError
			serverError.Message = "failed to join room: " + err.Error()

			return nil, serverError
		}

		return updatedRoom, nil
	}

	var participants []primitive.ObjectID
	participants = make([]primitive.ObjectID, 1)
	participants[0] = participant

	newRoom, err := roomRepository.Create(ctx, &repository.RoomModel{
		Name:         roomName,
		Participants: participants,
	})
	return newRoom, nil
}

func GetRoomById(c echo.Context) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	roomIdString := c.Param("roomId")
	roomId, err := primitive.ObjectIDFromHex(roomIdString)
	if err != nil {
		badRequest := echo.ErrBadRequest
		badRequest.Message = "invalid roomId: " + roomIdString

		return badRequest
	}

	roomRepository := repository.NewRoom()
	room, err := roomRepository.FindOne(ctx, bson.M{"_id": roomId})
	if err != nil {
		log.Println("no room exist")

		return echo.ErrNotFound
	}

	return c.JSON(http.StatusOK, echo.Map{
		"data": dto.ToRoomDto(*room),
	})
}

func GetRooms(c echo.Context) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pageString := c.QueryParam("page")
	limitString := c.QueryParam("limit")

	page, err := strconv.Atoi(pageString)
	if err != nil {
		page = 1
	}

	limit, err := strconv.Atoi(limitString)
	if err != nil {
		limit = 10
	}

	roomRepository := repository.NewRoom()
	rooms, err := roomRepository.Find(ctx, bson.M{}, page, limit, "created_at")
	if err != nil {
		log.Println("no room exist")

		return echo.ErrNotFound
	}

	return c.JSON(http.StatusOK, echo.Map{
		"data": dto.ToRoomListDto(rooms),
	})
}
