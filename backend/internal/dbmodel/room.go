package dbmodel

import "time"

type Room struct {
	RoomID              string    `gorm:"column:room_id;type:char(36);primaryKey;index:idx_rooms_room_code_created_at,priority:3;index:idx_rooms_created_at,priority:2"`
	RoomCode            string    `gorm:"column:room_code;type:char(6);not null;index:idx_rooms_room_code_created_at,priority:1;check:chk_rooms_room_code,room_code REGEXP '^[0-9]{6}$'"`
	State               string    `gorm:"column:state;type:varchar(16);not null;check:chk_rooms_state_pick,state IN ('waiting','playing','finished') AND pick_state IN ('idle','picking','exhausted') AND (state <> 'waiting' OR pick_state = 'idle') AND (state <> 'finished' OR pick_state = 'idle')"`
	PickState           string    `gorm:"column:pick_state;type:varchar(16);not null"`
	QRCodeVisible       bool      `gorm:"column:qr_code_visible;not null;default:true"`
	SettingsName        string    `gorm:"column:settings_name;type:varchar(255);not null"`
	SettingsDescription string    `gorm:"column:settings_description;type:text;not null"`
	CreatedAt           time.Time `gorm:"column:created_at;type:datetime(6);not null;default:CURRENT_TIMESTAMP(6);index:idx_rooms_room_code_created_at,priority:2;index:idx_rooms_created_at,priority:1"`
	UpdatedAt           time.Time `gorm:"column:updated_at;type:datetime(6);not null;default:CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6)"`
}

func (Room) TableName() string {
	return "rooms"
}

type RoomUser struct {
	RoomID    string    `gorm:"column:room_id;type:char(36);primaryKey;index:idx_room_users_user_id,priority:2"`
	UserID    string    `gorm:"column:user_id;type:varchar(64);primaryKey;index:idx_room_users_user_id,priority:1;check:chk_room_users_user_id,CHAR_LENGTH(user_id) > 0"`
	UserName  string    `gorm:"column:user_name;type:varchar(255);not null;check:chk_room_users_user_name,CHAR_LENGTH(user_name) > 0"`
	CreatedAt time.Time `gorm:"column:created_at;type:datetime(6);not null;default:CURRENT_TIMESTAMP(6)"`
	UpdatedAt time.Time `gorm:"column:updated_at;type:datetime(6);not null;default:CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6)"`
}

func (RoomUser) TableName() string {
	return "room_users"
}

type RoomAdmin struct {
	RoomID  string    `gorm:"column:room_id;type:char(36);primaryKey;index:idx_room_admins_user_id,priority:2"`
	UserID  string    `gorm:"column:user_id;type:varchar(64);primaryKey;index:idx_room_admins_user_id,priority:1"`
	AddedAt time.Time `gorm:"column:added_at;type:datetime(6);not null;default:CURRENT_TIMESTAMP(6)"`
}

func (RoomAdmin) TableName() string {
	return "room_admins"
}

type RoomParticipant struct {
	RoomID   string    `gorm:"column:room_id;type:char(36);primaryKey;index:idx_room_participants_user_id,priority:2"`
	UserID   string    `gorm:"column:user_id;type:varchar(64);primaryKey;index:idx_room_participants_user_id,priority:1"`
	JoinedAt time.Time `gorm:"column:joined_at;type:datetime(6);not null;default:CURRENT_TIMESTAMP(6)"`
}

func (RoomParticipant) TableName() string {
	return "room_participants"
}

type RoomCard struct {
	CardID      string    `gorm:"column:card_id;type:char(36);primaryKey;uniqueIndex:uq_room_cards_card_room_owner,priority:1"`
	RoomID      string    `gorm:"column:room_id;type:char(36);not null;uniqueIndex:uq_room_cards_room_owner,priority:1;uniqueIndex:uq_room_cards_room_card_number,priority:1;uniqueIndex:uq_room_cards_card_room_owner,priority:2;index:idx_room_cards_room,priority:1"`
	CardNumber  string    `gorm:"column:card_number;type:char(36);not null;uniqueIndex:uq_room_cards_room_card_number,priority:2;check:chk_room_cards_card_number,card_number REGEXP '^[0-9]{36}$'"`
	OwnerUserID string    `gorm:"column:owner_user_id;type:varchar(64);not null;uniqueIndex:uq_room_cards_room_owner,priority:2;uniqueIndex:uq_room_cards_card_room_owner,priority:3"`
	CreatedAt   time.Time `gorm:"column:created_at;type:datetime(6);not null;default:CURRENT_TIMESTAMP(6)"`
}

func (RoomCard) TableName() string {
	return "room_cards"
}

type RoomCardCell struct {
	CardID    string `gorm:"column:card_id;type:char(36);primaryKey;uniqueIndex:uq_room_card_cells_card_number,priority:1"`
	CellIndex uint8  `gorm:"column:cell_index;type:tinyint unsigned;primaryKey"`
	Number    *uint8 `gorm:"column:number;type:tinyint unsigned;uniqueIndex:uq_room_card_cells_card_number,priority:2;index:idx_room_card_cells_number,priority:1"`
	CellState string `gorm:"column:cell_state;type:varchar(16);not null;check:chk_room_card_cells_state,cell_index BETWEEN 0 AND 24 AND (number IS NULL OR number BETWEEN 1 AND 75) AND cell_state IN ('bingo','reach','open','closed') AND ((cell_index = 12 AND number IS NULL) OR (cell_index <> 12 AND number IS NOT NULL)) AND (number IS NOT NULL OR cell_state <> 'closed')"`
}

