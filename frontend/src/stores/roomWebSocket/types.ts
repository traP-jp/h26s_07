import type {
  DisplayWebSocketEvent,
  ParticipantWebSocketEvent,
  RoomId,
  WebSocketMode,
} from '@/api/schema'

export type RoomWebSocketEvent = ParticipantWebSocketEvent | DisplayWebSocketEvent

/** WebSocket 接続自体の状態。 */
export type RoomWebSocketConnectionStatus = 'idle' | 'connecting' | 'open' | 'closed' | 'error'

export type ConnectOptions = {
  roomId: RoomId
  mode: WebSocketMode
}
