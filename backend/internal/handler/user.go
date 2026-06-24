package handler

import (
	"net/http"

	"github.com/labstack/echo/v5"

	authmiddleware "github.com/traP-jp/h26_07/backend/internal/middleware"
)

type UserHandler struct{}

type UserResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func NewUserHandler() *UserHandler {
	return &UserHandler{}
}

func (h *UserHandler) GetMe(c *echo.Context) error {
	user, ok := authmiddleware.GetAuthenticatedUser(c)
	if !ok {
		return c.NoContent(http.StatusUnauthorized)
	}

	return c.JSON(http.StatusOK, UserResponse{
		ID:   user.Name,
		Name: user.Name,
	})
}
