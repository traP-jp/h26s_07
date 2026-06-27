package handler

import (
	"encoding/json"
	"fmt"

	"github.com/traP-jp/h26_07/backend/internal/openapi"
)

func marshalWebSocketEvent(mode openapi.WebSocketMode, event any) ([]byte, error) {
	payload, ok, err := marshalWebSocketEventForMode(mode, event)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("%T does not match websocket mode %q", event, mode)
	}
	return payload, nil
}

func marshalWebSocketEventForMode(mode openapi.WebSocketMode, event any) ([]byte, bool, error) {
	openapiEvent, ok, err := openapiWebSocketEvent(mode, event)
	if err != nil || !ok {
		return nil, ok, err
	}
	payload, err := json.Marshal(openapiEvent)
	return payload, true, err
}

func openapiWebSocketEvent(mode openapi.WebSocketMode, event any) (any, bool, error) {
	switch mode {
	case openapi.Participant:
		return participantWebSocketEvent(event)
	case openapi.Display:
		return displayWebSocketEvent(event)
	default:
		return nil, false, fmt.Errorf("unknown websocket mode %q", mode)
	}
}

func participantWebSocketEvent(event any) (openapi.ParticipantWebSocketEvent, bool, error) {
	if wrapped, ok := event.(openapi.ParticipantWebSocketEvent); ok {
		return wrapped, true, nil
	}
	if isDisplayWebSocketEvent(event) {
		return openapi.ParticipantWebSocketEvent{}, false, nil
	}

	var result openapi.ParticipantWebSocketEvent
	var err error
	switch event := event.(type) {
	case openapi.ParticipantInitializedEvent:
		err = result.FromParticipantInitializedEvent(event)
	case openapi.ParticipantGameStartedEvent:
		err = result.FromParticipantGameStartedEvent(event)
	case openapi.ParticipantPickStartedEvent:
		err = result.FromParticipantPickStartedEvent(event)
	case openapi.ParticipantPickCanceledEvent:
		err = result.FromParticipantPickCanceledEvent(event)
	case openapi.ParticipantPickFinishedEvent:
		err = result.FromParticipantPickFinishedEvent(event)
	case openapi.ParticipantGameFinishedEvent:
		err = result.FromParticipantGameFinishedEvent(event)
	case openapi.ParticipantMessageCreatedEvent:
		err = result.FromParticipantMessageCreatedEvent(event)
	case openapi.ParticipantAllPickedEvent:
		err = result.FromParticipantAllPickedEvent(event)
	case openapi.ParticipantGameSettingsUpdatedEvent:
		err = result.FromParticipantGameSettingsUpdatedEvent(event)
	default:
		return result, false, fmt.Errorf("%T is not a websocket event schema", event)
	}
	return result, true, err
}

func displayWebSocketEvent(event any) (openapi.DisplayWebSocketEvent, bool, error) {
	if wrapped, ok := event.(openapi.DisplayWebSocketEvent); ok {
		return wrapped, true, nil
	}
	if isParticipantWebSocketEvent(event) {
		return openapi.DisplayWebSocketEvent{}, false, nil
	}

	var result openapi.DisplayWebSocketEvent
	var err error
	switch event := event.(type) {
	case openapi.DisplayInitializedEvent:
		err = result.FromDisplayInitializedEvent(event)
	case openapi.DisplayGameStartedEvent:
		err = result.FromDisplayGameStartedEvent(event)
	case openapi.DisplayPickStartedEvent:
		err = result.FromDisplayPickStartedEvent(event)
	case openapi.DisplayPickCanceledEvent:
		err = result.FromDisplayPickCanceledEvent(event)
	case openapi.DisplayPickFinishedEvent:
		err = result.FromDisplayPickFinishedEvent(event)
	case openapi.DisplayGameFinishedEvent:
		err = result.FromDisplayGameFinishedEvent(event)
	case openapi.DisplayShowQRCodeEvent:
		err = result.FromDisplayShowQRCodeEvent(event)
	case openapi.DisplayHideQRCodeEvent:
		err = result.FromDisplayHideQRCodeEvent(event)
	case openapi.DisplayMessageCreatedEvent:
		err = result.FromDisplayMessageCreatedEvent(event)
	case openapi.DisplayAllPickedEvent:
		err = result.FromDisplayAllPickedEvent(event)
	case openapi.DisplayGameSettingsUpdatedEvent:
		err = result.FromDisplayGameSettingsUpdatedEvent(event)
	default:
		return result, false, fmt.Errorf("%T is not a websocket event schema", event)
	}
	return result, true, err
}

func isParticipantWebSocketEvent(event any) bool {
	switch event.(type) {
	case openapi.ParticipantWebSocketEvent,
		openapi.ParticipantInitializedEvent,
		openapi.ParticipantGameStartedEvent,
		openapi.ParticipantPickStartedEvent,
		openapi.ParticipantPickCanceledEvent,
		openapi.ParticipantPickFinishedEvent,
		openapi.ParticipantGameFinishedEvent,
		openapi.ParticipantMessageCreatedEvent,
		openapi.ParticipantAllPickedEvent,
		openapi.ParticipantGameSettingsUpdatedEvent:
		return true
	default:
		return false
	}
}

func isDisplayWebSocketEvent(event any) bool {
	switch event.(type) {
	case openapi.DisplayWebSocketEvent,
		openapi.DisplayInitializedEvent,
		openapi.DisplayGameStartedEvent,
		openapi.DisplayPickStartedEvent,
		openapi.DisplayPickCanceledEvent,
		openapi.DisplayPickFinishedEvent,
		openapi.DisplayGameFinishedEvent,
		openapi.DisplayShowQRCodeEvent,
		openapi.DisplayHideQRCodeEvent,
		openapi.DisplayMessageCreatedEvent,
		openapi.DisplayAllPickedEvent,
		openapi.DisplayGameSettingsUpdatedEvent:
		return true
	default:
		return false
	}
}
