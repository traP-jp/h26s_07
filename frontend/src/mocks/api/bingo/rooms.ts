import { HttpResponse } from 'msw'
import type { CreateRoomRequest, GameSettings, Room } from '@/api/schema'

import { http } from '../../http'
import { currentUser, readJson, sendEvent } from './core'
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
  http.get('/api/rooms', ({ request, response }) => {
    const includeFinished = new URL(request.url).searchParams.get('includeFinished') === 'true'
    return response(200).json(
      mockRooms.filter((room) => includeFinished || room.state !== 'finished'),
    )
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

  http.delete('/api/rooms/{roomId}', ({ params, request, response }) => {
    const index = mockRooms.findIndex((room) => room.roomId === params.roomId)
    if (index < 0) return response(404).json({ message: 'Room not found' })
    if (
      !mockRooms[index]!.settings.admins.some(
        (admin) => admin.userId === currentUser(request).userId,
      )
    ) {
      return HttpResponse.json({ message: 'Forbidden' }, { status: 403 })
    }
    mockRooms.splice(index, 1)
    sendEvent({ type: 'RoomDeleted', body: {} }, params.roomId)
    return response(204).empty()
  }),

  http.get('/api/rooms/{roomId}', ({ params, response }) => {
    const room = mockRooms.find((room) => room.roomId === params.roomId)
    if (!room) return response(404).json({ message: 'Room not found' })
    return response(200).json(room)
  }),
]
