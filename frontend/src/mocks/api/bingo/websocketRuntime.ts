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

function appendActivityLog(roomId: string, entries: ActivityLogEntry[]): void {
  activityLogsByRoom.set(roomId, [
    ...new Map(
      [...(activityLogsByRoom.get(roomId) ?? []), ...entries].map((entry) => [entry.id, entry]),
    ).values(),
  ])
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
        appendActivityLog(connection.roomId, body.newActivityLog)
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
