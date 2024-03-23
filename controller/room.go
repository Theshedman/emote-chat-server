package controller

import (
	"chat-server/auth"
	"chat-server/dto"
	"chat-server/repository"
	"context"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log"
	"net/http"
	"strconv"
	"time"
)

func Join(c echo.Context) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(*auth.JwtCustomClaims)
	participant, err := primitive.ObjectIDFromHex(claims.Subject)
	if err != nil {
		badRequest := echo.ErrBadRequest
		badRequest.Message = err

		return badRequest
	}

	roomName := c.Param("roomName")
	roomRepository := repository.NewRoom()

	existingRoom, err := roomRepository.FindOne(ctx, bson.M{"name": roomName})
	if err != nil {
		log.Printf("room '%s' does not exist. creating one...", roomName)
	}

	if existingRoom != nil {
		// a user should not join a room twice
		for _, v := range (*existingRoom).Participants {
			if v == participant {
				return c.JSON(http.StatusOK, echo.Map{
					"data": dto.ToRoomDto(*existingRoom),
				})
			}
		}

		(*existingRoom).Participants = append((*existingRoom).Participants, participant)
		updateRoom, err := roomRepository.Update(ctx, (*existingRoom).ID.Hex(), *existingRoom)
		if err != nil {
			serverError := echo.ErrInternalServerError
			serverError.Message = "failed to join room: " + err.Error()

			return serverError
		}

		return c.JSON(http.StatusOK, echo.Map{
			"data": dto.ToRoomDto(*updateRoom),
		})
	}

	var participants []primitive.ObjectID
	participants = make([]primitive.ObjectID, 1)
	participants[0] = participant

	newRoom, err := roomRepository.Create(ctx, &repository.RoomModel{
		Name:         roomName,
		Participants: participants,
	})
	if err != nil {
		serverError := echo.ErrInternalServerError
		serverError.Message = "failed to join room: " + err.Error()

		return serverError
	}

	return c.JSON(http.StatusOK, echo.Map{
		"data": dto.ToRoomDto(*newRoom),
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
	rooms, err := roomRepository.Find(ctx, page, limit)
	if err != nil {
		log.Println("no room exist")

		return echo.ErrNotFound
	}

	return c.JSON(http.StatusOK, echo.Map{
		"data": rooms,
	})
}
