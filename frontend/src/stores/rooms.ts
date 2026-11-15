import { computed, ref } from 'vue'
import { defineStore } from 'pinia'

import { apiClient } from '@/api/apiClient'
import type { Room, RoomCode, RoomId } from '@/api/schema'

export const useRoomsStore = defineStore('rooms', () => {
  const rooms = ref<Room[]>([])
  const initialized = ref(false)

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

  async function joinRoom(roomId: RoomId, signal?: AbortSignal): Promise<void> {
    const { error } = await apiClient.POST('/api/rooms/{roomId}/participants', {
      params: {
        path: {
          roomId,
        },
      },
      signal,
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

  return {
    rooms,
    roomsByCode,
    fetchRooms,
    init,
    getRoomByCode,
    getRoomIdByCode,
    joinRoom,
  }
})
