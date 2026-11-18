import type { Message } from '@/api/schema'

import { chatScrollHistory, mockMessages } from './fixtures/chats'
import { chatScrollRoom } from './fixtures/rooms'

const messagesByRoom = new Map<string, Message[]>()

export function messagesForRoom(roomId: string): Message[] {
  let messages = messagesByRoom.get(roomId)
  if (!messages) {
    const history = roomId === chatScrollRoom.roomId ? chatScrollHistory : mockMessages
    messages = history.map((message) => ({ ...message, author: { ...message.author } }))
    messagesByRoom.set(roomId, messages)
  }

  return messages
}

export function appendMessageToRoom(roomId: string, message: Message): void {
  messagesForRoom(roomId).push(message)
}
