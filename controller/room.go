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
	currentParticipant, err := primitive.ObjectIDFromHex(principal.ID)
	if err != nil {
		badRequest := echo.ErrBadRequest
		badRequest.Message = err

		return badRequest
	}

	roomType := c.Param("type")
	if roomType != repository.PrivateChatRoom && roomType != repository.GroupChatRoom {
		badRequest := echo.ErrBadRequest
		badRequest.Message = "invalid room type received: type must be either 'private' or 'group'"

		return badRequest
	}
	targetParticipantHex := c.QueryParam("targetParticipant")
	if roomType == repository.PrivateChatRoom && targetParticipantHex == "" {
		badRequest := echo.ErrBadRequest
		badRequest.Message = "missing 'targetParticipant' query param"

		return badRequest
	}

	var targetParticipant primitive.ObjectID
	if targetParticipantHex != "" {
		targetParticipant, err = primitive.ObjectIDFromHex(targetParticipantHex)
		if err != nil {
			badRequest := echo.ErrBadRequest
			badRequest.Message = "invalid targetParticipant received"

			return badRequest
		}
	}

	roomName := c.QueryParam("name")
	if roomType == repository.GroupChatRoom && roomName == "" {
		badRequest := echo.ErrBadRequest
		badRequest.Message = "room name is required for group chat"

		return badRequest
	}

	roomRepo := repository.NewRoom()
	var room *repository.RoomModel
	if roomType == repository.PrivateChatRoom {
		room, err = roomRepo.JoinPrivateChatRoom(currentParticipant, targetParticipant)
	} else if roomType == repository.GroupChatRoom {
		room, err = roomRepo.JoinGroupChatRoom(currentParticipant, roomName)
	}

	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, echo.Map{
		"data": dto.ToRoomDto(room),
	})
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
