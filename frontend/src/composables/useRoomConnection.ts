import { onBeforeUnmount, watch, type WatchSource } from 'vue'

import type { RoomCode, WebSocketMode } from '@/api/schema'
import { useRoomWebSocketStore } from '@/stores/roomWebSocket'

export function useRoomConnection(
  roomCode: WatchSource<RoomCode | undefined>,
  options: { mode: WebSocketMode; join?: boolean },
) {
  const roomWebSocketStore = useRoomWebSocketStore()

  watch(
    roomCode,
    (code) => {
      void roomWebSocketStore.connectToRoom(code, options)
    },
    { immediate: true },
  )

  onBeforeUnmount(() => {
    roomWebSocketStore.disconnect()
  })
}
