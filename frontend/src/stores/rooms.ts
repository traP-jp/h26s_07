import { computed, ref } from 'vue'
import { defineStore } from 'pinia'

import { apiClient } from '@/api/apiClient'
import type { Room, RoomCode, RoomId } from '@/api/schema'

export const useRoomsStore = defineStore('rooms', () => {
  const rooms = ref<Room[]>([])

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
  }

  async function init(): Promise<void> {
    await fetchRooms()
  }

  async function getRoomByCode(roomCode: RoomCode): Promise<Room | null> {
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

    // 参加処理は POST で完了しているため、ルーム一覧の更新失敗で参加自体を失敗扱いしない。
    try {
      await fetchRooms()
    } catch {
      // 既存のルーム一覧を維持して、参加処理を続行する。
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
