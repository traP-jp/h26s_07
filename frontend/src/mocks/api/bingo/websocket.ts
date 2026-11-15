import { ws } from 'msw'

import {
  clearSocketConnection,
  pathParam,
  setSocketConnection,
  type MockSocketConnection,
} from './core'
import { mockRooms } from './fixtures'
import { startMockRoomWebSocket } from './websocketRuntime'

const roomSocket = ws.link('*/api/rooms/:roomId/ws')

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

    const room = mockRooms.find((item) => item.roomId === roomId || item.roomCode === roomId)
    if (!room) {
      client.close(1008, 'Room not found')
      return
    }

    if (mode !== 'participant' && mode !== 'display') {
      client.close(1008, 'Invalid mode')
      return
    }

    if (
      mode === 'participant' &&
      !room.participants.some((participant) => participant.user.userId === userId)
    ) {
      client.close(1008, 'Participant not found')
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

    setSocketConnection(connection)
    client.addEventListener(
      'close',
      () => {
        clearSocketConnection(connection)
      },
      { once: true },
    )

    startMockRoomWebSocket(connection)
  },
)
