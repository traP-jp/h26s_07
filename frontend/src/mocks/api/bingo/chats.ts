import type { CreateMessageRequest, Message } from '@/api/schema'

import { http } from '../../http'
import { currentUser, readJson, sendEvent } from './core'
import { mockRooms } from './fixtures'
import { appendMessageToRoom, messagesForRoom } from './chatHistory'

const CANONICAL_UUID_PATTERN = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i
const HEX_UUID_PATTERN = /^[0-9a-f]{32}$/i

const normalizeRoomId = (value: string): string | undefined => {
  let candidate = value
  if (candidate.length === 45) {
    if (candidate.slice(0, 9).toLowerCase() !== 'urn:uuid:') return undefined
    candidate = candidate.slice('urn:uuid:'.length)
  } else if (candidate.length === 38) {
    if (candidate[0] !== '{' || candidate[37] !== '}') return undefined
    candidate = candidate.slice(1, 37)
  }

  if (HEX_UUID_PATTERN.test(candidate)) {
    candidate = `${candidate.slice(0, 8)}-${candidate.slice(8, 12)}-${candidate.slice(12, 16)}-${candidate.slice(16, 20)}-${candidate.slice(20)}`
  }

  return CANONICAL_UUID_PATTERN.test(candidate) ? candidate.toLowerCase() : undefined
}

export const chatHandlers = [
  http.get('/api/rooms/{roomId}/chats', ({ params, request, response }) => {
    const roomId = normalizeRoomId(params.roomId)
    if (!roomId) {
      return response(400).json({ message: 'Invalid roomId' })
    }

    const room = mockRooms.find((item) => item.roomId === roomId)
    if (!room) return response(404).json({ message: 'room not found' })

    const userId = currentUser(request).userId
    const canViewChat =
      room.participants.some((participant) => participant.user.userId === userId) ||
      room.settings.admins.some((admin) => admin.userId === userId)
    if (!canViewChat) return response(403).json({ message: 'room message not allowed' })

    return response(200).json(messagesForRoom(roomId))
  }),

  http.post('/api/rooms/{roomId}/chats', async ({ params, request, response }) => {
    const roomId = normalizeRoomId(params.roomId)
    if (!roomId) {
      return response(400).json({ message: 'Invalid roomId' })
    }

    const body = await readJson<CreateMessageRequest>(request)
    if (body === undefined) return response(400).json({ message: 'message invalid' })

    const room = mockRooms.find((item) => item.roomId === roomId)
    if (!room) return response(404).json({ message: 'room not found' })

    const author = currentUser(request)
    const isParticipant = room.participants.some(
      (participant) => participant.user.userId === author.userId,
    )
    const isAdmin = room.settings.admins.some((admin) => admin.userId === author.userId)
    if (!isParticipant && !isAdmin) {
      return response(403).json({ message: 'message not allowed' })
    }
    if (room.state === 'finished') {
      return response(409).json({ message: 'message not postable' })
    }

    if (
      typeof body?.content !== 'string' ||
      body.content.length === 0 ||
      Array.from(body.content).length > 500
    ) {
      return response(400).json({ message: 'message invalid' })
    }

    const message: Message = {
      messageId: crypto.randomUUID(),
      content: body.content,
      author,
      createdAt: new Date().toISOString(),
    }

    room.updatedAt = message.createdAt
    appendMessageToRoom(roomId, message)

    sendEvent({ type: 'MessageCreated', body: { message } }, { roomId })

    return response(200).json(message)
  }),
]
