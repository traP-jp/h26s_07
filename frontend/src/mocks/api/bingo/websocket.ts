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
  User,
  WebSocketEventType,
} from '@/api/schema'
import { ws } from 'msw'

import { pathParam, socketConnections, type MockSocketConnection } from './core'
import { fallbackRoom, mockRooms } from './rooms'

type MockWebSocketEvent<TBody> = {
  type: WebSocketEventType
  body: TBody
}

const roomSocket = ws.link('*/api/rooms/:roomId/ws')
const initialPickedBalls: PickedBall[] = [4, 18, 33]
const scheduledDemoPickRoomIds = new Set<string>()

const demoPickSteps = [
  {
    delayMs: 1800,
    pickedBall: 7,
    newReachUserIds: ['saba', 'rurun'],
    newBingoUserIds: [],
  },
  {
    delayMs: 4200,
    pickedBall: 22,
    newReachUserIds: ['howard127', 'kurao'],
    newBingoUserIds: ['saba'],
  },
  {
    delayMs: 6600,
    pickedBall: 45,
    newReachUserIds: ['minami'],
    newBingoUserIds: ['rurun'],
  },
  {
    delayMs: 9000,
    pickedBall: 61,
    newReachUserIds: ['yamada'],
    newBingoUserIds: ['howard127', 'kurao'],
  },
] as const

