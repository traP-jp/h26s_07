import { computed, reactive, ref, shallowRef, toRefs } from 'vue'
import { defineStore } from 'pinia'

import type { Message, RoomId, WebSocketMode } from '@/api/schema'

import { createRoomWebSocketConnection } from './roomWebSocket/connection'
import { mergeById } from './roomWebSocket/merge'
import { reduceRoomWebSocketEvent } from './roomWebSocket/reducers'
import { createInitialRoomWebSocketData } from './roomWebSocket/state'
import type { RoomWebSocketData } from './roomWebSocket/state'
import type { RoomWebSocketConnectionStatus, RoomWebSocketEvent } from './roomWebSocket/types'

export type { RoomWebSocketConnectionStatus, RoomWebSocketEvent } from './roomWebSocket/types'

export const useRoomWebSocketStore = defineStore('roomWebSocket', () => {
  const socket = shallowRef<WebSocket | null>(null)
  const roomId = ref<RoomId | null>(null)
  const mode = ref<WebSocketMode | null>(null)
  const status = ref<RoomWebSocketConnectionStatus>('idle')
  const reconnectAttempt = ref(0)
  const errorMessage = ref<string | null>(null)
  const data = reactive<RoomWebSocketData>({
    ...createInitialRoomWebSocketData(),
    initializationCount: 0,
  })
  const {
    initializationCount,
    activityLog,
    participants,
    roomDeleted,
    chatMessages,
    latestEvent,
    latestBall,
    latestCardChanges,
    latestNewBingos,
    latestNewReaches,
    latestMessage,
    roomState,
    settings,
    pickState,
    pickedBalls,
    bingoSummaries,
    reachSummaries,
    card,
    participantCount,
    qrCodeVisible,
  } = toRefs(data)

  function resetRoomState(): void {
    Object.assign(data, createInitialRoomWebSocketData())
  }

  function applyEvent(event: RoomWebSocketEvent): void {
    // Keep each write ordered; event handlers historically update shared fields more than once.
    for (const [key, value] of reduceRoomWebSocketEvent(data, event, mode.value)) {
      Reflect.set(data, key, value)
    }
  }

  const connection = createRoomWebSocketConnection({
    socket,
    roomId,
    mode,
    status,
    reconnectAttempt,
    errorMessage,
    roomDeleted,
    resetRoomState,
    handleEvent: applyEvent,
  })

  function mergeChatMessages(messages: Message[]): void {
    data.chatMessages = mergeById(data.chatMessages, messages, (message) => message.messageId)
  }

  const isConnected = computed(() => status.value === 'open')
  const canUseParticipantCard = computed(() => mode.value === 'participant' && card.value !== null)

  return {
    initializationCount,
    activityLog,
    participants,
    roomDeleted,
    chatMessages,
    mergeChatMessages,
    markRoomDeleted: connection.markRoomDeleted,
    reconnectAttempt,
    socket,
    roomId,
    mode,
    status,
    errorMessage,
    latestEvent,
    latestBall,
    latestCardChanges,
    latestNewBingos,
    latestNewReaches,
    latestMessage,
    roomState,
    settings,
    pickState,
    pickedBalls,
    bingoSummaries,
    reachSummaries,
    card,
    participantCount,
    qrCodeVisible,
    isConnected,
    canUseParticipantCard,
    isActiveConnection: connection.isActiveConnection,
    connect: connection.connect,
    disconnect: connection.disconnect,
    resetRoomState,
  }
})
