package handler

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/traP-jp/h26_07/backend/internal/model"
	"github.com/traP-jp/h26_07/backend/internal/openapi"
	"github.com/traP-jp/h26_07/backend/internal/repository"
)

type roomErrorMapping struct {
	matches func(error) bool
	status  int
	message string
}

type roomErrorResponse struct {
	status  int
	message string
}

var (
	roomNotFoundMapping = roomErrorMapping{
		matches: isRoomNotFoundError,
		status:  http.StatusNotFound,
		message: "room not found",
	}
	repositoryRoomNotFoundMapping = roomErrorMapping{
		matches: func(err error) bool { return errors.Is(err, repository.ErrRoomNotFound) },
		status:  http.StatusNotFound,
		message: "room not found",
	}
	adminRequiredMapping       = roomErrorCauses(http.StatusForbidden, "admin required", model.ErrRoomForbidden)
	roomSettingsInvalidMapping = roomErrorCauses(
		http.StatusBadRequest,
		"invalid room settings",
		model.ErrRoomSettingsInvalid,
	)
	roomNotJoinableMapping = roomErrorCauses(
		http.StatusConflict,
		"room is not joinable",
		model.ErrRoomNotJoinable,
	)
	messageInvalidMapping = roomErrorCauses(
		http.StatusBadRequest,
		"message invalid",
		model.ErrMessageInvalid,
	)
	messageNotAllowedMapping = roomErrorCauses(
		http.StatusForbidden,
		"message not allowed",
		model.ErrRoomMessageNotAllowed,
	)
	roomMessagesForbiddenMapping = roomErrorCauses(
		http.StatusForbidden,
		"room message not allowed",
		model.ErrRoomForbidden,
	)
	messageNotPostableMapping = roomErrorCauses(
		http.StatusConflict,
		"message not postable",
		model.ErrRoomMessageNotPostable,
	)
	roomNotConfigurableMapping = roomErrorCauses(
		http.StatusConflict,
		"room is not configurable",
		model.ErrRoomNotConfigurable,
	)
	pickNotStartableMapping = roomErrorCauses(
		http.StatusConflict,
		"pick is not startable",
		model.ErrRoomPickNotStartable,
		model.ErrNoDrawableBalls,
	)
	pickNotCancelableMapping = roomErrorCauses(
		http.StatusConflict,
		"pick is not cancelable",
		model.ErrRoomPickNotCancelable,
	)
	roomNotStartableMapping = roomErrorCauses(
		http.StatusConflict,
		"room not startable",
		model.ErrRoomNotStartable,
	)
	roomNotFinishableMapping = roomErrorCauses(
		http.StatusConflict,
		"room not finishable",
		model.ErrRoomNotFinishable,
	)
	pickNotFinishableMapping = roomErrorCauses(
		http.StatusConflict,
		"pick is not finishable",
		model.ErrRoomPickNotFinishable,
		model.ErrNoDrawableBalls,
		model.ErrBallAlreadyPicked,
	)
)

func roomErrorCauses(status int, message string, causes ...error) roomErrorMapping {
	return roomErrorMapping{
		matches: func(err error) bool {
			for _, cause := range causes {
				if errors.Is(err, cause) {
					return true
				}
			}
			return false
		},
		status:  status,
		message: message,
	}
}

func mapRoomServiceError(err error, fallback string, mappings ...roomErrorMapping) roomErrorResponse {
	for _, mapping := range mappings {
		if mapping.matches(err) {
			return roomErrorResponse{status: mapping.status, message: mapping.message}
		}
	}
	return roomErrorResponse{status: http.StatusInternalServerError, message: fallback}
}

func writeRoomServiceError(c *echo.Context, err error, fallback string, mappings ...roomErrorMapping) error {
	response := mapRoomServiceError(err, fallback, mappings...)
	return c.JSON(response.status, openapi.Error{Message: response.message})
}

func isRoomNotFoundError(err error) bool {
	return errors.Is(err, repository.ErrRoomNotFound) || errors.Is(err, model.ErrRoomNotFound)
}
