package dbmodel

import "gorm.io/gorm"

func Models() []any {
	return []any{
		&Room{},
		&RoomUser{},
		&RoomAdmin{},
		&RoomParticipant{},
		&RoomCard{},
		&RoomCardCell{},
		&RoomPickedBall{},
		&RoomBingoRecord{},
		&RoomReachRecord{},
		&RoomMessage{},
	}
}

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(Models()...)
}
