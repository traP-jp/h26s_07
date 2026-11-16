import { computed, ref } from 'vue'
import { storeToRefs } from 'pinia'

import type { RoomCode, RoomId, WebSocketMode } from '@/api/schema'
import { useCurrentUserStore } from '@/stores/currentUser'
import { useRoomsStore } from '@/stores/rooms'
import { useRoomWebSocketStore } from '@/stores/roomWebSocket'

type RoomConnectionOptions = {
  mode: WebSocketMode
  /** 未参加なら参加してから接続する */
  join?: boolean
}

export function useRoomConnection({ mode, join = false }: RoomConnectionOptions) {
  const roomsStore = useRoomsStore()
  const roomWebSocketStore = useRoomWebSocketStore()
  const currentUserStore = useCurrentUserStore()
  const { mode: connectedMode, roomId: connectedRoomId } = storeToRefs(roomWebSocketStore)

  const roomId = ref<RoomId | null>(null)
  const errorMessage = ref('')
  const isActiveRoom = computed(
    () => roomId.value !== null && connectedRoomId.value === roomId.value,
  )

  async function connectToRoom(roomCode: RoomCode | undefined): Promise<void> {
    if (connectedMode.value === mode) roomWebSocketStore.disconnect()
    roomId.value = null
    errorMessage.value = ''

    if (!roomCode) return

    try {
      const userId = join ? await currentUserStore.getUserId() : undefined
      const room = await roomsStore.getRoomByCode(roomCode)
      const nextRoomId = room?.roomId ?? null
      roomId.value = nextRoomId

      if (!nextRoomId) return

      if (join) {
        const isParticipant =
          room?.participants.some((participant) => participant.user.userId === userId) ?? false

        if (!isParticipant) {
          if (room?.state !== 'waiting') {
            errorMessage.value = 'このルームには参加できません。'
            return
          }

          await roomsStore.joinRoom(nextRoomId)
        }

        roomWebSocketStore.connect({ roomId: nextRoomId, mode, userId })
        return
      }

      if (!roomWebSocketStore.isActiveConnection({ roomId: nextRoomId, mode })) {
        roomWebSocketStore.connect({ roomId: nextRoomId, mode })
      }
    } catch (error) {
      errorMessage.value = error instanceof Error ? error.message : 'ルームに参加できませんでした。'
    }
  }

  return { roomId, isActiveRoom, errorMessage, connectToRoom }
}
