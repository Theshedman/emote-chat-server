package auth

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

type Principal struct {
	ID       string
	Username string
}

func GetPrincipal(c echo.Context) Principal {
	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(*JwtCustomClaims)

	return Principal{
		ID:       claims.Subject,
		Username: claims.UserName,
	}
}
