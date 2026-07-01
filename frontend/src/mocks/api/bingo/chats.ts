import type { CreateMessageRequest, Message, MessageCreatedBody } from '@/api/schema'

import { http } from '../../http'
import { broadcastRoomEvent, currentUser, pathParam, readJson } from './core'

const mockMessages: Message[] = [
  {
    messageId: '00000000-0000-4000-8000-000000000101',
    content: 'モックルームへようこそ',
    author: { userId: 'mumumu' },
    createdAt: '2026-06-27T10:02:00.000Z',
  },
  {
    messageId: '00000000-0000-4000-8000-000000000102',
    content: '準備できています',
    author: { userId: 'rurun' },
    createdAt: '2026-06-27T10:03:00.000Z',
  },
]

export const chatHandlers = [
  http.get('/api/rooms/{roomId}/chats', ({ response }) => {
    return response(200).json(mockMessages)
  }),

  http.post('/api/rooms/{roomId}/chats', async ({ request, response, params }) => {
    const body = await readJson<CreateMessageRequest>(request)
    const message: Message = {
      messageId: crypto.randomUUID(),
      content: body?.content ?? 'モックメッセージ',
      author: currentUser(request),
      createdAt: new Date().toISOString(),
    }

    mockMessages.push(message)

    const roomId = pathParam(params.roomId)
    if (roomId) {
      broadcastRoomEvent<MessageCreatedBody>(roomId, 'MessageCreated', { message })
    }

    return response(200).json(message)
  }),
]
