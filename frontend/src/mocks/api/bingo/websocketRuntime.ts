import type { ActivityLogEntry } from '@/api/schema'
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

const scheduledDemoChatRoomIds = new Set<string>()
const activityLogsByRoom = new Map<string, ActivityLogEntry[]>()
const activityLogIdsByConnection = new WeakMap<MockSocketConnection, Set<string>>()

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

function initializedEvent(connection: MockSocketConnection): MockWebSocketEvent {
  const room = roomByPathParam(connection.roomId)
  const pickedBalls = room.state === 'waiting' ? [] : initialPickedBalls

  if (connection.mode === 'participant') {
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
        const newActivityLog = orderedActivityLog(
          body.newActivityLog.filter((entry) => {
            if (knownIds.has(entry.id)) return false
            knownIds.add(entry.id)
            return true
          }),
        )
        const orderedEvent = {
          ...event,
          body: { ...body, newActivityLog },
        }
        if (!sendEventToConnection(connection, orderedEvent)) return
        appendActivityLog(connection.roomId, body.newActivityLog)
        activityLogIdsByConnection.set(connection, knownIds)
        return
      }
      sendEventToConnection(connection, event)
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

  sendEventToConnection(connection, initializedEvent(connection))
  scheduleDemoChatMessages(connection)

  if (room.state !== 'waiting' || connection.mode === 'display') {
    scheduleScript(connection, script.onConnect)
  }
}
