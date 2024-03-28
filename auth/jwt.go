package auth

import (
	"chat-server/repository"
	"github.com/golang-jwt/jwt/v5"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"os"
	"time"
)

type JwtCustomClaims struct {
	UserName string `json:"userName"`
	jwt.RegisteredClaims
}

// JwtCustomConfig Configure middleware with the custom claims type
func JwtCustomConfig() echojwt.Config {
	jwtSecret := os.Getenv("JWT_SECRET")

	return echojwt.Config{
		NewClaimsFunc: func(c echo.Context) jwt.Claims {
			return new(JwtCustomClaims)
		},
		SigningKey:  []byte(jwtSecret),
		TokenLookup: "header:Authorization:Bearer ,query:auth",
	}
}

// GenToken generate jwt token
func GenToken(userModel *repository.UserModel) (string, error) {
	claims := &JwtCustomClaims{
		UserName: (*userModel).Username,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   (*userModel).ID.Hex(),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 5)),
		},
	}

	// Create token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Generate encoded token and send it as response.
	jwtSecret := os.Getenv("JWT_SECRET")
	authToken, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", err
	}

	return authToken, nil
}

func AuthenticateWebsocketConnection(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Request().Header.Del("Origin")

		return next(c)
	}
}
