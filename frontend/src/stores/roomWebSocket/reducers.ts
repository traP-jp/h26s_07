import type {
  DisplayGameFinishedBody,
  DisplayGameStartedBody,
  DisplayInitializedBody,
  DisplayPickFinishedBody,
  ParticipantGameFinishedBody,
  ParticipantGameStartedBody,
  ParticipantInitializedBody,
  ParticipantPickFinishedBody,
  WebSocketMode,
} from '@/api/schema'

import { mergeById } from './merge'
import type { RoomWebSocketData } from './state'
import type { RoomWebSocketEvent } from './types'

type EventOfType<Type extends RoomWebSocketEvent['type']> = Extract<
  RoomWebSocketEvent,
  { type: Type }
>
type InitializedBody = ParticipantInitializedBody | DisplayInitializedBody
type GameStartedBody = ParticipantGameStartedBody | DisplayGameStartedBody
type PickFinishedBody = ParticipantPickFinishedBody | DisplayPickFinishedBody
type GameFinishedBody = ParticipantGameFinishedBody | DisplayGameFinishedBody
type RoomWebSocketUpdate = {
  [Key in keyof RoomWebSocketData]: [Key, RoomWebSocketData[Key]]
}[keyof RoomWebSocketData]

function reduceParticipantJoined(
  state: RoomWebSocketData,
  event: EventOfType<'ParticipantJoined'>,
): RoomWebSocketUpdate[] {
  return [
    [
      'participants',
      mergeById(state.participants, [event.body.participant], (user) => user.userId),
    ],
  ]
}

function reduceInitialized(
  state: RoomWebSocketData,
  event: EventOfType<'Initialized'>,
  mode: WebSocketMode | null,
): RoomWebSocketUpdate[] {
  const body = event.body as InitializedBody
  const updates: RoomWebSocketUpdate[] = [
    ['initializationCount', state.initializationCount + 1],
    ['activityLog', body.activityLog],
    ['participants', body.participants],
    ['roomState', body.state],
    ['settings', body.settings],
    ['pickState', body.pickState],
    ['latestBall', body.pickedBalls.at(-1) ?? null],
    ['latestCardChanges', null],
    ['latestNewBingos', []],
    ['latestNewReaches', []],
    ['pickedBalls', body.pickedBalls],
    ['bingoSummaries', body.bingoSummaries],
    ['reachSummaries', body.reachSummaries],
  ]

  if (mode === 'participant') {
    updates.push(['card', (body as ParticipantInitializedBody).card ?? null])
  } else {
    const displayBody = body as DisplayInitializedBody
    updates.push(['participantCount', displayBody.participantCount])
    updates.push(['qrCodeVisible', displayBody.qrCodeVisible])
  }

  return updates
}

function reduceGameStarted(
  event: EventOfType<'GameStarted'>,
  mode: WebSocketMode | null,
): RoomWebSocketUpdate[] {
  const body = event.body as GameStartedBody
  const updates: RoomWebSocketUpdate[] = [
    ['roomState', 'playing'],
    ['pickState', 'idle'],
  ]

  if (mode === 'participant') {
    updates.push(['card', (body as ParticipantGameStartedBody).card])
  } else {
    updates.push(['participantCount', (body as DisplayGameStartedBody).participantCount])
  }

  return updates
}

function reducePickFinished(
  state: RoomWebSocketData,
  event: EventOfType<'PickFinished'>,
  mode: WebSocketMode | null,
): RoomWebSocketUpdate[] {
  const body = event.body as PickFinishedBody
  const updates: RoomWebSocketUpdate[] = [
    ['activityLog', mergeById(state.activityLog, body.newActivityLog, (entry) => entry.id)],
    ['latestBall', body.pickedBall],
    ['latestNewBingos', body.newBingos],
    ['latestNewReaches', body.newReaches],
    ['pickState', body.pickState],
    ['pickedBalls', body.pickedBalls],
    ['bingoSummaries', body.bingoSummaries],
    ['reachSummaries', body.reachSummaries],
  ]

  if (mode === 'participant') {
    const participantBody = body as ParticipantPickFinishedBody
    updates.push(['card', participantBody.card])
    updates.push(['latestCardChanges', participantBody.cardChanges])
  } else {
    updates.push(['participantCount', (body as DisplayPickFinishedBody).participantCount])
    updates.push(['latestCardChanges', null])
  }

  return updates
}

function reduceGameFinished(
  event: EventOfType<'GameFinished'>,
  mode: WebSocketMode | null,
): RoomWebSocketUpdate[] {
  const body = event.body as GameFinishedBody
  const updates: RoomWebSocketUpdate[] = [
    ['roomState', body.state],
    ['pickState', body.pickState],
    ['bingoSummaries', body.bingoSummaries],
    ['reachSummaries', body.reachSummaries],
  ]

  if (mode === 'participant') {
    updates.push(['card', (body as ParticipantGameFinishedBody).card])
  } else {
    updates.push(['participantCount', (body as DisplayGameFinishedBody).participantCount])
  }

  return updates
}

function reduceMessageCreated(
  state: RoomWebSocketData,
  event: EventOfType<'MessageCreated'>,
): RoomWebSocketUpdate[] {
  return [
    ['latestMessage', event.body.message],
    [
      'chatMessages',
      mergeById(state.chatMessages, [event.body.message], (message) => message.messageId),
    ],
  ]
}

export function reduceRoomWebSocketEvent(
  state: RoomWebSocketData,
  event: RoomWebSocketEvent,
  mode: WebSocketMode | null,
): RoomWebSocketUpdate[] {
  const updates: RoomWebSocketUpdate[] = [['latestEvent', event]]

  if ('participantCount' in event.body && typeof event.body.participantCount === 'number') {
    updates.push(['participantCount', event.body.participantCount])
  }

  switch (event.type) {
    case 'RoomDeleted':
      return [...updates, ['roomDeleted', true]]
    case 'ParticipantJoined':
      return [...updates, ...reduceParticipantJoined(state, event)]
    case 'Initialized':
      return [...updates, ...reduceInitialized(state, event, mode)]
    case 'GameStarted':
      return [...updates, ...reduceGameStarted(event, mode)]
    case 'PickStarted':
      return [...updates, ['pickState', 'picking']]
    case 'PickCanceled':
      return [...updates, ['pickState', 'idle']]
    case 'PickFinished':
      return [...updates, ...reducePickFinished(state, event, mode)]
    case 'GameFinished':
      return [...updates, ...reduceGameFinished(event, mode)]
    case 'ShowQRCode':
      return [...updates, ['qrCodeVisible', true]]
    case 'HideQRCode':
      return [...updates, ['qrCodeVisible', false]]
    case 'MessageCreated':
      return [...updates, ...reduceMessageCreated(state, event)]
    case 'AllPicked':
      return [...updates, ['pickedBalls', event.body.pickedBalls], ['pickState', 'exhausted']]
    case 'GameSettingsUpdated':
      return [...updates, ['settings', event.body.settings]]
  }
}
