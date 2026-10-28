package handler

import (
	"github.com/traP-jp/h26_07/backend/internal/model"
	"github.com/traP-jp/h26_07/backend/internal/openapi"
	"github.com/traP-jp/h26_07/backend/internal/utils"
)

func buildInitializedEvent(room *model.Room, mode openapi.WebSocketMode, userID model.UserID) (any, bool, error) {
	switch mode {
	case openapi.Participant:
		body := openapi.ParticipantInitializedBody{
			ActivityLog:      utils.ConvertActivityLogToOpenAPI(room.ActivityLog()),
			Participants:     utils.ParticipantUsers(room),
			ParticipantCount: len(room.Participants),
			State:            openapi.RoomState(room.State),
			Settings:         convertRoomSettingsToOpenAPI(room.Settings),
			PickState:        openapi.PickState(room.PickState),
			PickedBalls:      utils.ConvertPickedBallsToOpenAPI(room.PickedBalls),
			BingoSummaries:   convertBingoSummariesToOpenAPI(room.BingoSummaries()),
			ReachSummaries:   convertReachSummariesToOpenAPI(room.ReachSummaries()),
		}
		if room.State != model.RoomStateWaiting {
			if card, ok := utils.FindCard(room, userID); ok {
				body.Card = new(utils.ConvertCardToOpenAPI(room, card))
			}
		}
		return openapi.ParticipantInitializedEvent{
			Type: openapi.ParticipantInitializedEventTypeInitialized,
			Body: body,
		}, true, nil
	case openapi.Display:
		return openapi.DisplayInitializedEvent{
			Type: openapi.DisplayInitializedEventTypeInitialized,
			Body: openapi.DisplayInitializedBody{
				ActivityLog:      utils.ConvertActivityLogToOpenAPI(room.ActivityLog()),
				Participants:     utils.ParticipantUsers(room),
				State:            openapi.RoomState(room.State),
				Settings:         convertRoomSettingsToOpenAPI(room.Settings),
				PickState:        openapi.PickState(room.PickState),
				ParticipantCount: room.ParticipantCount(),
				PickedBalls:      utils.ConvertPickedBallsToOpenAPI(room.PickedBalls),
				QrCodeVisible:    room.QrCodeVisible,
				BingoSummaries:   convertBingoSummariesToOpenAPI(room.BingoSummaries()),
				ReachSummaries:   convertReachSummariesToOpenAPI(room.ReachSummaries()),
			},
		}, true, nil
	default:
		return nil, false, nil
	}
}
