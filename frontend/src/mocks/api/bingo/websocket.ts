import type {
  AllPickedBody,
  Card,
  DisplayGameFinishedBody,
  DisplayGameStartedBody,
  DisplayInitializedBody,
  DisplayPickFinishedBody,
  HideQrCodeBody,
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
  ShowQrCodeBody,
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
const demoBingoLine = [1, 6, 11, 16, 21]

const demoPickSteps = [
  {
    delayMs: 1800,
    pickedBall: 17,
    newReachUserIds: ['kurosaki', 'rurun'],
    newBingoUserIds: [],
  },
  {
    delayMs: 4200,
    pickedBall: 22,
    newReachUserIds: ['howard127', 'kurao'],
    newBingoUserIds: ['kurosaki'],
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

const mockCards: Card[] = [
  {
    cardId: '00000000-0000-4000-8000-000000000201',
    cardNumber: '111111111111111111111111111111111111',
    ownerUserId: 'demo',
    cells: [
      { index: 0, number: 1, displayText: '1', cellState: 'closed' },
      { index: 1, number: 16, displayText: '16', cellState: 'closed' },
      { index: 2, number: 31, displayText: '31', cellState: 'closed' },
      { index: 3, number: 46, displayText: '46', cellState: 'closed' },
      { index: 4, number: 61, displayText: '61', cellState: 'closed' },

      { index: 5, number: 2, displayText: '2', cellState: 'closed' },
      { index: 6, number: 17, displayText: '17', cellState: 'closed' },
      { index: 7, number: 32, displayText: '32', cellState: 'closed' },
      { index: 8, number: 47, displayText: '47', cellState: 'closed' },
      { index: 9, number: 62, displayText: '62', cellState: 'closed' },

      { index: 10, number: 3, displayText: '3', cellState: 'closed' },
      { index: 11, number: 18, displayText: '18', cellState: 'closed' },
      { index: 12, number: null, displayText: 'FREE', cellState: 'open' },
      { index: 13, number: 48, displayText: '48', cellState: 'closed' },
      { index: 14, number: 63, displayText: '63', cellState: 'closed' },

      { index: 15, number: 4, displayText: '4', cellState: 'closed' },
      { index: 16, number: 19, displayText: '19', cellState: 'closed' },
      { index: 17, number: 34, displayText: '34', cellState: 'closed' },
      { index: 18, number: 49, displayText: '49', cellState: 'closed' },
      { index: 19, number: 64, displayText: '64', cellState: 'closed' },

      { index: 20, number: 5, displayText: '5', cellState: 'closed' },
      { index: 21, number: 20, displayText: '20', cellState: 'closed' },
      { index: 22, number: 35, displayText: '35', cellState: 'closed' },
      { index: 23, number: 50, displayText: '50', cellState: 'closed' },
      { index: 24, number: 65, displayText: '65', cellState: 'closed' },
    ],
    bingoLines: [],
    reachLines: [],
  },

  {
    cardId: '00000000-0000-4000-8000-000000000202',
    cardNumber: '222222222222222222222222222222222222',
    ownerUserId: 'demo',
    cells: [
      { index: 0, number: 1, displayText: '1', cellState: 'closed' },
      { index: 1, number: 16, displayText: '16', cellState: 'closed' },
      { index: 2, number: 31, displayText: '31', cellState: 'closed' },
      { index: 3, number: 46, displayText: '46', cellState: 'closed' },
      { index: 4, number: 61, displayText: '61', cellState: 'closed' },

      { index: 5, number: 2, displayText: '2', cellState: 'closed' },
      { index: 6, number: 17, displayText: '17', cellState: 'open' },
      { index: 7, number: 32, displayText: '32', cellState: 'closed' },
      { index: 8, number: 47, displayText: '47', cellState: 'closed' },
      { index: 9, number: 62, displayText: '62', cellState: 'closed' },

      { index: 10, number: 3, displayText: '3', cellState: 'closed' },
      { index: 11, number: 18, displayText: '18', cellState: 'closed' },
      { index: 12, number: null, displayText: 'FREE', cellState: 'open' },
      { index: 13, number: 48, displayText: '48', cellState: 'closed' },
      { index: 14, number: 63, displayText: '63', cellState: 'closed' },

      { index: 15, number: 4, displayText: '4', cellState: 'closed' },
      { index: 16, number: 19, displayText: '19', cellState: 'closed' },
      { index: 17, number: 34, displayText: '34', cellState: 'closed' },
      { index: 18, number: 49, displayText: '49', cellState: 'closed' },
      { index: 19, number: 64, displayText: '64', cellState: 'closed' },

      { index: 20, number: 5, displayText: '5', cellState: 'closed' },
      { index: 21, number: 20, displayText: '20', cellState: 'closed' },
      { index: 22, number: 35, displayText: '35', cellState: 'closed' },
      { index: 23, number: 50, displayText: '50', cellState: 'closed' },
      { index: 24, number: 65, displayText: '65', cellState: 'closed' },
    ],
    bingoLines: [],
    reachLines: [],
  },

  {
    cardId: '00000000-0000-4000-8000-000000000203',
    cardNumber: '333333333333333333333333333333333333',
    ownerUserId: 'demo',
    cells: [
      { index: 0, number: 1, displayText: '1', cellState: 'closed' },
      { index: 1, number: 16, displayText: '16', cellState: 'closed' },
      { index: 2, number: 31, displayText: '31', cellState: 'closed' },
      { index: 3, number: 46, displayText: '46', cellState: 'closed' },
      { index: 4, number: 61, displayText: '61', cellState: 'closed' },

      { index: 5, number: 2, displayText: '2', cellState: 'closed' },
      { index: 6, number: 17, displayText: '17', cellState: 'reach' },
      { index: 7, number: 32, displayText: '32', cellState: 'closed' },
      { index: 8, number: 47, displayText: '47', cellState: 'closed' },
      { index: 9, number: 62, displayText: '62', cellState: 'closed' },

      { index: 10, number: 3, displayText: '3', cellState: 'closed' },
      { index: 11, number: 18, displayText: '18', cellState: 'reach' },
      { index: 12, number: null, displayText: 'FREE', cellState: 'open' },
      { index: 13, number: 48, displayText: '48', cellState: 'closed' },
      { index: 14, number: 63, displayText: '63', cellState: 'closed' },

      { index: 15, number: 4, displayText: '4', cellState: 'closed' },
      { index: 16, number: 19, displayText: '19', cellState: 'reach' },
      { index: 17, number: 34, displayText: '34', cellState: 'closed' },
      { index: 18, number: 49, displayText: '49', cellState: 'closed' },
      { index: 19, number: 64, displayText: '64', cellState: 'closed' },

      { index: 20, number: 5, displayText: '5', cellState: 'closed' },
      { index: 21, number: 20, displayText: '20', cellState: 'reach' },
      { index: 22, number: 35, displayText: '35', cellState: 'closed' },
      { index: 23, number: 50, displayText: '50', cellState: 'closed' },
      { index: 24, number: 65, displayText: '65', cellState: 'closed' },
    ],
    bingoLines: [],
    reachLines: [demoBingoLine],
  },

  {
    cardId: '00000000-0000-4000-8000-000000000204',
    cardNumber: '444444444444444444444444444444444444',
    ownerUserId: 'demo',
    cells: [
      { index: 0, number: 1, displayText: '1', cellState: 'closed' },
      { index: 1, number: 16, displayText: '16', cellState: 'bingo' },
      { index: 2, number: 31, displayText: '31', cellState: 'closed' },
      { index: 3, number: 46, displayText: '46', cellState: 'closed' },
      { index: 4, number: 61, displayText: '61', cellState: 'open' },

      { index: 5, number: 2, displayText: '2', cellState: 'closed' },
      { index: 6, number: 17, displayText: '17', cellState: 'bingo' },
      { index: 7, number: 32, displayText: '32', cellState: 'closed' },
      { index: 8, number: 47, displayText: '47', cellState: 'closed' },
      { index: 9, number: 62, displayText: '62', cellState: 'closed' },

      { index: 10, number: 3, displayText: '3', cellState: 'closed' },
      { index: 11, number: 18, displayText: '18', cellState: 'bingo' },
      { index: 12, number: null, displayText: 'FREE', cellState: 'open' },
      { index: 13, number: 48, displayText: '48', cellState: 'closed' },
      { index: 14, number: 63, displayText: '63', cellState: 'closed' },

      { index: 15, number: 4, displayText: '4', cellState: 'closed' },
      { index: 16, number: 19, displayText: '19', cellState: 'bingo' },
      { index: 17, number: 34, displayText: '34', cellState: 'closed' },
      { index: 18, number: 49, displayText: '49', cellState: 'closed' },
      { index: 19, number: 64, displayText: '64', cellState: 'closed' },

      { index: 20, number: 5, displayText: '5', cellState: 'closed' },
      { index: 21, number: 20, displayText: '20', cellState: 'bingo' },
      { index: 22, number: 35, displayText: '35', cellState: 'closed' },
      { index: 23, number: 50, displayText: '50', cellState: 'closed' },
      { index: 24, number: 65, displayText: '65', cellState: 'closed' },
    ],
    bingoLines: [demoBingoLine],
    reachLines: [],
  },
]

let mockCardIndex = 0

function currentMockCard(): Card {
  return mockCards[Math.min(mockCardIndex, mockCards.length - 1)]!
}

function nextMockCard(): Card {
  mockCardIndex += 1
  return currentMockCard()
}

function demoPickStepIndex(step: (typeof demoPickSteps)[number]): number {
  const index = demoPickSteps.indexOf(step)
  return index === -1 ? 0 : index
}

function cardChangesForDemoPickStep(step: (typeof demoPickSteps)[number]) {
  const stepIndex = demoPickStepIndex(step)

  return {
    openedCellIndices: [[6], [11, 16, 21], [1], [4]][stepIndex] ?? [],
    newReachLines: stepIndex === 1 ? [demoBingoLine] : [],
    newBingoLines: stepIndex === 2 ? [demoBingoLine] : [],
  }
}

function pickedBallsForDemoPickStep(step: (typeof demoPickSteps)[number]): PickedBall[] {
  const stepIndex = demoPickStepIndex(step)
  const pickedBalls = demoPickSteps.slice(0, stepIndex + 1).map((candidate) => candidate.pickedBall)

  return [...new Set([...initialPickedBalls, ...pickedBalls])]
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
    pickedBalls: pickedBallsForDemoPickStep(step),
    newReaches,
    newBingos,
  }
}

function initializedBody(
  connection: MockSocketConnection,
): ParticipantInitializedBody | DisplayInitializedBody {
  const room = roomByPathParam(connection.roomId)
  const pickedBalls = room.state === 'waiting' ? [] : initialPickedBalls

  if (connection.mode === 'participant') {
    return {
      state: room.state,
      settings: room.settings,
      pickState: room.pickState,
      pickedBalls,
      bingoSummaries: room.bingoSummaries,
      reachSummaries: room.reachSummaries,
      ...(room.state === 'waiting' ? {} : { card: currentMockCard() }),
    }
  }

  return {
    state: room.state,
    settings: room.settings,
    pickState: room.pickState,
    participantCount: room.participants.length,
    pickedBalls,
    qrCodeVisible: room.qrCodeVisible,
    bingoSummaries: room.bingoSummaries,
    reachSummaries: room.reachSummaries,
  }
}

function sendGameStarted(connection: MockSocketConnection): void {
  const room = roomByPathParam(connection.roomId)

  if (connection.mode === 'participant') {
    sendEvent<ParticipantGameStartedBody>(connection, 'GameStarted', { card: currentMockCard() })
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
      card: nextMockCard(),
      cardChanges: cardChangesForDemoPickStep(step),
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
      card: currentMockCard(),
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
    if (mode === 'display' && roomByPathParam(roomId).state === 'waiting') {
      window.setTimeout(() => {
        sendEvent<ShowQrCodeBody>(connection, 'ShowQRCode', {})
      }, 1000)
      window.setTimeout(() => {
        sendEvent<HideQrCodeBody>(connection, 'HideQRCode', {})
      }, 5000)
    }
    scheduleDemoChatMessages(roomId)
    if (roomByPathParam(roomId).state !== 'waiting') {
      scheduleDemoPickEvents(connection)
    }
  },
)
