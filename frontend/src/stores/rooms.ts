import { computed, ref } from 'vue'
import { defineStore } from 'pinia'

import { apiClient } from '@/api/apiClient'
import type { Room, RoomCode, RoomId } from '@/api/schema'

export const useRoomsStore = defineStore('rooms', () => {
  const rooms = ref<Room[]>([])
  const initialized = ref(false)

  const roomsByCode = computed(() => {
    return new Map(rooms.value.map((room) => [room.roomCode, room]))
  })

  async function fetchRooms(): Promise<void> {
    const { data, error } = await apiClient.GET('/api/rooms', {
      params: { query: { includeFinished: true } },
    })

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

    await fetchRooms()
  }

  async function deleteRoom(roomId: RoomId): Promise<void> {
    const { error } = await apiClient.DELETE('/api/rooms/{roomId}', {
      params: { path: { roomId } },
    })
    if (error !== undefined) throw new Error('ルームの削除に失敗しました。')
    rooms.value = rooms.value.filter((room) => room.roomId !== roomId)
    initialized.value = false
    await fetchRooms()
  }

  return {
    deleteRoom,
    rooms,
    roomsByCode,
    fetchRooms,
    init,
    getRoomByCode,
    getRoomIdByCode,
    joinRoom,
  }
})
