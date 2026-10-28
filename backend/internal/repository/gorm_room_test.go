package repository

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/traP-jp/h26_07/backend/internal/model"
	"github.com/traP-jp/h26_07/backend/internal/testutil"
)

func TestGormRoomRepositorySaveFindAndList(t *testing.T) {
	ctx := context.Background()
	db := newRoomRepositoryTestDB(t)
	repo := NewGormRoomRepository(db)

	createdAt := time.Date(2026, 6, 27, 12, 0, 0, 0, time.UTC)
	roomID := mustRoomID("11111111-1111-1111-1111-111111111111")
	room := model.NewRoom(roomID, "123456", model.RoomSettings{
		Name:        "first game",
		Description: "description",
		Admins:      []model.UserID{"owner"},
	}, createdAt)

	if err := room.Join("alice", createdAt.Add(time.Minute)); err != nil {
		t.Fatalf("failed to join room: %v", err)
	}
	room.Cards = []model.Card{testCard("22222222-2222-2222-2222-222222222222", "alice")}
	room.PickedBalls = []model.BallNumber{1, 2}
	room.BingoRecords = []model.BingoRecord{{
		RecordID:  mustRecordID("33333333-3333-3333-3333-333333333333"),
		UserID:    "alice",
		Line:      0,
		Order:     1,
		CreatedAt: createdAt.Add(2 * time.Minute),
	}}
	room.ReachRecords = []model.ReachRecord{{
		RecordID:      mustRecordID("44444444-4444-4444-4444-444444444444"),
		UserID:        "alice",
		Line:          1,
		LastCellIndex: 4,
		CreatedAt:     createdAt.Add(3 * time.Minute),
	}, {
		RecordID:      mustRecordID("77777777-7777-7777-7777-777777777777"),
		UserID:        "bob",
		Line:          2,
		LastCellIndex: 9,
		CreatedAt:     createdAt.Add(4 * time.Minute),
	}}
	room.Messages = []model.Message{{
		MessageID: mustMessageID("55555555-5555-5555-5555-555555555555"),
		Content:   "hello",
		Author:    "alice",
		CreatedAt: createdAt.Add(5 * time.Minute),
	}}
	room.UpdatedAt = createdAt.Add(6 * time.Minute)

	if err := repo.Save(ctx, room); err != nil {
		t.Fatalf("failed to save room: %v", err)
	}

	summaries, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("failed to list room summaries: %v", err)
	}
	if len(summaries) != 1 {
		t.Fatalf("expected 1 room summary, got %d", len(summaries))
	}
	if summaries[0].RoomID != roomID || summaries[0].Settings.Name != "first game" {
		t.Fatalf("unexpected room summary identity: %+v", summaries[0])
	}
	if len(summaries[0].Participants) != 1 || summaries[0].Participants[0].UserID != "alice" {
		t.Fatalf("unexpected room summary participants: %+v", summaries[0].Participants)
	}
	if len(summaries[0].BingoSummaries) != 1 ||
		summaries[0].BingoSummaries[0].UserID != "alice" ||
		len(summaries[0].BingoSummaries[0].BingoOrders) != 1 ||
		summaries[0].BingoSummaries[0].BingoOrders[0] != 1 ||
		!summaries[0].BingoSummaries[0].CreatedAt.Equal(createdAt.Add(2*time.Minute)) {
		t.Fatalf("unexpected room summary bingos: %+v", summaries[0].BingoSummaries)
	}
	if len(summaries[0].ReachSummaries) != 1 ||
		summaries[0].ReachSummaries[0].UserID != "bob" ||
		!summaries[0].ReachSummaries[0].CreatedAt.Equal(createdAt.Add(4*time.Minute)) {
		t.Fatalf("unexpected room summary reaches: %+v", summaries[0].ReachSummaries)
	}

	loaded, err := repo.FindByID(ctx, roomID)
	if err != nil {
		t.Fatalf("failed to find room: %v", err)
	}

	if loaded.RoomID != roomID || loaded.RoomCode != "123456" {
		t.Fatalf("unexpected loaded room identity: %+v", loaded)
	}
	if loaded.Settings.Name != "first game" || len(loaded.Settings.Admins) != 1 || loaded.Settings.Admins[0] != "owner" {
		t.Fatalf("unexpected loaded settings: %+v", loaded.Settings)
	}
	if len(loaded.Participants) != 1 || loaded.Participants[0].UserID != "alice" {
		t.Fatalf("unexpected loaded participants: %+v", loaded.Participants)
	}
	if len(loaded.Cards) != 1 {
		t.Fatalf("expected 1 card, got %d", len(loaded.Cards))
	}
	if loaded.Cards[0].OwnerUserID != "alice" {
		t.Fatalf("unexpected loaded card owner: %+v", loaded.Cards[0])
	}
	if loaded.Cards[0].CardNumber != model.CardNumber("000000000000000000000000000000000222") {
		t.Fatalf("unexpected loaded card number: %+v", loaded.Cards[0])
	}
	if loaded.Cards[0].Cells[0].Number == nil || *loaded.Cards[0].Cells[0].Number != 1 {
		t.Fatalf("unexpected loaded first card cell: %+v", loaded.Cards[0].Cells[0])
	}
	if loaded.Cards[0].Cells[12].Number != nil || loaded.Cards[0].Cells[12].CellState != model.CardCellStateOpen {
		t.Fatalf("unexpected loaded free card cell: %+v", loaded.Cards[0].Cells[12])
	}
	if len(loaded.PickedBalls) != 2 || loaded.PickedBalls[0] != 1 || loaded.PickedBalls[1] != 2 {
		t.Fatalf("unexpected loaded picked balls: %+v", loaded.PickedBalls)
	}
	if len(loaded.BingoRecords) != 1 || loaded.BingoRecords[0].Order != 1 {
		t.Fatalf("unexpected loaded bingo records: %+v", loaded.BingoRecords)
	}
	if len(loaded.ReachRecords) != 2 ||
		loaded.ReachRecords[0].UserID != "alice" ||
		loaded.ReachRecords[0].LastCellIndex != 4 ||
		loaded.ReachRecords[1].UserID != "bob" ||
		loaded.ReachRecords[1].LastCellIndex != 9 {
		t.Fatalf("unexpected loaded reach records: %+v", loaded.ReachRecords)
	}
	if len(loaded.Messages) != 1 || loaded.Messages[0].Content != "hello" || loaded.Messages[0].Author != "alice" {
		t.Fatalf("unexpected loaded messages: %+v", loaded.Messages)
	}

	room.Settings.Name = "updated game"
	room.Participants = nil
	room.Cards = nil
	room.PickedBalls = []model.BallNumber{1}
	room.BingoRecords = nil
	room.ReachRecords = nil
	room.Messages = nil
	room.UpdatedAt = createdAt.Add(10 * time.Minute)
	if err := repo.Save(ctx, room); err != nil {
		t.Fatalf("failed to update room: %v", err)
	}

	updated, err := repo.FindByID(ctx, roomID)
	if err != nil {
		t.Fatalf("failed to find updated room: %v", err)
	}
	if updated.Settings.Name != "updated game" {
		t.Fatalf("unexpected updated settings: %+v", updated.Settings)
	}
	if len(updated.Participants) != 0 || len(updated.Cards) != 0 || len(updated.BingoRecords) != 0 || len(updated.ReachRecords) != 0 || len(updated.Messages) != 0 {
		t.Fatalf("expected replaced child rows to be empty: %+v", updated)
	}
	if len(updated.PickedBalls) != 1 || updated.PickedBalls[0] != 1 {
		t.Fatalf("unexpected updated picked balls: %+v", updated.PickedBalls)
	}

	second := model.NewRoom(
		mustRoomID("66666666-6666-6666-6666-666666666666"),
		"654321",
		model.RoomSettings{
			Name:        "second game",
			Description: "",
			Admins:      []model.UserID{"owner"},
		},
		createdAt.Add(time.Hour),
	)
	if err := repo.Save(ctx, second); err != nil {
		t.Fatalf("failed to save second room: %v", err)
	}

	summaries, err = repo.List(ctx)
	if err != nil {
		t.Fatalf("failed to list rooms: %v", err)
	}
	if len(summaries) != 2 {
		t.Fatalf("expected 2 rooms, got %d", len(summaries))
	}
	if summaries[0].RoomID != second.RoomID || summaries[1].RoomID != roomID {
		t.Fatalf("unexpected room order: %+v", summaries)
	}
}

