package controller

import (
	"chat-server/repository"
	"context"
	"github.com/labstack/echo/v4"
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

	page, err := strconv.Atoi(pageString)
	if err != nil {
		page = 1
	}

	limit, err := strconv.Atoi(limitString)
	if err != nil {
		limit = 10
	}

	messageRepository := repository.NewMessage()
	messages, err := messageRepository.Find(ctx, page, limit)
	if err != nil {
		log.Println("no room exist")

		return echo.ErrNotFound
	}

	return c.JSON(http.StatusOK, echo.Map{
		"data": messages,
	})
}
