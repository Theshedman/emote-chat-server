package controller

import (
	"chat-server/dto"
	"chat-server/repository"
	"context"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"log"
	"net/http"
	"strconv"
	"time"
)

func GetUsers(c echo.Context) error {
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

	userRepository := repository.NewUser()
	users, err := userRepository.Find(ctx, bson.M{}, page, limit, "created_at")
	if err != nil {
		log.Println("No user available")

		return echo.ErrNotFound
	}

	return c.JSON(http.StatusOK, echo.Map{
		"data": dto.ToUserListDto(users),
	})
}
