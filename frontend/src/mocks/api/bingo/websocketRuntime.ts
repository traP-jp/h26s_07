import type { ActivityLogEntry } from '@/api/schema'
import { sendEvent, type MockSocketConnection } from './core'
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

function initializedEvent(connection: MockSocketConnection): MockWebSocketEvent {
  const room = roomByPathParam(connection.roomId)
  const pickedBalls = room.state === 'waiting' ? [] : initialPickedBalls

  if (connection.mode === 'participant') {
    return {
      type: 'Initialized',
      body: {
        activityLog: room.activityLog,
        participants: room.participants.map((participant) => participant.user),
        participantCount: room.participants.length,
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
      activityLog: room.activityLog,
      participants: room.participants.map((participant) => participant.user),
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
        const room = roomByPathParam(connection.roomId)
        const body = event.body as { newActivityLog: ActivityLogEntry[] }
        room.activityLog = [
          ...new Map(
            [...room.activityLog, ...body.newActivityLog].map((entry) => [entry.id, entry]),
          ).values(),
        ]
      }
      sendEvent(event, connection.roomId)
    }, item.delayMs)
  }
}

function scheduleDemoChatMessages(roomId: string): void {
  if (scheduledDemoChatRoomIds.has(roomId)) {
    return
  }

  scheduledDemoChatRoomIds.add(roomId)

  for (const [index, messageTemplate] of demoChatMessages.entries()) {
    window.setTimeout(
      () => {
        sendEvent({
          type: 'MessageCreated',
          body: {
            message: {
              ...messageTemplate,
              messageId: crypto.randomUUID(),
              createdAt: new Date().toISOString(),
            },
          },
        })
      },
      2000 + index * 2000,
    )
  }
}

export function startMockRoomWebSocket(connection: MockSocketConnection): void {
  const room = roomByPathParam(connection.roomId)
  const script = scriptForRoom(room)

  sendEvent(initializedEvent(connection))
  scheduleDemoChatMessages(connection.roomId)

  if (room.state !== 'waiting' || connection.mode === 'display') {
    scheduleScript(connection, script.onConnect)
  }
}
