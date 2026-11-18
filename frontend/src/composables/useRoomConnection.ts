import { onBeforeUnmount, watch, type WatchSource } from 'vue'

import type { RoomCode, WebSocketMode } from '@/api/schema'
import { useRoomsStore } from '@/stores/rooms'
import { useRoomWebSocketStore } from '@/stores/roomWebSocket'

export function useRoomConnection(
  roomCode: WatchSource<RoomCode | undefined>,
  options: { mode: WebSocketMode; join?: boolean },
) {
  const roomsStore = useRoomsStore()
  const roomWebSocketStore = useRoomWebSocketStore()

  watch(
    roomCode,
    (code) => {
      void roomsStore.connectToRoom(code, options)
    },
    { immediate: true },
  )

  onBeforeUnmount(() => {
    roomWebSocketStore.disconnect()
  })
}
