package controller

import (
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

func GetUserById(c echo.Context) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	userIdString := c.Param("userId")
	userId, err := primitive.ObjectIDFromHex(userIdString)
	if err != nil {
		badRequest := echo.ErrBadRequest
		badRequest.Message = "invalid userId: " + userIdString

		return badRequest
	}

	userRepo := repository.NewUser()
	user, err := userRepo.FindOne(ctx, bson.M{"_id": userId})
	if err != nil {
		log.Println("no user exist with the provided id")

		return echo.ErrNotFound
	}

	return c.JSON(http.StatusOK, echo.Map{
		"data": dto.ToUserDto(*user),
	})
}
