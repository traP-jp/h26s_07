import type { CreateRoomRequest, GameSettings, Room } from '@/api/schema'

import { http } from '../../http'
import { readJson } from './core'
import { fallbackRoom, mockRooms } from './fixtures/index'

function roomFromSettings(settings: GameSettings): Room {
  const createdAt = new Date().toISOString()
  return {
    ...fallbackRoom,
    roomId: crypto.randomUUID(),
    settings,
    participants: [],
    bingoSummaries: [],
    reachSummaries: [],
    createdAt,
    updatedAt: createdAt,
  }
}

export const roomHandlers = [
  http.get('/api/rooms', ({ response }) => {
    return response(200).json(mockRooms)
  }),

  http.post('/api/rooms', async ({ request, response }) => {
    const body = await readJson<CreateRoomRequest>(request)
    const settings: GameSettings = {
      name: body?.settings.name ?? 'デモビンゴ',
      description: body?.settings.description ?? 'モック API で作成したビンゴルームです。',
      admins: (body?.settings.adminUserIds ?? ['mumumu']).map((userId) => ({ userId })),
    }

    return response(200).json(roomFromSettings(settings))
  }),

  http.get('/api/rooms/{roomId}', ({ params, response }) => {
    return response(200).json(
      mockRooms.find((room) => room.roomId === params.roomId) ?? fallbackRoom,
    )
  }),
]
