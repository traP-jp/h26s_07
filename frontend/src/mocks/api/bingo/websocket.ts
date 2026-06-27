import type {
  AllPickedBody,
  DisplayGameFinishedBody,
  DisplayGameSettingsUpdatedBody,
  DisplayGameStartedBody,
  DisplayInitializedBody,
  DisplayPickFinishedBody,
  HideQrCodeBody,
  Message,
  MessageCreatedBody,
  ParticipantGameFinishedBody,
  ParticipantGameSettingsUpdatedBody,
  ParticipantGameStartedBody,
  ParticipantInitializedBody,
  ParticipantPickFinishedBody,
  PickCanceledBody,
  PickStartedBody,
  ShowQrCodeBody,
  WebSocketEventType,
} from '@/api/schema'
import { ws } from 'msw'

import {
  createMessage,
  createUser,
  emptyCardChanges,
  finishPick,
  getRoom,
  isApiError,
  isParticipant,
  pathParam,
  socketConnections,
  startRoom,
  touch,
  type FinishPickResult,
  type MockRoom,
  type MockSocketConnection,
} from './core'

type MockWebSocketEvent<TBody> = {
  type: WebSocketEventType
  body: TBody
}

const roomSocket = ws.link('*/api/rooms/:roomId/ws')

const demoChatMessages = [
  { delayMs: 2000, authorId: 'rurun', content: '準備できています' },
  { delayMs: 5000, authorId: 'mumumu', content: 'そろそろ始めます' },
  { delayMs: 9000, authorId: 'howard127', content: 'ビンゴ楽しみ' },
  { delayMs: 12000, authorId: 'mumumu', content: 'では始めます' },
  { delayMs: 15000, authorId: 'kurao', content: 'よろしくお願いします' },
] as const

const scheduledDemoChatRoomIds = new Set<string>()

function sendEvent<TBody>(
  connection: MockSocketConnection,
  type: WebSocketEventType,
  body: TBody,
): void {
  const event = {
    type,
    body,
  } satisfies MockWebSocketEvent<TBody>

  connection.send(JSON.stringify(event))
}

function roomConnectionList(roomId: string): MockSocketConnection[] {
  return [...socketConnections].filter((connection) => connection.roomId === roomId)
}

function participantInitializedBody(
  roomState: MockRoom,
  userId: string,
): ParticipantInitializedBody {
  const card = roomState.cards.get(userId)
  const body: ParticipantInitializedBody = {
    state: roomState.room.state,
    settings: roomState.room.settings,
    pickState: roomState.room.pickState,
    pickedBalls: roomState.pickedBalls,
    bingoSummaries: roomState.room.bingoSummaries,
  }

  if (card !== undefined && roomState.room.state !== 'waiting') {
    body.card = card
  }

  return body
}

function displayInitializedBody(roomState: MockRoom): DisplayInitializedBody {
  return {
    state: roomState.room.state,
    settings: roomState.room.settings,
    pickState: roomState.room.pickState,
    participantCount: roomState.room.participants.length,
    pickedBalls: roomState.pickedBalls,
    qrCodeVisible: roomState.room.qrCodeVisible,
    bingoSummaries: roomState.room.bingoSummaries,
  }
}

function sendInitialized(connection: MockSocketConnection, roomState: MockRoom): void {
  if (connection.mode === 'participant') {
    sendEvent<ParticipantInitializedBody>(
      connection,
      'Initialized',
      participantInitializedBody(roomState, connection.userId),
    )
    return
  }

  sendEvent<DisplayInitializedBody>(connection, 'Initialized', displayInitializedBody(roomState))
}

function broadcastRoom(
  roomState: MockRoom,
  send: (connection: MockSocketConnection) => void,
): void {
  for (const connection of roomConnectionList(roomState.room.roomId)) {
    send(connection)
  }
}