func TestGormRoomRepositoryFindByIDNotFound(t *testing.T) {
	repo := NewGormRoomRepository(newRoomRepositoryTestDB(t))

	_, err := repo.FindByID(context.Background(), mustRoomID("77777777-7777-7777-7777-777777777777"))
	if !errors.Is(err, ErrRoomNotFound) {
		t.Fatalf("expected ErrRoomNotFound, got %v", err)
	}
}

func TestGormRoomRepositorySaveRejectsDuplicateCardNumbers(t *testing.T) {
	ctx := context.Background()
	repo := NewGormRoomRepository(newRoomRepositoryTestDB(t))
	now := time.Date(2026, 6, 27, 12, 0, 0, 0, time.UTC)
	room := model.NewRoom(mustRoomID("99999999-9999-9999-9999-999999999999"), "999999", model.RoomSettings{
		Name:        "duplicate card number game",
		Description: "",
		Admins:      []model.UserID{"owner"},
	}, now)

	if err := room.Join("alice", now); err != nil {
		t.Fatalf("failed to join alice: %v", err)
	}
	if err := room.Join("bob", now); err != nil {
		t.Fatalf("failed to join bob: %v", err)
	}
	room.Cards = []model.Card{
		testCard("22222222-2222-2222-2222-222222222222", "alice"),
		testCard("33333333-3333-3333-3333-333333333333", "bob"),
	}

	err := repo.Save(ctx, room)
	if !errors.Is(err, ErrInvalidRoomAggregate) {
		t.Fatalf("expected ErrInvalidRoomAggregate, got %v", err)
	}
}

