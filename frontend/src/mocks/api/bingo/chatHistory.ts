import type { Message } from '@/api/schema'

import { mockMessages } from './fixtures/chats'

const messagesByRoom = new Map<string, Message[]>()

export function messagesForRoom(roomId: string): Message[] {
  let messages = messagesByRoom.get(roomId)
  if (!messages) {
    messages = mockMessages.map((message) => ({ ...message, author: { ...message.author } }))
    messagesByRoom.set(roomId, messages)
  }

  return messages
}

export function appendMessageToRoom(roomId: string, message: Message): void {
  messagesForRoom(roomId).push(message)
}
