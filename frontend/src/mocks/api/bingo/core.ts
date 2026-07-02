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

let currentSocketConnection: MockSocketConnection | undefined

export function setSocketConnection(connection: MockSocketConnection): void {
  currentSocketConnection = connection
}

export function clearSocketConnection(connection: MockSocketConnection): void {
  if (currentSocketConnection === connection) {
    currentSocketConnection = undefined
  }
}

export function sendEvent<TBody>(event: { type: WebSocketEventType; body: TBody }): void {
  currentSocketConnection?.send(JSON.stringify(event))
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
