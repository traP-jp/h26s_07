import type { CreateMessageRequest, Message } from '@/api/schema'

import { http } from '../../http'
import { currentUser, readJson, sendEvent } from './core'
import { appendMessageToRoom, messagesForRoom } from './chatHistory'

export const chatHandlers = [
  http.get('/api/rooms/{roomId}/chats', ({ params, response }) => {
    return response(200).json(messagesForRoom(params.roomId))
  }),

  http.post('/api/rooms/{roomId}/chats', async ({ params, request, response }) => {
    const body = await readJson<CreateMessageRequest>(request)
    const message: Message = {
      messageId: crypto.randomUUID(),
      content: body?.content ?? 'モックメッセージ',
      author: currentUser(request),
      createdAt: new Date().toISOString(),
    }

    appendMessageToRoom(params.roomId, message)

    sendEvent({ type: 'MessageCreated', body: { message } }, { roomId: params.roomId })

    return response(200).json(message)
  }),
]
