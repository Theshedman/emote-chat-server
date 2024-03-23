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
	"net/http"
	"time"
)

func Login(c echo.Context) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	loginData := new(dto.UserLogin)
	if err := c.Bind(loginData); err != nil {
		return err
	}

	userRepository := repository.NewUser()

	filter := bson.M{"user_name": loginData.Username}
	userModel, err := userRepository.FindOne(ctx, filter)
	if err != nil {
		unAuthorizeErr := echo.ErrUnauthorized
		unAuthorizeErr.Message = "invalid username/password"

		return unAuthorizeErr
	}

	isValidPassword := password.Verify(loginData.Password, (*userModel).Password)
	if !isValidPassword {
		unAuthorizeErr := echo.ErrUnauthorized
		unAuthorizeErr.Message = "invalid username/password"

		return unAuthorizeErr
	}

	authToken, err := auth.GenToken(*userModel)
	if err != nil {
		return fmt.Errorf("failed to generate auth token: %w", err)
	}

	return c.JSON(http.StatusOK, echo.Map{
		"data":  dto.ToUserDto(*userModel),
		"token": authToken,
	})
}
