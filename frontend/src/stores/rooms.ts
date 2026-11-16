import { computed, ref } from 'vue'
import { defineStore } from 'pinia'

import { apiClient } from '@/api/apiClient'
import type { Room, RoomCode, RoomId, WebSocketMode } from '@/api/schema'
import { useCurrentUserStore } from '@/stores/currentUser'
import { useRoomWebSocketStore } from '@/stores/roomWebSocket'

export const useRoomsStore = defineStore('rooms', () => {
  const roomWebSocketStore = useRoomWebSocketStore()
  const currentUserStore = useCurrentUserStore()

  const rooms = ref<Room[]>([])
  const initialized = ref(false)
  const activeRoomId = ref<RoomId | null>(null)
  const joinError = ref('')

  const isActiveRoom = computed(
    () => activeRoomId.value !== null && roomWebSocketStore.roomId === activeRoomId.value,
  )

  const roomsByCode = computed(() => {
    const byCode = new Map<RoomCode, Room>()
    for (const room of rooms.value) {
      const previous = byCode.get(room.roomCode)
      if (
        !previous ||
        (previous.state === 'finished' && room.state !== 'finished') ||
        ((previous.state === 'finished') === (room.state === 'finished') &&
          Date.parse(room.createdAt) > Date.parse(previous.createdAt))
      ) {
        byCode.set(room.roomCode, room)
      }
    }
    return byCode
  })

  async function fetchRooms(): Promise<void> {
    const { data, error } = await apiClient.GET('/api/rooms')

    if (error !== undefined || data === undefined) {
      throw new Error('Failed to fetch rooms')
    }

    rooms.value = data
    initialized.value = true
  }

  async function init(force = false): Promise<void> {
    if (initialized.value && !force) {
      return
    }

    await fetchRooms()
  }

  async function getRoomByCode(roomCode: RoomCode): Promise<Room | null> {
    await init()
    if (!roomsByCode.value.has(roomCode)) await init(true)
    return roomsByCode.value.get(roomCode) ?? null
  }

  async function getRoomIdByCode(roomCode: RoomCode): Promise<RoomId | null> {
    const room = await getRoomByCode(roomCode)
    return room?.roomId ?? null
  }

  async function joinRoom(roomId: RoomId): Promise<void> {
    const { error } = await apiClient.POST('/api/rooms/{roomId}/participants', {
      params: {
        path: {
          roomId,
        },
      },
    })

    if (error !== undefined) {
      throw new Error('Failed to join room')
    }

    // The join is committed by the POST. A follow-up room-list refresh is
    // best-effort; its failure must not make the caller treat the join as failed.
    try {
      await fetchRooms()
    } catch {
      // Keep the existing room cache and let the caller continue with the join.
    }
  }

  async function connectToRoom(
    roomCode: RoomCode | undefined,
    { mode, join = false }: { mode: WebSocketMode; join?: boolean },
  ): Promise<void> {
    if (roomWebSocketStore.mode === mode) roomWebSocketStore.disconnect()
    activeRoomId.value = null
    joinError.value = ''

    if (!roomCode) return

    try {
      const userId = join ? await currentUserStore.getUserId() : undefined
      const room = await getRoomByCode(roomCode)
      const roomId = room?.roomId ?? null
      activeRoomId.value = roomId

      if (!roomId) return

      if (join) {
        const isParticipant =
          room?.participants.some((participant) => participant.user.userId === userId) ?? false

        if (!isParticipant) {
          if (room?.state !== 'waiting') {
            joinError.value = 'このルームには参加できません。'
            return
          }

          await joinRoom(roomId)
        }

        roomWebSocketStore.connect({ roomId, mode, userId })
        return
      }

      if (!roomWebSocketStore.isActiveConnection({ roomId, mode })) {
        roomWebSocketStore.connect({ roomId, mode })
      }
    } catch (error) {
      joinError.value = error instanceof Error ? error.message : 'ルームに参加できませんでした。'
    }
  }

  return {
    rooms,
    roomsByCode,
    activeRoomId,
    isActiveRoom,
    joinError,
    fetchRooms,
    init,
    getRoomByCode,
    getRoomIdByCode,
    joinRoom,
    connectToRoom,
  }
})
