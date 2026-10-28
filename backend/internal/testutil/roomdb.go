// Package testutil provides the portable SQLite room schema used by backend tests.
package testutil

import (
	"testing"

	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// NewRoomDB omits MariaDB-specific defaults and checks from the production schema.
func NewRoomDB(t testing.TB) *gorm.DB {
	t.Helper()
	db := openRoomDB(t)
	createRoomDBSchema(t, db)
	return db
}

func openRoomDB(t testing.TB) *gorm.DB {
	t.Helper()
	dsn := "file:" + uuid.NewString() + "?mode=memory&cache=shared&_busy_timeout=10000"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatalf("failed to open sqlite database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	return db
}

func createRoomDBSchema(t testing.TB, db *gorm.DB) {
	t.Helper()
	for _, statement := range roomDBSchemaStatements() {
		if err := db.Exec(statement).Error; err != nil {
			t.Fatalf("failed to create test schema: %v", err)
		}
	}
}
