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

function appendActivityLog(roomId: string, entries: ActivityLogEntry[]): void {
  activityLogsByRoom.set(roomId, [
    ...new Map(
      [...(activityLogsByRoom.get(roomId) ?? []), ...entries].map((entry) => [entry.id, entry]),
    ).values(),
  ])
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
  const pickedBalls = room.state === 'waiting' ? [] : initialPickedBalls

  if (connection.mode === 'participant') {
    const participantIndex = room.participants.findIndex(
      (participant) => participant.user.userId === connection.userId,
    )
    const card =
      room.state === 'waiting' || participantIndex < 0
        ? undefined
        : participantCard(
            connection,
            cardWithPickedBalls(mockCards[participantIndex % mockCards.length]!, pickedBalls),
          )

    return {
      type: 'Initialized',
      body: {
        activityLog: activityLogsByRoom.get(connection.roomId) ?? [],
        state: room.state,
        settings: room.settings,
        pickState: room.pickState,
        pickedBalls,
        bingoSummaries: room.bingoSummaries,
        reachSummaries: room.reachSummaries,
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
      pickState: room.pickState,
      participantCount: room.participants.length,
      pickedBalls,
      qrCodeVisible: room.qrCodeVisible,
      bingoSummaries: room.bingoSummaries,
      reachSummaries: room.reachSummaries,
    },
  }
}

function scheduleScript(
  connection: MockSocketConnection,
  script: MockRoomWebSocketScript['onConnect'],
): void {
  for (const item of script) {
    window.setTimeout(() => {
      const event = eventForMode(item.event, connection.mode)
      if (event.type === 'PickFinished') {
        const body = event.body as { newActivityLog: ActivityLogEntry[] }
        const knownIds = new Set(activityLogIdsByConnection.get(connection) ?? [])
        const newActivityLog = body.newActivityLog.filter((entry) => {
          if (knownIds.has(entry.id)) return false
          knownIds.add(entry.id)
          return true
        })
        const delivered = sendEventToConnection(
          connection,
          eventForConnection(connection, { ...event, body: { ...body, newActivityLog } }),
        )
        if (!delivered) return
        appendActivityLog(connection.roomId, body.newActivityLog)
        activityLogIdsByConnection.set(connection, knownIds)
        return
      }
      sendEventToConnection(connection, eventForConnection(connection, event))
    }, item.delayMs)
  }
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