func (RoomCardCell) TableName() string {
	return "room_card_cells"
}

type RoomPickedBall struct {
	PickedBallID string    `gorm:"column:picked_ball_id;type:char(36);primaryKey"`
	RoomID       string    `gorm:"column:room_id;type:char(36);not null;uniqueIndex:uq_room_picked_balls_room_order,priority:1;uniqueIndex:uq_room_picked_balls_room_number,priority:1;index:idx_room_picked_balls_room_picked_at,priority:1"`
	PickOrder    uint16    `gorm:"column:pick_order;type:smallint unsigned;not null;uniqueIndex:uq_room_picked_balls_room_order,priority:2;check:chk_room_picked_balls_order,pick_order BETWEEN 1 AND 75"`
	Number       uint8     `gorm:"column:number;type:tinyint unsigned;not null;uniqueIndex:uq_room_picked_balls_room_number,priority:2;check:chk_room_picked_balls_number,number BETWEEN 1 AND 75"`
	PickedAt     time.Time `gorm:"column:picked_at;type:datetime(6);not null;default:CURRENT_TIMESTAMP(6);index:idx_room_picked_balls_room_picked_at,priority:2"`
}

func (RoomPickedBall) TableName() string {
	return "room_picked_balls"
}

type RoomBingoRecord struct {
	RecordID   string    `gorm:"column:record_id;type:char(36);primaryKey"`
	RoomID     string    `gorm:"column:room_id;type:char(36);not null;uniqueIndex:uq_room_bingo_records_room_order,priority:1;uniqueIndex:uq_room_bingo_records_room_user_line,priority:1;index:idx_room_bingo_records_card_room_user,priority:2;index:idx_room_bingo_records_room_user_order,priority:1"`
	UserID     string    `gorm:"column:user_id;type:varchar(64);not null;uniqueIndex:uq_room_bingo_records_room_user_line,priority:2;index:idx_room_bingo_records_card_room_user,priority:3;index:idx_room_bingo_records_room_user_order,priority:2"`
	CardID     string    `gorm:"column:card_id;type:char(36);not null;index:idx_room_bingo_records_card_room_user,priority:1"`
	LineIndex  uint8     `gorm:"column:line_index;type:tinyint unsigned;not null;uniqueIndex:uq_room_bingo_records_room_user_line,priority:3;check:chk_room_bingo_records_line,line_index BETWEEN 0 AND 11"`
	BingoOrder uint      `gorm:"column:bingo_order;type:int unsigned;not null;uniqueIndex:uq_room_bingo_records_room_order,priority:2;index:idx_room_bingo_records_room_user_order,priority:3;check:chk_room_bingo_records_order,bingo_order >= 1"`
	CreatedAt  time.Time `gorm:"column:created_at;type:datetime(6);not null;default:CURRENT_TIMESTAMP(6)"`
}

func (RoomBingoRecord) TableName() string {
	return "room_bingo_records"
}

type RoomReachRecord struct {
	RecordID      string    `gorm:"column:record_id;type:char(36);primaryKey"`
	RoomID        string    `gorm:"column:room_id;type:char(36);not null;uniqueIndex:uq_room_reach_records_room_user,priority:1;index:idx_room_reach_records_room_created,priority:1"`
	UserID        string    `gorm:"column:user_id;type:varchar(64);not null;uniqueIndex:uq_room_reach_records_room_user,priority:2"`
	LineIndex     uint8     `gorm:"column:line_index;type:tinyint unsigned;not null;check:chk_room_reach_records_line,line_index BETWEEN 0 AND 11"`
	LastCellIndex uint8     `gorm:"column:last_cell_index;type:tinyint unsigned;not null;check:chk_room_reach_records_last_cell,last_cell_index BETWEEN 0 AND 24"`
	CreatedAt     time.Time `gorm:"column:created_at;type:datetime(6);not null;default:CURRENT_TIMESTAMP(6);index:idx_room_reach_records_room_created,priority:2"`
}

func (RoomReachRecord) TableName() string {
	return "room_reach_records"
}

type RoomMessage struct {
	MessageID    string    `gorm:"column:message_id;type:char(36);primaryKey"`
	RoomID       string    `gorm:"column:room_id;type:char(36);not null;index:idx_room_messages_room_created,priority:1;index:idx_room_messages_room_author,priority:1;index:idx_room_messages_author,priority:2"`
	AuthorUserID string    `gorm:"column:author_user_id;type:varchar(64);not null;index:idx_room_messages_room_author,priority:2;index:idx_room_messages_author,priority:1"`
	Content      string    `gorm:"column:content;type:varchar(500);not null;check:chk_room_messages_content,CHAR_LENGTH(content) > 0"`
	CreatedAt    time.Time `gorm:"column:created_at;type:datetime(6);not null;default:CURRENT_TIMESTAMP(6);index:idx_room_messages_room_created,priority:2"`
}

func (RoomMessage) TableName() string {
	return "room_messages"
}
