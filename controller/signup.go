package controller

import (
	"chat-server/auth"
	"chat-server/dto"
	"chat-server/password"
	"chat-server/repository"
	"context"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"log"
	"net/http"
	"os"
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

	existingUser, err := userRepository.FindOne(ctx, bson.M{"user_name": modelData.Username})
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

	// Set custom claims
	claims := &auth.JwtCustomClaims{
		UserName: (*newUser).Username,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   (*newUser).ID.Hex(),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 72)),
		},
	}

	// Create token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Generate encoded token and send it as response.
	jwtSecret := os.Getenv("JWT_SECRET")
	authToken, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, echo.Map{
		"data":  dto.ToUserDto(modelData),
		"token": authToken,
	})
}
