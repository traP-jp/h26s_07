import { ws } from 'msw'

import {
  clearSocketConnection,
  pathParam,
  setSocketConnection,
  type MockSocketConnection,
} from './core'
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
