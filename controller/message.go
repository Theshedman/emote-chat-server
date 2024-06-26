package controller

import (
	"chat-server/dto"
	"chat-server/repository"
	"context"
	"fmt"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log"
	"net/http"
	"strconv"
	"time"
)

func GetMessages(c echo.Context) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pageString := c.QueryParam("page")
	limitString := c.QueryParam("limit")

	roomId := c.QueryParam("roomId")

	page, err := strconv.Atoi(pageString)
	if err != nil {
		page = 1
	}

	limit, err := strconv.Atoi(limitString)
	if err != nil {
		limit = 10
	}

	filter := bson.M{}
	if len(roomId) > 0 {
		objectID, err := primitive.ObjectIDFromHex(roomId)
		if err != nil {
			return fmt.Errorf("invalid channel id received: %w", err)
		}

		filter = bson.M{"room_id": objectID}
	}

	messageRepository := repository.NewMessage()
	messages, err := messageRepository.Find(ctx, filter, page, limit, "")
	if err != nil {
		log.Println("no room exist")

		return echo.ErrNotFound
	}

	return c.JSON(http.StatusOK, echo.Map{
		"data": dto.ToMessageListDto(messages),
	})
}
