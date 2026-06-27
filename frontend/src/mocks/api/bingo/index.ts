import { chatHandlers } from './chats'
import { controlHandlers } from './controls'
import { meHandlers } from './me'
import { participantHandlers } from './participants'
import { roomHandlers } from './rooms'
import { settingHandlers } from './settings'
import { roomWebSocketHandler } from './websocket'

export const bingoHandlers = [
  roomWebSocketHandler,
  ...meHandlers,
  ...roomHandlers,
  ...participantHandlers,
  ...chatHandlers,
  ...controlHandlers,
  ...settingHandlers,
]
