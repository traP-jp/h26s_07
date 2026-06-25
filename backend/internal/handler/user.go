package handler

import (
	"net/http"

	"github.com/labstack/echo/v5"

	authmiddleware "github.com/traP-jp/h26_07/backend/internal/middleware"
	"github.com/traP-jp/h26_07/backend/internal/openapi"
)

type UserHandler struct{}

func NewUserHandler() *UserHandler {
	return &UserHandler{}
}

func (h *UserHandler) GetMe(c *echo.Context) error {
	user, ok := authmiddleware.GetAuthenticatedUser(c)
	if !ok {
		return c.NoContent(http.StatusUnauthorized)
	}

	return c.JSON(http.StatusOK, openapi.User{
		ID:   user.Name,
		Name: user.Name,
	})
}
