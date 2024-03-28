package controller

import (
	"github.com/labstack/echo/v4"
	"net/http"
)

func Health(c echo.Context) error {
	return c.JSON(http.StatusOK, echo.Map{
		"status": "Server is up and running!",
	})
}
