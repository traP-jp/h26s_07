import type { ActivityLogEntry, Card } from '@/api/schema'
import { sendEvent, sendEventToConnection, type MockSocketConnection } from './core'
import { appendMessageToRoom } from './chatHistory'
import {
  demoChatMessages,
  eventForMode,
  initialPickedBalls,
  type MockRoomWebSocketScript,
  type MockWebSocketEvent,
  roomByPathParam,
  scriptForRoom,
} from './fixtures/index'
import { mockCards } from './fixtures/websocket'

const scheduledDemoChatRoomIds = new Set<string>()
const activityLogsByRoom = new Map<string, ActivityLogEntry[]>()
const participantCardIds = new Map<string, string>()
const participantCardNumbers = new Map<string, string>()
const activityLogIdsByConnection = new WeakMap<MockSocketConnection, Set<string>>()
const scriptProgressByConnection = new Map<string, number>()
const latestPickByConnection = new Map<string, MockWebSocketEvent>()
const pickStateByConnection = new Map<string, 'idle' | 'picking'>()

function connectionKey(connection: MockSocketConnection): string {
  return `${connection.roomId}:${connection.mode}:${connection.userId}`
}

function orderedActivityLog(entries: ActivityLogEntry[]): ActivityLogEntry[] {
  return [...entries].sort((a, b) => {
    const timeA = /^(.*T\d\d:\d\d:\d\d)(?:\.(\d+))?(Z|[+-]\d\d:\d\d)$/.exec(a.createdAt)
    const timeB = /^(.*T\d\d:\d\d:\d\d)(?:\.(\d+))?(Z|[+-]\d\d:\d\d)$/.exec(b.createdAt)
    if (timeA && timeB) {
      const secondDifference =
        Date.parse(`${timeA[1]}${timeA[3]}`) - Date.parse(`${timeB[1]}${timeB[3]}`)
      if (secondDifference !== 0) return secondDifference
      const precision = Math.max(timeA[2]?.length ?? 0, timeB[2]?.length ?? 0)
      const fractionA = (timeA[2] ?? '').padEnd(precision, '0')
      const fractionB = (timeB[2] ?? '').padEnd(precision, '0')
      if (fractionA !== fractionB) return fractionA < fractionB ? -1 : 1
    } else {
      const timeDifference = Date.parse(a.createdAt) - Date.parse(b.createdAt)
      if (timeDifference !== 0) return timeDifference
    }
    return a.id < b.id ? -1 : a.id > b.id ? 1 : 0
  })
}

function appendActivityLog(roomId: string, entries: ActivityLogEntry[]): void {
  const uniqueEntries = new Map(
    [...(activityLogsByRoom.get(roomId) ?? []), ...entries].map((entry) => [entry.id, entry]),
  )
  activityLogsByRoom.set(roomId, orderedActivityLog([...uniqueEntries.values()]))
}

function participantCard(connection: MockSocketConnection, template: Card): Card {
  const key = `${connection.roomId}:${connection.userId}`
  let cardId = participantCardIds.get(key)
  if (!cardId) {
    cardId = crypto.randomUUID()
    participantCardIds.set(key, cardId)
    participantCardNumbers.set(key, template.cardNumber)
  }
  return {
    ...template,
    cardId,
    cardNumber: participantCardNumbers.get(key)!,
    ownerUserId: connection.userId,
  }
}

function cardWithPickedBalls(card: Card, pickedBalls: number[]): Card {
  const picked = new Set(pickedBalls)
  return {
    ...card,
    cells: card.cells.map((cell) =>
      cell.number !== null && picked.has(cell.number) && cell.cellState === 'closed'
        ? { ...cell, cellState: 'open' }
        : cell,
    ),
  }
}

export function eventForConnection(
  connection: MockSocketConnection,
  event: MockWebSocketEvent,
): MockWebSocketEvent {
  if (connection.mode !== 'participant') return event

  switch (event.type) {
    case 'GameStarted':
    case 'PickFinished':
    case 'GameFinished': {
      const body = event.body as { card: Card; pickedBalls?: number[] }
      const pickedBalls =
        body.pickedBalls ?? (event.type === 'GameStarted' ? initialPickedBalls : [])
      const card = cardWithPickedBalls(body.card, pickedBalls)
      return { ...event, body: { ...body, card: participantCard(connection, card) } }
    }
    default:
      return event
  }
}

