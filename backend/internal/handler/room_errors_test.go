package handler

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/traP-jp/h26_07/backend/internal/model"
	"github.com/traP-jp/h26_07/backend/internal/repository"
)

func TestMapRoomServiceErrorPreservesResponses(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		fallback string
		mappings []roomErrorMapping
		status   int
		message  string
	}{
		{
			name:     "create settings",
			err:      model.ErrRoomSettingsInvalid,
			fallback: "failed to create room",
			mappings: []roomErrorMapping{roomSettingsInvalidMapping},
			status:   http.StatusBadRequest,
			message:  "invalid room settings",
		},
		{
			name:     "create fallback",
			err:      errors.New("database error"),
			fallback: "failed to create room",
			status:   http.StatusInternalServerError,
			message:  "failed to create room",
		},
		{
			name:     "get repository not found",
			err:      repository.ErrRoomNotFound,
			fallback: "internal server error",
			mappings: []roomErrorMapping{repositoryRoomNotFoundMapping},
			status:   http.StatusNotFound,
			message:  "room not found",
		},
		{
			name:     "get model not found remains fallback",
			err:      model.ErrRoomNotFound,
			fallback: "internal server error",
			mappings: []roomErrorMapping{repositoryRoomNotFoundMapping},
			status:   http.StatusInternalServerError,
			message:  "internal server error",
		},
		{
			name:     "list fallback capitalization",
			err:      errors.New("database error"),
			fallback: "Internal Server Error",
			status:   http.StatusInternalServerError,
			message:  "Internal Server Error",
		},
		{
			name:     "room not found",
			err:      model.ErrRoomNotFound,
			fallback: "internal server error",
			mappings: []roomErrorMapping{roomNotFoundMapping},
			status:   http.StatusNotFound,
			message:  "room not found",
		},
		{
			name:     "admin required",
			err:      model.ErrRoomForbidden,
			fallback: "internal server error",
			mappings: []roomErrorMapping{adminRequiredMapping},
			status:   http.StatusForbidden,
			message:  "admin required",
		},
		{
			name:     "join conflict",
			err:      model.ErrRoomNotJoinable,
			fallback: "internal server error",
			mappings: []roomErrorMapping{roomNotFoundMapping, roomNotJoinableMapping},
			status:   http.StatusConflict,
			message:  "room is not joinable",
		},
		{
			name:     "message invalid",
			err:      model.ErrMessageInvalid,
			fallback: "internal server error",
			mappings: []roomErrorMapping{
				messageInvalidMapping, roomNotFoundMapping, messageNotAllowedMapping, messageNotPostableMapping,
			},
			status:  http.StatusBadRequest,
			message: "message invalid",
		},
		{
			name:     "message not found",
			err:      repository.ErrRoomNotFound,
			fallback: "internal server error",
			mappings: []roomErrorMapping{
				messageInvalidMapping, roomNotFoundMapping, messageNotAllowedMapping, messageNotPostableMapping,
			},
			status:  http.StatusNotFound,
			message: "room not found",
		},
		{
			name:     "message forbidden",
			err:      model.ErrRoomMessageNotAllowed,
			fallback: "internal server error",
			mappings: []roomErrorMapping{
				messageInvalidMapping, roomNotFoundMapping, messageNotAllowedMapping, messageNotPostableMapping,
			},
			status:  http.StatusForbidden,
			message: "message not allowed",
		},
		{
			name:     "messages forbidden",
			err:      model.ErrRoomForbidden,
			fallback: "internal server error",
			mappings: []roomErrorMapping{roomNotFoundMapping, roomMessagesForbiddenMapping},
			status:   http.StatusForbidden,
			message:  "room message not allowed",
		},
		{
			name:     "settings conflict",
			err:      model.ErrRoomNotConfigurable,
			fallback: "internal server error",
			mappings: []roomErrorMapping{
				roomNotFoundMapping, adminRequiredMapping, roomSettingsInvalidMapping, roomNotConfigurableMapping,
			},
			status:  http.StatusConflict,
			message: "room is not configurable",
		},
		{
			name:     "pick start conflict",
			err:      model.ErrNoDrawableBalls,
			fallback: "internal server error",
			mappings: []roomErrorMapping{roomNotFoundMapping, adminRequiredMapping, pickNotStartableMapping},
			status:   http.StatusConflict,
			message:  "pick is not startable",
		},
		{
			name:     "pick cancel conflict",
			err:      model.ErrRoomPickNotCancelable,
			fallback: "internal server error",
			mappings: []roomErrorMapping{roomNotFoundMapping, adminRequiredMapping, pickNotCancelableMapping},
			status:   http.StatusConflict,
			message:  "pick is not cancelable",
		},
		{
			name:     "game start conflict",
			err:      model.ErrRoomNotStartable,
			fallback: "internal server error",
			mappings: []roomErrorMapping{roomNotFoundMapping, adminRequiredMapping, roomNotStartableMapping},
			status:   http.StatusConflict,
			message:  "room not startable",
		},
		{
			name:     "game finish conflict",
			err:      model.ErrRoomNotFinishable,
			fallback: "internal server error",
			mappings: []roomErrorMapping{roomNotFoundMapping, adminRequiredMapping, roomNotFinishableMapping},
			status:   http.StatusConflict,
			message:  "room not finishable",
		},
		{
			name:     "pick finish conflict",
			err:      model.ErrBallAlreadyPicked,
			fallback: "internal server error",
			mappings: []roomErrorMapping{roomNotFoundMapping, adminRequiredMapping, pickNotFinishableMapping},
			status:   http.StatusConflict,
			message:  "pick is not finishable",
		},
		{
			name:     "wrapped cause",
			err:      fmt.Errorf("service operation: %w", model.ErrRoomNotJoinable),
			fallback: "internal server error",
			mappings: []roomErrorMapping{roomNotFoundMapping, roomNotJoinableMapping},
			status:   http.StatusConflict,
			message:  "room is not joinable",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			response := mapRoomServiceError(test.err, test.fallback, test.mappings...)
			if response.status != test.status || response.message != test.message {
				t.Fatalf("got (%d, %q), want (%d, %q)", response.status, response.message, test.status, test.message)
			}
		})
	}
}
