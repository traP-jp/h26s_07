package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v5"

	"github.com/traP-jp/h26_07/backend/internal/config"
	"github.com/traP-jp/h26_07/backend/internal/server"
)

func main() {
	cfg := config.Load()
	e := server.New(cfg)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	startConfig := echo.StartConfig{
		Address:         cfg.Addr(),
		HideBanner:      true,
		HidePort:        true,
		GracefulTimeout: 10 * time.Second,
		OnShutdownError: func(err error) {
			slog.Error("failed to shutdown server", slog.Any("error", err))
		},
	}

	if err := startConfig.Start(ctx, e); err != nil {
		slog.Error("failed to start server", slog.Any("error", err))
		os.Exit(1)
	}
}