function initializedEvent(connection: MockSocketConnection): MockWebSocketEvent {
  const room = roomByPathParam(connection.roomId)
  const key = connectionKey(connection)
  const latestPick = latestPickByConnection.get(key)
  const latestPickBody = latestPick?.body as
    | {
        card?: Card
        pickedBalls?: number[]
        bingoSummaries?: typeof room.bingoSummaries
        reachSummaries?: typeof room.reachSummaries
      }
    | undefined
  const pickedBalls =
    latestPickBody?.pickedBalls ?? (room.state === 'waiting' ? [] : initialPickedBalls)
  const bingoSummaries = latestPickBody?.bingoSummaries ?? room.bingoSummaries
  const reachSummaries = latestPickBody?.reachSummaries ?? room.reachSummaries

  if (connection.mode === 'participant') {
    const participantIndex = room.participants.findIndex(
      (participant) => participant.user.userId === connection.userId,
    )
    const card =
      room.state === 'waiting' || participantIndex < 0
        ? undefined
        : (latestPickBody?.card ??
          participantCard(
            connection,
            cardWithPickedBalls(mockCards[participantIndex % mockCards.length]!, pickedBalls),
          ))

    return {
      type: 'Initialized',
      body: {
        activityLog: activityLogsByRoom.get(connection.roomId) ?? [],
        state: room.state,
        settings: room.settings,
        pickState: pickStateByConnection.get(key) ?? room.pickState,
        pickedBalls,
        bingoSummaries,
        reachSummaries,
        ...(card ? { card } : {}),
      },
    }
  }

  return {
    type: 'Initialized',
    body: {
      participants: room.participants.map((participant) => participant.user),
      activityLog: activityLogsByRoom.get(connection.roomId) ?? [],
      state: room.state,
      settings: room.settings,
      pickState: pickStateByConnection.get(key) ?? room.pickState,
      participantCount: room.participants.length,
      pickedBalls,
      qrCodeVisible: room.qrCodeVisible,
      bingoSummaries,
      reachSummaries,
    },
  }
}

function scheduleScript(
  connection: MockSocketConnection,
  script: MockRoomWebSocketScript['onConnect'],
): void {
  const key = connectionKey(connection)
  const progress = scriptProgressByConnection.get(key) ?? 0
  script.forEach((item, index) => {
    if (index < progress) return
    window.setTimeout(() => {
      const event = eventForMode(item.event, connection.mode)
      if (event.type === 'PickFinished') {
        const body = event.body as { newActivityLog: ActivityLogEntry[] }
        const knownIds = new Set(activityLogIdsByConnection.get(connection) ?? [])
        const newActivityLog = orderedActivityLog(
          body.newActivityLog.filter((entry) => {
            if (knownIds.has(entry.id)) return false
            knownIds.add(entry.id)
            return true
          }),
        )
        const connectionEvent = eventForConnection(connection, {
          ...event,
          body: { ...body, newActivityLog },
        })
        const delivered = sendEventToConnection(connection, connectionEvent)
        if (!delivered) return
        appendActivityLog(connection.roomId, body.newActivityLog)
        activityLogIdsByConnection.set(connection, knownIds)
        latestPickByConnection.set(key, connectionEvent)
        pickStateByConnection.set(key, 'idle')
        scriptProgressByConnection.set(key, index + 1)
        return
      }
      if (sendEventToConnection(connection, eventForConnection(connection, event))) {
        if (event.type === 'PickStarted') pickStateByConnection.set(key, 'picking')
        scriptProgressByConnection.set(key, index + 1)
      }
    }, item.delayMs)
  })
}

function scheduleDemoChatMessages(connection: MockSocketConnection): void {
  const { roomId } = connection
  if (scheduledDemoChatRoomIds.has(roomId)) {
    return
  }

  scheduledDemoChatRoomIds.add(roomId)

  for (const [index, messageTemplate] of demoChatMessages.entries()) {
    window.setTimeout(
      () => {
        const message = {
          ...messageTemplate,
          messageId: crypto.randomUUID(),
          createdAt: new Date().toISOString(),
        }
        appendMessageToRoom(roomId, message)
        // 履歴へ保存したroomイベントなので、同じroomの再接続先にも配信する。
        sendEvent({ type: 'MessageCreated', body: { message } }, { roomId })
      },
      2000 + index * 2000,
    )
  }
}

export function startMockRoomWebSocket(connection: MockSocketConnection): void {
  const room = roomByPathParam(connection.roomId)
  const script = scriptForRoom(room)

  const event = initializedEvent(connection)
  const body = event.body as { activityLog: ActivityLogEntry[] }
  activityLogIdsByConnection.set(connection, new Set(body.activityLog.map((entry) => entry.id)))
  sendEventToConnection(connection, event)
  scheduleDemoChatMessages(connection)

  if (room.state !== 'waiting' || connection.mode === 'display') {
    scheduleScript(connection, script.onConnect)
  }
}
