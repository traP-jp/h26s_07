package server

import (
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"

	"github.com/traP-jp/h26_07/backend/internal/config"
	"github.com/traP-jp/h26_07/backend/internal/handler"
	authmiddleware "github.com/traP-jp/h26_07/backend/internal/middleware"
)

func New(cfg config.Config) *echo.Echo {
	e := echo.New()

	e.Use(middleware.Recover())
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogMethod:  true,
		LogURI:     true,
		LogStatus:  true,
		LogLatency: true,
		LogValuesFunc: func(c *echo.Context, v middleware.RequestLoggerValues) error {
			slog.InfoContext(
				c.Request().Context(),
				"request",
				slog.String("method", v.Method),
				slog.String("uri", v.URI),
				slog.Int("status", v.Status),
				slog.Duration("latency", v.Latency),
			)
			return nil
		},
	}))
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: cfg.CORSAllowOrigins,
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodOptions},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization, "X-Forwarded-User"},
	}))

	registerRoutes(e)

	return e
}

func registerRoutes(e *echo.Echo) {
	healthHandler := handler.NewHealthHandler()
	userHandler := handler.NewUserHandler()

	e.GET("/healthz", healthHandler.Get)

	api := e.Group("/api")
	api.Use(authmiddleware.ForwardedUser)
	api.GET("/me", userHandler.GetMe)
}
