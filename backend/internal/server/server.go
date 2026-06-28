package server

import (
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"gorm.io/gorm"

	"github.com/traP-jp/h26_07/backend/internal/config"
	"github.com/traP-jp/h26_07/backend/internal/handler"
	authmiddleware "github.com/traP-jp/h26_07/backend/internal/middleware"
	"github.com/traP-jp/h26_07/backend/internal/repository"
	"github.com/traP-jp/h26_07/backend/internal/service"
)

func New(cfg config.Config, db *gorm.DB) *echo.Echo {
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

	registerRoutes(e, cfg, db)

	return e
}

func registerRoutes(e *echo.Echo, cfg config.Config, db *gorm.DB) {
	healthHandler := handler.NewHealthHandler()
	userHandler := handler.NewUserHandler()

	e.GET("/healthz", healthHandler.Get)

	api := e.Group("/api")
	api.Use(authmiddleware.ForwardedUser)
	api.GET("/me", userHandler.GetMe)

	if db != nil {
		transactionRunner := repository.NewGormTransactionRunner(db)
		roomRepository := repository.NewGormRoomRepository(db)
		webSocketHub := handler.NewWebSocketHub()
		eventSender := handler.NewWebSocketEventSender(webSocketHub)
		roomService := service.NewRoomService(transactionRunner, roomRepository, eventSender)
		roomHandler := handler.NewRoomHandler(roomService)
		roomWebSocketHandler := handler.NewRoomWebSocketHandler(roomService, webSocketHub, cfg.CORSAllowOrigins)
		api.POST("/rooms", roomHandler.PostRoom)
		api.GET("/rooms/:roomId", roomHandler.GetRoom)
		api.GET("/rooms", roomHandler.ListRooms)
		api.GET("/rooms/:roomId/ws", roomWebSocketHandler.Connect)
		api.POST("/rooms/:roomId/participants", roomHandler.PostParticipant)
		api.POST("/rooms/:roomId/chats", roomHandler.PostMessage)
		api.GET("/rooms/:roomId/chats", roomHandler.GetMessages)
		api.PUT("/rooms/:roomId/settings", roomHandler.PutSettings)
		api.POST("/rooms/:roomId/control/qrcode/show", roomHandler.ShowQRCode)
		api.POST("/rooms/:roomId/control/qrcode/hide", roomHandler.HideQRCode)
		api.POST("/rooms/:roomId/control/start", roomHandler.StartGame)
		api.POST("/rooms/:roomId/control/pick/start", roomHandler.PostPickStart)
		api.POST("/rooms/:roomId/control/pick/cancel", roomHandler.PostPickCancel)
		api.POST("/rooms/:roomId/control/finish", roomHandler.FinishGame)
	}
}