func TestGormTransactionRunnerRollsBackRoomRepositoryChanges(t *testing.T) {
	ctx := context.Background()
	db := newRoomRepositoryTestDB(t)
	repo := NewGormRoomRepository(db)
	runner := NewGormTransactionRunner(db)

	room := model.NewRoom(
		mustRoomID("88888888-8888-8888-8888-888888888888"),
		"888888",
		model.RoomSettings{
			Name:        "rollback game",
			Description: "",
			Admins:      []model.UserID{"owner"},
		},
		time.Date(2026, 6, 27, 13, 0, 0, 0, time.UTC),
	)
	rollbackErr := errors.New("rollback")

	err := runner.WithinTransaction(ctx, func(ctx context.Context) error {
		if err := repo.Save(ctx, room); err != nil {
			return err
		}
		if _, err := repo.FindByID(ctx, room.RoomID); err != nil {
			return fmt.Errorf("expected room to be visible inside transaction: %w", err)
		}
		return rollbackErr
	})
	if !errors.Is(err, rollbackErr) {
		t.Fatalf("expected rollback error, got %v", err)
	}

	_, err = repo.FindByID(ctx, room.RoomID)
	if !errors.Is(err, ErrRoomNotFound) {
		t.Fatalf("expected rolled back room to be absent, got %v", err)
	}
}

func newRoomRepositoryTestDB(t *testing.T) *gorm.DB { return testutil.NewRoomDB(t) }

func testCard(cardID string, ownerUserID model.UserID) model.Card {
	var cells [25]model.CardCell
	for i := range cells {
		cell := model.CardCell{
			Index:     model.CellIndex(i),
			CellState: model.CardCellStateClosed,
		}
		if i == 12 {
			cell.CellState = model.CardCellStateOpen
		} else {
			number := model.BallNumber(i + 1)
			cell.Number = &number
		}
		cells[i] = cell
	}

	return model.Card{
		CardID:      model.CardID(uuid.MustParse(cardID)),
		CardNumber:  "000000000000000000000000000000000222",
		OwnerUserID: ownerUserID,
		Cells:       cells,
	}
}

func mustRoomID(value string) model.RoomID {
	return model.RoomID(uuid.MustParse(value))
}

func mustRecordID(value string) model.RecordID {
	return model.RecordID(uuid.MustParse(value))
}

func mustMessageID(value string) model.MessageID {
	return model.MessageID(uuid.MustParse(value))
}
