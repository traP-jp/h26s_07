import type { Card, Message, PickedBall, WebSocketEventType } from '@/api/schema'
import type { MockSocketConnection } from '../core'
import { fallbackRoom, mockRooms } from './rooms'

export const initialPickedBalls: PickedBall[] = [4, 18, 33]
export const demoBingoLine = [1, 6, 11, 16, 21]

export type MockWebSocketEvent<TBody = unknown> = {
  type: WebSocketEventType
  body: TBody
}

export type ModeWebSocketEvent = {
  participant: MockWebSocketEvent
  display: MockWebSocketEvent
}

export type MockWebSocketScriptEvent = MockWebSocketEvent | ModeWebSocketEvent

export type MockWebSocketScriptItem = {
  delayMs: number
  event: MockWebSocketScriptEvent
}

export type MockRoomWebSocketScript = {
  onConnect: MockWebSocketScriptItem[]
}

export const mockCards: Card[] = [
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

const demoUsers = {
  kurosaki: { userId: 'kurosaki' },
  rurun: { userId: 'rurun' },
  howard127: { userId: 'howard127' },
  kurao: { userId: 'kurao' },
  minami: { userId: 'minami' },
  yamada: { userId: 'yamada' },
} as const

const emptyCardChanges = {
  openedCellIndices: [],
  newReachLines: [],
  newBingoLines: [],
}

const pickSnapshots = [
  {
    pickedBall: 17,
    pickedBalls: [...initialPickedBalls, 17],
    card: mockCards[1]!,
    cardChanges: {
      openedCellIndices: [6],
      newReachLines: [],
      newBingoLines: [],
    },
    bingoSummaries: [],
    reachSummaries: [
      { user: demoUsers.kurosaki, createdAt: '2026-06-27T11:11:00.000Z' },
      { user: demoUsers.rurun, createdAt: '2026-06-27T11:11:00.000Z' },
    ],
    newBingos: [],
    newReaches: [{ user: demoUsers.kurosaki }, { user: demoUsers.rurun }],
  },
  {
    pickedBall: 22,
    pickedBalls: [...initialPickedBalls, 17, 22],
    card: mockCards[2]!,
    cardChanges: {
      openedCellIndices: [11, 16, 21],
      newReachLines: [demoBingoLine],
      newBingoLines: [],
    },
    bingoSummaries: [
      { user: demoUsers.kurosaki, bingoOrders: [1], createdAt: '2026-06-27T11:12:00.000Z' },
    ],
    reachSummaries: [
      { user: demoUsers.rurun, createdAt: '2026-06-27T11:11:00.000Z' },
      { user: demoUsers.howard127, createdAt: '2026-06-27T11:12:00.000Z' },
      { user: demoUsers.kurao, createdAt: '2026-06-27T11:12:00.000Z' },
    ],
    newBingos: [{ user: demoUsers.kurosaki, newBingoOrders: [1], bingoOrders: [1] }],
    newReaches: [{ user: demoUsers.howard127 }, { user: demoUsers.kurao }],
  },
  {
    pickedBall: 45,
    pickedBalls: [...initialPickedBalls, 17, 22, 45],
    card: mockCards[3]!,
    cardChanges: {
      openedCellIndices: [1],
      newReachLines: [],
      newBingoLines: [demoBingoLine],
    },
    bingoSummaries: [
      { user: demoUsers.kurosaki, bingoOrders: [1], createdAt: '2026-06-27T11:12:00.000Z' },
      { user: demoUsers.rurun, bingoOrders: [2], createdAt: '2026-06-27T11:13:00.000Z' },
    ],
    reachSummaries: [
      { user: demoUsers.howard127, createdAt: '2026-06-27T11:12:00.000Z' },
      { user: demoUsers.kurao, createdAt: '2026-06-27T11:12:00.000Z' },
      { user: demoUsers.minami, createdAt: '2026-06-27T11:13:00.000Z' },
    ],
    newBingos: [{ user: demoUsers.rurun, newBingoOrders: [2], bingoOrders: [2] }],
    newReaches: [{ user: demoUsers.minami }],
  },
  {
    pickedBall: 61,
    pickedBalls: [...initialPickedBalls, 17, 22, 45, 61],
    card: mockCards[3]!,
    cardChanges: {
      openedCellIndices: [4],
      newReachLines: [],
      newBingoLines: [],
    },
    bingoSummaries: [
      { user: demoUsers.kurosaki, bingoOrders: [1], createdAt: '2026-06-27T11:12:00.000Z' },
      { user: demoUsers.rurun, bingoOrders: [2], createdAt: '2026-06-27T11:13:00.000Z' },
      { user: demoUsers.howard127, bingoOrders: [3], createdAt: '2026-06-27T11:14:00.000Z' },
      { user: demoUsers.kurao, bingoOrders: [4], createdAt: '2026-06-27T11:14:00.000Z' },
    ],
    reachSummaries: [
      { user: demoUsers.minami, createdAt: '2026-06-27T11:13:00.000Z' },
      { user: demoUsers.yamada, createdAt: '2026-06-27T11:14:00.000Z' },
    ],
    newBingos: [
      { user: demoUsers.howard127, newBingoOrders: [3], bingoOrders: [3] },
      { user: demoUsers.kurao, newBingoOrders: [4], bingoOrders: [4] },
    ],
    newReaches: [{ user: demoUsers.yamada }],
  },
]

const lastPickSnapshot = pickSnapshots[3]!

const pickFinishedEvent = (snapshot: (typeof pickSnapshots)[number]): ModeWebSocketEvent => ({
  participant: {
    type: 'PickFinished',
    body: {
      pickedBall: snapshot.pickedBall,
      pickState: 'idle',
      card: snapshot.card,
      cardChanges: snapshot.cardChanges,
      pickedBalls: snapshot.pickedBalls,
      bingoSummaries: snapshot.bingoSummaries,
      reachSummaries: snapshot.reachSummaries,
      newBingos: snapshot.newBingos,
      newReaches: snapshot.newReaches,
    },
  },
  display: {
    type: 'PickFinished',
    body: {
      pickedBall: snapshot.pickedBall,
      pickState: 'idle',
      participantCount: 6,
      bingoSummaries: snapshot.bingoSummaries,
      reachSummaries: snapshot.reachSummaries,
      newBingos: snapshot.newBingos,
      newReaches: snapshot.newReaches,
      pickedBalls: snapshot.pickedBalls,
    },
  },
})

const pickStartedEvent: MockWebSocketEvent = { type: 'PickStarted', body: {} }

export const waitingWebSocketScript: MockRoomWebSocketScript = {
  onConnect: [
    { delayMs: 1000, event: { type: 'ShowQRCode', body: {} } },
    { delayMs: 5000, event: { type: 'HideQRCode', body: {} } },
  ],
}

export const playingWebSocketScript: MockRoomWebSocketScript = {
  onConnect: [
    {
      delayMs: 0,
      event: {
        participant: { type: 'GameStarted', body: { card: mockCards[0]! } },
        display: { type: 'GameStarted', body: { participantCount: 6 } },
      },
    },
    { delayMs: 1800, event: pickStartedEvent },
    { delayMs: 2500, event: pickFinishedEvent(pickSnapshots[0]!) },
    { delayMs: 4200, event: pickStartedEvent },
    { delayMs: 4900, event: pickFinishedEvent(pickSnapshots[1]!) },
    { delayMs: 6600, event: pickStartedEvent },
    { delayMs: 7300, event: pickFinishedEvent(pickSnapshots[2]!) },
    { delayMs: 9000, event: pickStartedEvent },
    { delayMs: 9700, event: pickFinishedEvent(pickSnapshots[3]!) },
    {
      delayMs: 11000,
      event: { type: 'PickStarted', body: {} },
    },
    {
      delayMs: 11700,
      event: pickFinishedEvent({
        ...lastPickSnapshot,
        pickedBall: 2,
        pickedBalls: [...lastPickSnapshot.pickedBalls, 2],
        cardChanges: emptyCardChanges,
        newBingos: [],
        newReaches: [],
      }),
    },
    { delayMs: 13000, event: pickStartedEvent },
    {
      delayMs: 13700,
      event: pickFinishedEvent({
        ...lastPickSnapshot,
        pickedBall: 24,
        pickedBalls: [...lastPickSnapshot.pickedBalls, 2, 24],
        cardChanges: emptyCardChanges,
        newBingos: [],
        newReaches: [],
      }),
    },
    { delayMs: 15000, event: pickStartedEvent },
    {
      delayMs: 15700,
      event: pickFinishedEvent({
        ...lastPickSnapshot,
        pickedBall: 28,
        pickedBalls: [...lastPickSnapshot.pickedBalls, 2, 24, 28],
        cardChanges: emptyCardChanges,
        newBingos: [],
        newReaches: [],
      }),
    },
    { delayMs: 17000, event: pickStartedEvent },
    {
      delayMs: 17700,
      event: pickFinishedEvent({
        ...lastPickSnapshot,
        pickedBall: 32,
        pickedBalls: [...lastPickSnapshot.pickedBalls, 2, 24, 28, 32],
        cardChanges: emptyCardChanges,
        newBingos: [],
        newReaches: [],
      }),
    },
  ],
}

export const finishedWebSocketScript: MockRoomWebSocketScript = {
  onConnect: [
    {
      delayMs: 0,
      event: {
        participant: {
          type: 'GameFinished',
          body: {
            state: 'finished',
            pickState: 'idle',
            card: lastPickSnapshot.card,
            bingoSummaries: lastPickSnapshot.bingoSummaries,
            reachSummaries: lastPickSnapshot.reachSummaries,
          },
        },
        display: {
          type: 'GameFinished',
          body: {
            state: 'finished',
            pickState: 'idle',
            participantCount: 6,
            bingoSummaries: lastPickSnapshot.bingoSummaries,
            reachSummaries: lastPickSnapshot.reachSummaries,
          },
        },
      },
    },
  ],
}

export const roomWebSocketScripts: Record<string, MockRoomWebSocketScript> = {
  '111111': waitingWebSocketScript,
  '123456': waitingWebSocketScript,
  '234567': playingWebSocketScript,
  '345678': finishedWebSocketScript,
  '456789': waitingWebSocketScript,
  '567890': playingWebSocketScript,
  '678901': finishedWebSocketScript,
  '789012': waitingWebSocketScript,
  '890123': playingWebSocketScript,
  '901234': finishedWebSocketScript,
  '012345': waitingWebSocketScript,
}

export function roomByPathParam(roomParam: string) {
  return (
    mockRooms.find((room) => room.roomId === roomParam || room.roomCode === roomParam) ??
    fallbackRoom
  )
}

export function eventForMode(
  event: MockWebSocketScriptEvent,
  mode: MockSocketConnection['mode'],
): MockWebSocketEvent {
  return 'participant' in event ? event[mode] : event
}

export function scriptForRoom(room: ReturnType<typeof roomByPathParam>): MockRoomWebSocketScript {
  return roomWebSocketScripts[room.roomCode] ?? waitingWebSocketScript
}

export const demoChatMessages = [
  {
    author: { userId: 'mumumu' },
    content: 'いい感じに進んでる',
  },
  {
    author: { userId: 'rurun' },
    content: 'リーチきた！',
  },
  {
    author: { userId: 'kurao' },
    content: '次こそビンゴしたい',
  },
  {
    author: { userId: 'howard127' },
    content: 'あと1個でいけそう',
  },
  {
    author: { userId: 'kurosaki' },
    content: '今の演出めっちゃ目立つ',
  },
  {
    author: { userId: 'howard127' },
    content: 'B列ぜんぜん来ない',
  },
  {
    author: { userId: 'rurun' },
    content: 'ビンゴした人おめでとう！',
  },
  {
    author: { userId: 'mumumu' },
    content: '次の抽選お願いします',
  },
] satisfies Pick<Message, 'author' | 'content'>[]
