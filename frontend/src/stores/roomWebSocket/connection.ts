import { onScopeDispose } from 'vue'
import type { Ref, ShallowRef } from 'vue'

import { apiClient } from '@/api/apiClient'
import type { RoomId, WebSocketMode } from '@/api/schema'

import { createRoomWebSocketUrl } from './url'
import type { ConnectOptions, RoomWebSocketConnectionStatus, RoomWebSocketEvent } from './types'

type RoomWebSocketConnectionOptions = {
  socket: ShallowRef<WebSocket | null>
  roomId: Ref<RoomId | null>
  mode: Ref<WebSocketMode | null>
  status: Ref<RoomWebSocketConnectionStatus>
  reconnectAttempt: Ref<number>
  errorMessage: Ref<string | null>
  roomDeleted: Ref<boolean>
  resetRoomState: () => void
  handleEvent: (event: RoomWebSocketEvent) => void
}

export function createRoomWebSocketConnection({
  socket,
  roomId,
  mode,
  status,
  reconnectAttempt,
  errorMessage,
  roomDeleted,
  resetRoomState,
  handleEvent,
}: RoomWebSocketConnectionOptions) {
  let reconnectTimer: ReturnType<typeof setTimeout> | undefined
  let generation = 0
  let recoveringGeneration: number | undefined
  let enabled = false
  let lastMessageAt = 0
  const staleAfter = 120_000
  // A socket can stay OPEN while the browser is offline and miss server events.
  let sawOffline = false

  function isActiveConnection(options: ConnectOptions): boolean {
    return (
      roomId.value === options.roomId &&
      mode.value === options.mode &&
      (status.value === 'connecting' || status.value === 'open')
    )
  }

  function connect(options: ConnectOptions): void {
    if (roomDeleted.value && roomId.value === options.roomId) return
    if (isActiveConnection(options)) return

    const changed = roomId.value !== options.roomId || mode.value !== options.mode
    disconnect()
    if (changed) {
      resetRoomState()
      reconnectAttempt.value = 0
    }
    const connectionGeneration = generation
    enabled = true
    installListeners()

    const wsUrl = createRoomWebSocketUrl(options.roomId, options.mode)

    roomId.value = options.roomId
    mode.value = options.mode
    status.value = 'connecting'
    errorMessage.value = null

    openSocket(wsUrl, connectionGeneration)
  }

  function openSocket(wsUrl: string, connectionGeneration: number): void {
    let nextSocket: WebSocket
    try {
      nextSocket = new WebSocket(wsUrl)
    } catch {
      status.value = 'error'
      scheduleReconnect()
      return
    }
    socket.value = nextSocket

    registerSocketListeners(nextSocket, connectionGeneration)
  }

  function registerSocketListeners(nextSocket: WebSocket, connectionGeneration: number): void {
    nextSocket.addEventListener('open', () => {
      if (generation !== connectionGeneration || socket.value !== nextSocket) return
      status.value = 'open'
      reconnectAttempt.value = 0
      lastMessageAt = Date.now()
    })

    nextSocket.addEventListener('message', (messageEvent: MessageEvent<string>) => {
      if (generation !== connectionGeneration || socket.value !== nextSocket) return

      lastMessageAt = Date.now()
      try {
        const event = JSON.parse(messageEvent.data) as RoomWebSocketEvent
        handleEvent(event)
        if (event.type === 'RoomDeleted') markRoomDeleted()
      } catch (error) {
        errorMessage.value =
          error instanceof Error ? error.message : 'Failed to parse WebSocket event'
      }
    })

    nextSocket.addEventListener('error', () => {
      if (generation !== connectionGeneration || socket.value !== nextSocket) return
      status.value = 'error'
      errorMessage.value = 'WebSocket connection failed'
      nextSocket.close()
      scheduleReconnect()
    })

    nextSocket.addEventListener('close', () => {
      if (generation !== connectionGeneration || socket.value !== nextSocket) return
      status.value = 'closed'
      socket.value = null
      scheduleReconnect()
    })
  }

  function scheduleReconnect(): void {
    if (!enabled || reconnectTimer) return
    const delay = Math.min(10_000, 500 * 2 ** Math.min(reconnectAttempt.value, 5))
    reconnectAttempt.value += 1
    const scheduledGeneration = generation
    reconnectTimer = setTimeout(
      () => {
        if (generation === scheduledGeneration) void reconnectNow()
      },
      delay * (0.75 + Math.random() * 0.25),
    )
  }

  async function reconnectNow(): Promise<void> {
    if (!enabled || !roomId.value || !mode.value) return
    clearTimeout(reconnectTimer)
    reconnectTimer = undefined
    const recoveryGeneration = generation
    // online and visibilitychange can fire together; recover only once per socket generation.
    if (recoveringGeneration === recoveryGeneration) return
    recoveringGeneration = recoveryGeneration
    const attempt = reconnectAttempt.value
    const options = { roomId: roomId.value, mode: mode.value }
    try {
      const { response } = await apiClient.GET('/api/rooms/{roomId}', {
        params: { path: { roomId: options.roomId } },
      })
      if (
        generation !== recoveryGeneration ||
        !enabled ||
        roomId.value !== options.roomId ||
        mode.value !== options.mode
      )
        return
      if (response.status === 404) {
        markRoomDeleted()
        return
      }
    } catch {
      // Retry transient network failures using the normal backoff.
    }
    if (
      generation !== recoveryGeneration ||
      !enabled ||
      roomId.value !== options.roomId ||
      mode.value !== options.mode
    )
      return

    // A WebSocket may reconnect while missing events sent during downtime.
    // Replacing it requests a snapshot even when its readyState still says OPEN.
    disconnect()
    reconnectAttempt.value = attempt
    connect(options)
  }

  function resumeConnection(): void {
    if (document.visibilityState === 'hidden' || !enabled || !navigator.onLine) return
    if (
      sawOffline ||
      socket.value?.readyState !== WebSocket.OPEN ||
      Date.now() - lastMessageAt > staleAfter
    ) {
      sawOffline = false
      void reconnectNow()
    }
  }

  function markOffline(): void {
    sawOffline = true
  }

  function installListeners(): void {
    window.addEventListener('offline', markOffline)
    window.addEventListener('online', resumeConnection)
    document.addEventListener('visibilitychange', resumeConnection)
  }

  function disconnect(): void {
    // The generation invalidates callbacks from the old socket and any pending recovery.
    generation += 1
    recoveringGeneration = undefined
    enabled = false
    clearTimeout(reconnectTimer)
    reconnectTimer = undefined
    window.removeEventListener('offline', markOffline)
    window.removeEventListener('online', resumeConnection)
    document.removeEventListener('visibilitychange', resumeConnection)
    const currentSocket = socket.value
    if (currentSocket) {
      socket.value = null
      currentSocket.close()
    }
    status.value = 'closed'
  }

  function markRoomDeleted(): void {
    roomDeleted.value = true
    disconnect()
  }

  onScopeDispose(disconnect)

  return {
    connect,
    disconnect,
    isActiveConnection,
    markRoomDeleted,
  }
}
