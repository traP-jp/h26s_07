package handler

import (
	"github.com/google/uuid"
	"github.com/traP-jp/h26_07/backend/internal/model"
	"github.com/traP-jp/h26_07/backend/internal/openapi"
	"github.com/traP-jp/h26_07/backend/internal/service"
	"github.com/traP-jp/h26_07/backend/internal/utils"
)

func convertUserIDsToModel(userIDS []openapi.UserID) []model.UserID {
	var result []model.UserID
	for _, userID := range userIDS {
		result = append(result, model.UserID(userID))
	}
	return result
}

func convertOptionalUserIDsToModel(userIDs *[]openapi.UserID) []model.UserID {
	if userIDs == nil {
		return []model.UserID{}
	}
	return convertUserIDsToModel(*userIDs)
}

func convertOptionalUserIDsToModelPointer(userIDs *[]openapi.UserID) *[]model.UserID {
	if userIDs == nil {
		return nil
	}
	converted := convertUserIDsToModel(*userIDs)
	return &converted
}

func convertUserIDsToOpenAPI(userIDs []model.UserID) []openapi.User {
	var result []openapi.User
	for _, userID := range userIDs {
		result = append(result, openapi.User{UserID: openapi.UserID(userID)})
	}
	return result
}

func convertRoomSettingsToModel(settings openapi.GameSettingsInput) model.RoomSettings {
	return model.RoomSettings{
		Name:        settings.Name,
		Description: settings.Description,
		Admins:      convertOptionalUserIDsToModel(settings.AdminUserIds),
	}
}

func convertRoomSettingsUpdateToService(settings openapi.GameSettingsInput) service.RoomSettingsUpdate {
	return service.RoomSettingsUpdate{
		Name:        settings.Name,
		Description: settings.Description,
		Admins:      convertOptionalUserIDsToModelPointer(settings.AdminUserIds),
	}
}

func adminUserIDsSpecifiedEmpty(settings openapi.GameSettingsInput) bool {
	return settings.AdminUserIds != nil && len(*settings.AdminUserIds) == 0
}

func convertRoomSettingsToOpenAPI(settings model.RoomSettings) openapi.GameSettings {
	return openapi.GameSettings{
		Name:        settings.Name,
		Description: settings.Description,
		Admins:      convertUserIDsToOpenAPI(settings.Admins),
	}
}

func convertParticipantsToOpenAPI(participants []model.Participant) []openapi.ParticipantSummary {
	result := make([]openapi.ParticipantSummary, 0, len(participants))
	for _, participant := range participants {
		result = append(result, openapi.ParticipantSummary{
			User:     openapi.User{UserID: openapi.UserID(participant.UserID)},
			JoinedAt: participant.JoinedAt,
		})
	}
	return result
}

func convertBingoSummariesToOpenAPI(bingoSummaries []model.BingoSummary) []openapi.BingoSummary {
	result := make([]openapi.BingoSummary, 0, len(bingoSummaries))
	for _, bingoSummary := range bingoSummaries {
		bingoOrders := make([]int, 0, len(bingoSummary.BingoOrders))
		for _, bingoOrder := range bingoSummary.BingoOrders {
			bingoOrders = append(bingoOrders, int(bingoOrder))
		}
		result = append(result, openapi.BingoSummary{
			BingoOrders: bingoOrders,
			CreatedAt:   bingoSummary.CreatedAt,
			User:        openapi.User{UserID: openapi.UserID(bingoSummary.UserID)},
		})
	}
	return result
}

func convertReachSummariesToOpenAPI(reachSummaries []model.ReachSummary) []openapi.ReachSummary {
	result := make([]openapi.ReachSummary, 0, len(reachSummaries))
	for _, reachSummary := range reachSummaries {
		result = append(result, openapi.ReachSummary{
			CreatedAt: reachSummary.CreatedAt,
			User:      openapi.User{UserID: openapi.UserID(reachSummary.UserID)},
		})
	}
	return result
}

func convertRoom(room *model.Room) openapi.Room {
	return openapi.Room{
		ActivityLog:    utils.ConvertActivityLogToOpenAPI(room.ActivityLog()),
		RoomID:         openapi.RoomID(room.RoomID.String()),
		Settings:       convertRoomSettingsToOpenAPI(room.Settings),
		State:          openapi.RoomState(room.State),
		PickState:      openapi.PickState(room.PickState),
		QrCodeVisible:  room.QrCodeVisible,
		Participants:   convertParticipantsToOpenAPI(room.Participants),
		BingoSummaries: convertBingoSummariesToOpenAPI(room.BingoSummaries()),
		ReachSummaries: convertReachSummariesToOpenAPI(room.ReachSummaries()),
		CreatedAt:      room.CreatedAt,
		UpdatedAt:      room.UpdatedAt,
		RoomCode:       openapi.RoomCode(room.RoomCode),
	}
}

func convertRoomSummary(rooms []model.RoomSummary) []openapi.Room {
	result := make([]openapi.Room, 0, len(rooms))
	for _, room := range rooms {
		result = append(result,
			openapi.Room{
				ActivityLog:    utils.ConvertActivityLogToOpenAPI(room.ActivityLogEntries),
				RoomID:         openapi.RoomID(room.RoomID.String()),
				Settings:       convertRoomSettingsToOpenAPI(room.Settings),
				State:          openapi.RoomState(room.State),
				PickState:      openapi.PickState(room.PickState),
				QrCodeVisible:  room.QrCodeVisible,
				Participants:   convertParticipantsToOpenAPI(room.Participants),
				BingoSummaries: convertBingoSummariesToOpenAPI(room.BingoSummaries),
				ReachSummaries: convertReachSummariesToOpenAPI(room.ReachSummaries),
				CreatedAt:      room.CreatedAt,
				UpdatedAt:      room.UpdatedAt,
				RoomCode:       openapi.RoomCode(room.RoomCode),
			},
		)
	}
	return result
}

func convertMessageToOpenAPI(message model.Message) openapi.Message {
	return openapi.Message{
		MessageID: uuid.UUID(message.MessageID).String(),
		Content:   message.Content,
		Author: openapi.User{
			UserID: openapi.UserID(message.Author),
		},
		CreatedAt: message.CreatedAt,
	}
}
