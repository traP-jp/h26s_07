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
    const { data, error } = await apiClient.GET('/api/rooms')

    if (error !== undefined || data === undefined) {
      throw new Error('Failed to fetch rooms')
    }

    rooms.value = data
    initialized.value = true
  }

  async function init(): Promise<void> {
    if (initialized.value) {
      return
    }

    await fetchRooms()
  }

  async function getRoomByCode(roomCode: RoomCode): Promise<Room | null> {
    await init()
    return roomsByCode.value.get(roomCode) ?? null
  }

  async function getRoomIdByCode(roomCode: RoomCode): Promise<RoomId | null> {
    const room = await getRoomByCode(roomCode)
    return room?.roomId ?? null
  }

  return {
    rooms,
    roomsByCode,
    fetchRooms,
    init,
    getRoomByCode,
    getRoomIdByCode,
  }
})
