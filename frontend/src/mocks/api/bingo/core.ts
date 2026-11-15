import type { User, UserId, WebSocketEventType, WebSocketMode } from '@/api/schema'

type JsonRequest = {
  json(): Promise<unknown>
}

export type MockSocketConnection = {
  roomId: string
  mode: WebSocketMode
  userId: UserId
  send(data: string): void
  close(code?: number, reason?: string): void
}

const socketConnections = new Set<MockSocketConnection>()

export function setSocketConnection(connection: MockSocketConnection): void {
  socketConnections.add(connection)
}

export function clearSocketConnection(connection: MockSocketConnection): void {
  socketConnections.delete(connection)
}

export function sendEvent<TBody>(
  event: { type: WebSocketEventType; body: TBody },
  target: { roomId: string; mode?: WebSocketMode },
): void {
  const serializedEvent = JSON.stringify(event)
  for (const connection of socketConnections) {
    if (connection.roomId !== target.roomId) continue
    if (target.mode && connection.mode !== target.mode) continue
    connection.send(serializedEvent)
  }
}

export function sendEventToConnection<TBody>(
  connection: MockSocketConnection,
  event: { type: WebSocketEventType; body: TBody },
): void {
  if (!socketConnections.has(connection)) return
  connection.send(JSON.stringify(event))
}

export function currentUser(request: Request): User {
  const forwardedUser = request.headers.get('X-Forwarded-User')?.trim()
  return {
    userId: forwardedUser || 'mumumu',
  }
}

export async function readJson<T>(request: JsonRequest): Promise<T | undefined> {
  try {
    return (await request.json()) as T
  } catch {
    return undefined
  }
}

export function pathParam(value: unknown): string | undefined {
  if (Array.isArray(value)) {
    return typeof value[0] === 'string' ? value[0] : undefined
  }

  return typeof value === 'string' ? value : undefined
}
