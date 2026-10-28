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

	db, err := gorm.Open(sqlite.Open("file:"+uuid.NewString()+"?mode=memory&cache=shared&_busy_timeout=10000"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatalf("failed to open sqlite database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	statements := []string{
		`CREATE TABLE rooms (
			room_id TEXT PRIMARY KEY,
			room_code TEXT NOT NULL,
			state TEXT NOT NULL,
			pick_state TEXT NOT NULL,
			qr_code_visible BOOLEAN NOT NULL,
			settings_name TEXT NOT NULL,
			settings_description TEXT NOT NULL,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL
		)`,
		`CREATE TABLE room_users (
			room_id TEXT NOT NULL,
			user_id TEXT NOT NULL,
			user_name TEXT NOT NULL,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL,
			PRIMARY KEY (room_id, user_id)
		)`,
		`CREATE TABLE room_admins (
			room_id TEXT NOT NULL,
			user_id TEXT NOT NULL,
			added_at DATETIME NOT NULL,
			PRIMARY KEY (room_id, user_id)
		)`,
		`CREATE TABLE room_participants (
			room_id TEXT NOT NULL,
			user_id TEXT NOT NULL,
			joined_at DATETIME NOT NULL,
			PRIMARY KEY (room_id, user_id)
		)`,
		`CREATE TABLE room_cards (
			card_id TEXT PRIMARY KEY,
			room_id TEXT NOT NULL,
			card_number TEXT NOT NULL,
			owner_user_id TEXT NOT NULL,
			created_at DATETIME NOT NULL,
			UNIQUE (room_id, owner_user_id),
			UNIQUE (room_id, card_number)
		)`,
		`CREATE TABLE room_card_cells (
			card_id TEXT NOT NULL,
			cell_index INTEGER NOT NULL,
			number INTEGER,
			cell_state TEXT NOT NULL,
			PRIMARY KEY (card_id, cell_index),
			UNIQUE (card_id, number)
		)`,
		`CREATE TABLE room_picked_balls (
			picked_ball_id TEXT PRIMARY KEY,
			room_id TEXT NOT NULL,
			pick_order INTEGER NOT NULL,
			number INTEGER NOT NULL,
			picked_at DATETIME NOT NULL,
			UNIQUE (room_id, pick_order),
			UNIQUE (room_id, number)
		)`,
		`CREATE TABLE room_bingo_records (
			record_id TEXT PRIMARY KEY,
			room_id TEXT NOT NULL,
			user_id TEXT NOT NULL,
			card_id TEXT NOT NULL,
			line_index INTEGER NOT NULL,
			bingo_order INTEGER NOT NULL,
			created_at DATETIME NOT NULL,
			UNIQUE (room_id, bingo_order),
			UNIQUE (room_id, user_id, line_index)
		)`,
		`CREATE TABLE room_reach_records (
			record_id TEXT PRIMARY KEY,
			room_id TEXT NOT NULL,
			user_id TEXT NOT NULL,
			line_index INTEGER NOT NULL,
			last_cell_index INTEGER NOT NULL,
			created_at DATETIME NOT NULL,
			UNIQUE (room_id, user_id)
		)`,
		`CREATE TABLE room_messages (
			message_id TEXT PRIMARY KEY,
			room_id TEXT NOT NULL,
			author_user_id TEXT NOT NULL,
			content TEXT NOT NULL,
			created_at DATETIME NOT NULL
		)`,
	}

	for _, statement := range statements {
		if err := db.Exec(statement).Error; err != nil {
			t.Fatalf("failed to create test schema: %v", err)
		}
	}
	return db
}