const mockCard: Card = {
  cardId: '00000000-0000-4000-8000-000000000201',
  cardNumber: '000000000000000000000000000000000201',
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
      cellState: initialPickedBalls.includes(number) ? 'open' : 'closed',
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

function roomByPathParam(roomParam: string): Room {
  return (
    mockRooms.find((room) => room.roomId === roomParam || room.roomCode === roomParam) ??
    fallbackRoom
  )
}

function roomUser(room: Room, userId: string): User {
  return (
    room.participants.find((participant) => participant.user.userId === userId)?.user ?? {
      userId,
    }
  )
}

function addReachSummary(room: Room, user: User): void {
  const alreadyReached = room.reachSummaries.some((summary) => summary.user.userId === user.userId)
  const alreadyBingo = room.bingoSummaries.some((summary) => summary.user.userId === user.userId)

  if (!alreadyReached && !alreadyBingo) {
    room.reachSummaries.push({ user })
  }
}

function addBingoSummary(room: Room, user: User): number[] {
  room.reachSummaries = room.reachSummaries.filter((summary) => summary.user.userId !== user.userId)

  const summary = room.bingoSummaries.find((candidate) => candidate.user.userId === user.userId)
  const nextOrder =
    Math.max(0, ...room.bingoSummaries.flatMap((candidate) => candidate.bingoOrders)) + 1

  if (summary) {
    summary.bingoOrders.push(nextOrder)
    return summary.bingoOrders
  }

  const bingoOrders = [nextOrder]
  room.bingoSummaries.push({ user, bingoOrders })
  return bingoOrders
}

function applyDemoPickStep(room: Room, step: (typeof demoPickSteps)[number]) {
  const newReaches = step.newReachUserIds.map((userId) => {
    const user = roomUser(room, userId)
    addReachSummary(room, user)
    return { user }
  })

  const newBingos = step.newBingoUserIds.map((userId) => {
    const user = roomUser(room, userId)
    const previousOrders =
      room.bingoSummaries.find((summary) => summary.user.userId === userId)?.bingoOrders ?? []
    const bingoOrders = addBingoSummary(room, user)
    const newBingoOrders = bingoOrders.filter((order) => !previousOrders.includes(order))

    return {
      user,
      newBingoOrders,
      bingoOrders,
    }
  })

  room.state = 'playing'
  room.pickState = 'idle'
  room.updatedAt = new Date().toISOString()

  return {
    pickedBall: step.pickedBall,
    pickedBalls: [...new Set([...initialPickedBalls, step.pickedBall])],
    newReaches,
    newBingos,
  }
}

function initializedBody(
  connection: MockSocketConnection,
): ParticipantInitializedBody | DisplayInitializedBody {
  const room = roomByPathParam(connection.roomId)

  if (connection.mode === 'participant') {
    return {
      state: room.state,
      settings: room.settings,
      pickState: room.pickState,
      pickedBalls: initialPickedBalls,
      bingoSummaries: room.bingoSummaries,
      reachSummaries: room.reachSummaries,
      card: mockCard,
    }
  }

  return {
    state: room.state,
    settings: room.settings,
    pickState: room.pickState,
    participantCount: room.participants.length,
    pickedBalls: initialPickedBalls,
    qrCodeVisible: room.qrCodeVisible,
    bingoSummaries: room.bingoSummaries,
    reachSummaries: room.reachSummaries,
  }
}

function sendGameStarted(connection: MockSocketConnection): void {
  const room = roomByPathParam(connection.roomId)

  if (connection.mode === 'participant') {
    sendEvent<ParticipantGameStartedBody>(connection, 'GameStarted', { card: mockCard })
    return
  }

  sendEvent<DisplayGameStartedBody>(connection, 'GameStarted', {
    participantCount: room.participants.length,
  })
}

function sendPickFinished(
  connection: MockSocketConnection,
  step: (typeof demoPickSteps)[number] = demoPickSteps[0],
): void {
  const room = roomByPathParam(connection.roomId)
  const result = applyDemoPickStep(room, step)

  if (connection.mode === 'participant') {
    sendEvent<ParticipantPickFinishedBody>(connection, 'PickFinished', {
      pickedBall: result.pickedBall,
      pickState: 'idle',
      card: mockCard,
      cardChanges: {
        openedCellIndices: [0],
        newReachLines: [],
        newBingoLines: [],
      },
      pickedBalls: result.pickedBalls,
      bingoSummaries: room.bingoSummaries,
      reachSummaries: room.reachSummaries,
      newBingos: result.newBingos,
      newReaches: result.newReaches,
    })
    return
  }

  sendEvent<DisplayPickFinishedBody>(connection, 'PickFinished', {
    pickedBall: result.pickedBall,
    pickState: 'idle',
    participantCount: room.participants.length,
    bingoSummaries: room.bingoSummaries,
    reachSummaries: room.reachSummaries,
    newBingos: result.newBingos,
    newReaches: result.newReaches,
    pickedBalls: result.pickedBalls,
  })
}

function sendGameFinished(connection: MockSocketConnection): void {
  const room = roomByPathParam(connection.roomId)

  if (connection.mode === 'participant') {
    sendEvent<ParticipantGameFinishedBody>(connection, 'GameFinished', {
      state: 'finished',
      pickState: 'idle',
      card: mockCard,
      bingoSummaries: room.bingoSummaries,
      reachSummaries: room.reachSummaries,
    })
    return
  }

  sendEvent<DisplayGameFinishedBody>(connection, 'GameFinished', {
    state: 'finished',
    pickState: 'idle',
    participantCount: room.participants.length,
    bingoSummaries: room.bingoSummaries,
    reachSummaries: room.reachSummaries,
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

function scheduleDemoPickEvents(connection: MockSocketConnection): void {
  if (scheduledDemoPickRoomIds.has(connection.roomId)) {
    return
  }

  scheduledDemoPickRoomIds.add(connection.roomId)

  for (const step of demoPickSteps) {
    window.setTimeout(() => {
      sendEvent<PickStartedBody>(connection, 'PickStarted', {})
      window.setTimeout(() => {
        sendPickFinished(connection, step)
      }, 700)
    }, step.delayMs)
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
    sendEvent<AllPickedBody>(connection, 'AllPicked', { pickedBalls: initialPickedBalls })
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
    if (mode === 'display') {
      scheduleDemoPickEvents(connection)
    }
  },
)
