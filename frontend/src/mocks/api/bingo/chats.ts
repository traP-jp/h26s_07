import type { CreateMessageRequest, Message } from '@/api/schema'

import { http } from '../../http'
import { currentUser, readJson, sendEvent } from './core'
import { mockMessages } from './fixtures/index'

export const chatHandlers = [
  http.get('/api/rooms/{roomId}/chats', ({ response }) => {
    return response(200).json(mockMessages)
  }),

  http.post('/api/rooms/{roomId}/chats', async ({ request, response }) => {
    const body = await readJson<CreateMessageRequest>(request)
    const message: Message = {
      messageId: crypto.randomUUID(),
      content: body?.content ?? 'モックメッセージ',
      author: currentUser(request),
      createdAt: new Date().toISOString(),
    }

    mockMessages.push(message)

    sendEvent({ type: 'MessageCreated', body: { message } })

    return response(200).json(message)
  }),
]
