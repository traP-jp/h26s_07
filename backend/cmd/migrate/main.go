package main

import (
	"log"

	gormmysql "gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/traP-jp/h26_07/backend/internal/config"
	"github.com/traP-jp/h26_07/backend/internal/dbmodel"
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

	if err := dbmodel.AutoMigrate(db); err != nil {
		log.Fatalf("failed to auto migrate: %v", err)
	}
}
