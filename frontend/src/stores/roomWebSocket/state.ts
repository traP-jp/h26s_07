import type {
  ActivityLogEntry,
  BingoSummary,
  BingoUpdate,
  Card,
  CardChanges,
  GameSettings,
  Message,
  PickState,
  PickedBall,
  ReachSummary,
  ReachUpdate,
  RoomState,
  User,
} from '@/api/schema'

import type { RoomWebSocketEvent } from './types'

export type RoomWebSocketData = {
  initializationCount: number
  activityLog: ActivityLogEntry[]
  participants: User[]
  roomDeleted: boolean
  chatMessages: Message[]
  latestEvent: RoomWebSocketEvent | null
  latestBall: PickedBall | null
  latestCardChanges: CardChanges | null
  latestNewBingos: BingoUpdate[]
  latestNewReaches: ReachUpdate[]
  latestMessage: Message | null
  roomState: RoomState | null
  settings: GameSettings | null
  pickState: PickState | null
  pickedBalls: PickedBall[]
  bingoSummaries: BingoSummary[]
  reachSummaries: ReachSummary[]
  card: Card | null
  participantCount: number | null
  qrCodeVisible: boolean | null
}

export type ResettableRoomWebSocketData = Omit<RoomWebSocketData, 'initializationCount'>

export function createInitialRoomWebSocketData(): ResettableRoomWebSocketData {
  return {
    activityLog: [],
    participants: [],
    roomDeleted: false,
    chatMessages: [],
    latestEvent: null,
    latestBall: null,
    latestCardChanges: null,
    latestNewBingos: [],
    latestNewReaches: [],
    latestMessage: null,
    roomState: null,
    settings: null,
    pickState: null,
    pickedBalls: [],
    bingoSummaries: [],
    reachSummaries: [],
    card: null,
    participantCount: null,
    qrCodeVisible: null,
  }
}
