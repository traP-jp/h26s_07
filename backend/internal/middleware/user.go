package middleware

import (
	"strings"

	"github.com/labstack/echo/v5"
)

const userContextKey = "authenticatedUser"

type AuthenticatedUser struct {
	Name string
}

func ForwardedUser(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c *echo.Context) error {
		userID := strings.TrimSpace(c.Request().Header.Get("X-Forwarded-User"))
		if userID == "" {
			userID = "traP"
		}

		c.Set(userContextKey, AuthenticatedUser{
			Name: userID,
		})

		return next(c)
	}
}

func GetAuthenticatedUser(c *echo.Context) (AuthenticatedUser, bool) {
	user, ok := c.Get(userContextKey).(AuthenticatedUser)
	return user, ok
}