export function broadcastGameStarted(roomState: MockRoom): void {
  broadcastRoom(roomState, (connection) => {
    if (connection.mode === 'participant') {
      const card = roomState.cards.get(connection.userId)
      if (card !== undefined) {
        sendEvent<ParticipantGameStartedBody>(connection, 'GameStarted', { card })
      }
      return
    }

    sendEvent<DisplayGameStartedBody>(connection, 'GameStarted', {
      participantCount: roomState.room.participants.length,
    })
  })
}

export function broadcastPickStarted(roomState: MockRoom): void {
  broadcastRoom(roomState, (connection) => {
    sendEvent<PickStartedBody>(connection, 'PickStarted', {})
  })
}

export function broadcastPickCanceled(roomState: MockRoom): void {
  broadcastRoom(roomState, (connection) => {
    sendEvent<PickCanceledBody>(connection, 'PickCanceled', {})
  })
}

export function broadcastPickFinished(roomState: MockRoom, result: FinishPickResult): void {
  broadcastRoom(roomState, (connection) => {
    if (connection.mode === 'participant') {
      const card = roomState.cards.get(connection.userId)
      if (card === undefined) {
        return
      }

      sendEvent<ParticipantPickFinishedBody>(connection, 'PickFinished', {
        pickedBall: result.pickedBall,
        pickState: roomState.room.pickState,
        card,
        cardChanges: result.cardChangesByUserId.get(connection.userId) ?? emptyCardChanges(),
        pickedBalls: roomState.pickedBalls,
        bingoSummaries: roomState.room.bingoSummaries,
        newBingos: result.newBingos,
        newReaches: result.newReaches,
      })
      return
    }

    sendEvent<DisplayPickFinishedBody>(connection, 'PickFinished', {
      pickedBall: result.pickedBall,
      pickState: roomState.room.pickState,
      participantCount: roomState.room.participants.length,
      bingoSummaries: roomState.room.bingoSummaries,
      newBingos: result.newBingos,
      newReaches: result.newReaches,
      pickedBalls: roomState.pickedBalls,
    })
  })
}

export function broadcastGameFinished(roomState: MockRoom): void {
  broadcastRoom(roomState, (connection) => {
    if (connection.mode === 'participant') {
      const card = roomState.cards.get(connection.userId)
      if (card !== undefined) {
        sendEvent<ParticipantGameFinishedBody>(connection, 'GameFinished', {
          state: 'finished',
          pickState: 'idle',
          card,
          bingoSummaries: roomState.room.bingoSummaries,
        })
      }
      return
    }

    sendEvent<DisplayGameFinishedBody>(connection, 'GameFinished', {
      state: 'finished',
      pickState: 'idle',
      participantCount: roomState.room.participants.length,
      bingoSummaries: roomState.room.bingoSummaries,
    })
  })
}

export function broadcastMessageCreated(roomState: MockRoom, message: Message): void {
  broadcastRoom(roomState, (connection) => {
    sendEvent<MessageCreatedBody>(connection, 'MessageCreated', { message })
  })
}

function scheduleDemoChatMessages(roomState: MockRoom): void {
  const roomId = roomState.room.roomId
  if (scheduledDemoChatRoomIds.has(roomId)) {
    return
  }

  scheduledDemoChatRoomIds.add(roomId)
  for (const demoMessage of demoChatMessages) {
    window.setTimeout(() => {
      const latestRoomState = getRoom(roomId)
      if (
        latestRoomState === undefined ||
        latestRoomState.room.state === 'finished' ||
        roomConnectionList(roomId).length === 0
      ) {
        return
      }

      const message = createMessage(createUser(demoMessage.authorId), demoMessage.content)
      latestRoomState.messages.push(message)
      touch(latestRoomState.room)
      broadcastMessageCreated(latestRoomState, message)
    }, demoMessage.delayMs)
  }
}

export function broadcastAllPicked(roomState: MockRoom): void {
  const body: AllPickedBody = { pickedBalls: roomState.pickedBalls }
  broadcastRoom(roomState, (connection) => {
    sendEvent<AllPickedBody>(connection, 'AllPicked', body)
  })
}

