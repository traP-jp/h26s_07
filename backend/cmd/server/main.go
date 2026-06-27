package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v5"
	gormmysql "gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/traP-jp/h26_07/backend/internal/config"
	"github.com/traP-jp/h26_07/backend/internal/server"
)

func main() {
	cfg := config.Load()
	dsn := cfg.Database.DSN()
	if dsn == "" {
		log.Fatal("database connection is required: set DATABASE_URL or DB_HOST/DB_PORT/DB_USER/DB_PASSWORD/DB_NAME")
	}

	db, err := gorm.Open(gormmysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("failed to get database handle: %v", err)
	}
	defer sqlDB.Close()
	if err := sqlDB.Ping(); err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	e := server.New(cfg, db)

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
