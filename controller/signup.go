package controller

import (
	"chat-server/auth"
	"chat-server/dto"
	"chat-server/password"
	"chat-server/repository"
	"context"
	"fmt"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"log"
	"net/http"
	"time"
)

func Signup(c echo.Context) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	signupData := new(dto.NewUser)
	if err := c.Bind(signupData); err != nil {
		return err
	}

	userRepository := repository.NewUser()
	modelData, err := dto.ToUserModel(*signupData)
	if err != nil {
		return err
	}

	existingUser, err := userRepository.FindOne(ctx, bson.M{"username": modelData.Username})
	if err != nil {
		log.Println("username is free.")
	}

	if existingUser != nil {
		httpError := echo.ErrBadRequest

		httpError.Message = "user with the same username already exist"

		return httpError
	}

	hash, err := password.Hash(signupData.Password)
	if err != nil {
		return err
	}

	modelData.Password = hash
	newUser, err := userRepository.Create(ctx, modelData)
	if err != nil {
		return err
	}

	authToken, err := auth.GenToken(*newUser)
	if err != nil {
		return fmt.Errorf("failed to generate auth token: %w", err)
	}

	return c.JSON(http.StatusOK, echo.Map{
		"data":  dto.ToUserDto(modelData),
		"token": authToken,
	})
}
