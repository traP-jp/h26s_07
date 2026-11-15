import { sendEventToConnection, type MockSocketConnection } from './core'
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
      sendEventToConnection(connection, eventForMode(item.event, connection.mode))
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
        sendEventToConnection(connection, {
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

  sendEventToConnection(connection, initializedEvent(connection))
  scheduleDemoChatMessages(connection)

  if (room.state !== 'waiting' || connection.mode === 'display') {
    scheduleScript(connection, script.onConnect)
  }
}
