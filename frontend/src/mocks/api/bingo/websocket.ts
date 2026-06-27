import type {
  AllPickedBody,
  Card,
  DisplayGameFinishedBody,
  DisplayGameStartedBody,
  DisplayInitializedBody,
  DisplayPickFinishedBody,
  Message,
  MessageCreatedBody,
  PickedBall,
  ParticipantGameFinishedBody,
  ParticipantGameStartedBody,
  ParticipantInitializedBody,
  ParticipantPickFinishedBody,
  PickCanceledBody,
  PickStartedBody,
  Room,
  WebSocketEventType,
} from '@/api/schema'
import { ws } from 'msw'

import { pathParam, socketConnections, type MockSocketConnection } from './core'

type MockWebSocketEvent<TBody> = {
  type: WebSocketEventType
  body: TBody
}

const roomSocket = ws.link('*/api/rooms/:roomId/ws')
const mockPickedBalls: PickedBall[] = [7, 22, 45]
const mockRoom: Room = {
  roomId: '00000000-0000-4000-8000-000000000001',
  roomCode: '123456',
  state: 'waiting',
  pickState: 'idle',
  qrCodeVisible: true,
  participants: [
    {
      user: { userId: 'mumumu' },
      joinedAt: '2026-06-27T10:00:00.000Z',
    },
    {
      user: { userId: 'saba' },
      joinedAt: '2026-06-27T10:01:00.000Z',
    },
  ],
  bingoSummaries: [],
  settings: {
    name: 'デモビンゴ',
    description: 'モック API で動かす待機中のビンゴルームです。',
    admins: [{ userId: 'mumumu' }],
  },
  createdAt: '2026-06-27T10:00:00.000Z',
  updatedAt: '2026-06-27T10:00:00.000Z',
}
const mockCard: Card = {
  cardId: '00000000-0000-4000-8000-000000000201',
  ownerUserId: 'mumumu',
  cells: Array.from({ length: 25 }, (_, index) => {
    if (index === 12) {
      return {
        index,
        number: null,
        displayText: 'FREE',
        cellState: 'open',
      }
    }

    const column = index % 5
    const number = column * 15 + Math.floor(index / 5) + 1
    return {
      index,
      number,
      displayText: String(number),
      cellState: mockPickedBalls.includes(number) ? 'open' : 'closed',
    }
  }),
  bingoLines: [],
  reachLines: [],
}
const mockMessages: Message[] = [
  {
    messageId: '00000000-0000-4000-8000-000000000101',
    content: 'モックルームへようこそ',
    author: { userId: 'mumumu' },
    createdAt: '2026-06-27T10:02:00.000Z',
  },
  {
    messageId: '00000000-0000-4000-8000-000000000102',
    content: '準備できています',
    author: { userId: 'rurun' },
    createdAt: '2026-06-27T10:03:00.000Z',
  },
]

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

function broadcastRoom<TBody>(roomId: string, type: WebSocketEventType, body: TBody): void {
  for (const connection of roomConnectionList(roomId)) {
    sendEvent(connection, type, body)
  }
}

function initializedBody(
  connection: MockSocketConnection,
): ParticipantInitializedBody | DisplayInitializedBody {
  if (connection.mode === 'participant') {
    return {
      state: mockRoom.state,
      settings: mockRoom.settings,
      pickState: mockRoom.pickState,
      pickedBalls: mockPickedBalls,
      bingoSummaries: mockRoom.bingoSummaries,
      card: mockCard,
    }
  }

  return {
    state: mockRoom.state,
    settings: mockRoom.settings,
    pickState: mockRoom.pickState,
    participantCount: mockRoom.participants.length,
    pickedBalls: mockPickedBalls,
    qrCodeVisible: mockRoom.qrCodeVisible,
    bingoSummaries: mockRoom.bingoSummaries,
  }
}

function sendGameStarted(connection: MockSocketConnection): void {
  if (connection.mode === 'participant') {
    sendEvent<ParticipantGameStartedBody>(connection, 'GameStarted', { card: mockCard })
    return
  }

  sendEvent<DisplayGameStartedBody>(connection, 'GameStarted', {
    participantCount: mockRoom.participants.length,
  })
}

function sendPickFinished(connection: MockSocketConnection): void {
  if (connection.mode === 'participant') {
    sendEvent<ParticipantPickFinishedBody>(connection, 'PickFinished', {
      pickedBall: 7,
      pickState: 'idle',
      card: mockCard,
      cardChanges: {
        openedCellIndices: [0],
        newReachLines: [],
        newBingoLines: [],
      },
      pickedBalls: mockPickedBalls,
      bingoSummaries: mockRoom.bingoSummaries,
      newBingos: [],
      newReaches: [],
    })
    return
  }

  sendEvent<DisplayPickFinishedBody>(connection, 'PickFinished', {
    pickedBall: 7,
    pickState: 'idle',
    participantCount: mockRoom.participants.length,
    bingoSummaries: mockRoom.bingoSummaries,
    newBingos: [],
    newReaches: [],
    pickedBalls: mockPickedBalls,
  })
}

function sendGameFinished(connection: MockSocketConnection): void {
  if (connection.mode === 'participant') {
    sendEvent<ParticipantGameFinishedBody>(connection, 'GameFinished', {
      state: 'finished',
      pickState: 'idle',
      card: mockCard,
      bingoSummaries: mockRoom.bingoSummaries,
    })
    return
  }

  sendEvent<DisplayGameFinishedBody>(connection, 'GameFinished', {
    state: 'finished',
    pickState: 'idle',
    participantCount: mockRoom.participants.length,
    bingoSummaries: mockRoom.bingoSummaries,
  })
}

function scheduleDemoChatMessages(roomId: string): void {
  for (const [index, message] of mockMessages.entries()) {
    window.setTimeout(
      () => {
        broadcastRoom<MessageCreatedBody>(roomId, 'MessageCreated', { message })
      },
      2000 + index * 3000,
    )
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

  if (payload.type === 'mock:start') {
    sendGameStarted(connection)
    return
  }

  if (payload.type === 'mock:pick') {
    sendEvent<PickStartedBody>(connection, 'PickStarted', {})
    window.setTimeout(() => {
      sendPickFinished(connection)
    }, 800)
    return
  }

  if (payload.type === 'mock:finish') {
    sendEvent<PickCanceledBody>(connection, 'PickCanceled', {})
    sendGameFinished(connection)
    return
  }

  if (payload.type === 'mock:all-picked') {
    sendEvent<AllPickedBody>(connection, 'AllPicked', { pickedBalls: mockPickedBalls })
  }
}

export const roomWebSocketHandler = roomSocket.addEventListener(
  'connection',
  ({ client, params }) => {
    const roomId = pathParam(params.roomId)
    const mode = client.url.searchParams.get('mode')
    const userId = client.url.searchParams.get('userId')?.trim() || 'mumumu'

    if (roomId === undefined) {
      client.close(1008, 'Room not found')
      return
    }

    if (mode !== 'participant' && mode !== 'display') {
      client.close(1008, 'Invalid mode')
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

    sendEvent(connection, 'Initialized', initializedBody(connection))
    scheduleDemoChatMessages(roomId)
  },
)