export function broadcastGameSettingsUpdated(roomState: MockRoom): void {
  broadcastRoom(roomState, (connection) => {
    if (connection.mode === 'participant') {
      sendEvent<ParticipantGameSettingsUpdatedBody>(connection, 'GameSettingsUpdated', {
        settings: roomState.room.settings,
      })
      return
    }

    sendEvent<DisplayGameSettingsUpdatedBody>(connection, 'GameSettingsUpdated', {
      settings: roomState.room.settings,
    })
  })
}

export function broadcastShowQRCode(roomState: MockRoom): void {
  broadcastRoom(roomState, (connection) => {
    if (connection.mode === 'display') {
      sendEvent<ShowQrCodeBody>(connection, 'ShowQRCode', {})
    }
  })
}

export function broadcastHideQRCode(roomState: MockRoom): void {
  broadcastRoom(roomState, (connection) => {
    if (connection.mode === 'display') {
      sendEvent<HideQrCodeBody>(connection, 'HideQRCode', {})
    }
  })
}

function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => {
    window.setTimeout(resolve, ms)
  })
}

async function runMockPickCycle(roomState: MockRoom): Promise<void> {
  if (roomState.room.state === 'waiting') {
    const startError = startRoom(roomState)
    if (startError !== undefined) {
      return
    }
    broadcastGameStarted(roomState)
  }

  if (roomState.room.state !== 'playing' || roomState.room.pickState !== 'idle') {
    return
  }

  roomState.room.pickState = 'picking'
  touch(roomState.room)
  broadcastPickStarted(roomState)

  await sleep(800)

  const result = finishPick(roomState)
  if (isApiError(result)) {
    return
  }

  broadcastPickFinished(roomState, result)
  if (result.allPicked) {
    broadcastAllPicked(roomState)
  }
}

function handleSocketMessage(connection: MockSocketConnection, data: unknown): void {
  if (typeof data !== 'string') {
    return
  }

  let payload: { type?: string } | undefined
  try {
    payload = JSON.parse(data) as { type?: string }
  } catch {
    return
  }

  const roomState = getRoom(connection.roomId)
  if (roomState === undefined) {
    return
  }

  if (payload.type === 'mock:start') {
    const startError = startRoom(roomState)
    if (startError === undefined) {
      broadcastGameStarted(roomState)
    }
    return
  }

  if (payload.type === 'mock:pick') {
    void runMockPickCycle(roomState)
    return
  }

  if (payload.type === 'mock:finish' && roomState.room.state === 'playing') {
    const wasPicking = roomState.room.pickState === 'picking'
    roomState.room.state = 'finished'
    roomState.room.pickState = 'idle'
    touch(roomState.room)

    if (wasPicking) {
      broadcastPickCanceled(roomState)
    }
    broadcastGameFinished(roomState)
  }
}

export const roomWebSocketHandler = roomSocket.addEventListener(
  'connection',
  ({ client, params }) => {
    const roomId = pathParam(params.roomId)
    const roomState = roomId === undefined ? undefined : getRoom(roomId)
    const mode = client.url.searchParams.get('mode')
    const userId = client.url.searchParams.get('userId')?.trim() || 'mumumu'

    if (roomId === undefined || roomState === undefined) {
      client.close(1008, 'Room not found')
      return
    }

    if (mode !== 'participant' && mode !== 'display') {
      client.close(1008, 'Invalid mode')
      return
    }

    if (mode === 'participant' && !isParticipant(roomState.room, userId)) {
      client.close(1008, 'Participant required')
      return
    }

    const connection: MockSocketConnection = {
      roomId,
      mode,
      userId,
      send(data) {
        client.send(data)
      },
      close(code, reason) {
        client.close(code, reason)
      },
    }

    socketConnections.add(connection)
    client.addEventListener(
      'close',
      () => {
        socketConnections.delete(connection)
      },
      { once: true },
    )
    client.addEventListener('message', (event) => {
      handleSocketMessage(connection, event.data)
    })

    sendInitialized(connection, roomState)
    scheduleDemoChatMessages(roomState)
  },
)
