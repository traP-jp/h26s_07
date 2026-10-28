import type { RoomId, WebSocketMode } from '@/api/schema'

export function createRoomWebSocketUrl(roomId: RoomId, mode: WebSocketMode): string {
  const url = new URL(`/api/rooms/${roomId}/ws`, window.location.origin)
  url.protocol = url.protocol === 'https:' ? 'wss:' : 'ws:'
  url.searchParams.set('mode', mode)
  return url.toString()
}
