import { computed, onScopeDispose, ref, shallowRef } from 'vue'
import { defineStore } from 'pinia'

import type {
  User,
  BingoSummary,
  BingoUpdate,
  Card,
  CardChanges,
  DisplayGameFinishedBody,
  DisplayGameStartedBody,
  DisplayInitializedBody,
  DisplayPickFinishedBody,
  GameSettings,
  Message,
  ParticipantGameFinishedBody,
  ParticipantGameStartedBody,
  ParticipantInitializedBody,
  ParticipantPickFinishedBody,
  PickState,
  PickedBall,
  ReachSummary,
  ReachUpdate,
  RoomId,
  RoomState,
  DisplayWebSocketEvent,
  ParticipantWebSocketEvent,
  WebSocketMode,
} from '@/api/schema'

export type RoomWebSocketEvent = ParticipantWebSocketEvent | DisplayWebSocketEvent

/** WebSocket 接続自体の状態。 */
export type RoomWebSocketConnectionStatus = 'idle' | 'connecting' | 'open' | 'closed' | 'error'

type InitializedBody = ParticipantInitializedBody | DisplayInitializedBody
type GameStartedBody = ParticipantGameStartedBody | DisplayGameStartedBody
type PickFinishedBody = ParticipantPickFinishedBody | DisplayPickFinishedBody
type GameFinishedBody = ParticipantGameFinishedBody | DisplayGameFinishedBody

type ConnectOptions = {
  roomId: RoomId
  mode: WebSocketMode
}

function createRoomWebSocketUrl(roomId: RoomId, mode: WebSocketMode) {
  const url = new URL(`/api/rooms/${roomId}/ws`, window.location.origin)
  url.protocol = url.protocol === 'https:' ? 'wss:' : 'ws:'
  url.searchParams.set('mode', mode)
  return url.toString()
}

export const useRoomWebSocketStore = defineStore('roomWebSocket', () => {
  const socket = shallowRef<WebSocket | null>(null)
  const roomId = ref<RoomId | null>(null)
  const mode = ref<WebSocketMode | null>(null)
  const status = ref<RoomWebSocketConnectionStatus>('idle')
  const reconnectAttempt = ref(0)
  let reconnectTimer: ReturnType<typeof setTimeout> | undefined
  let generation = 0
  let sawOffline = false
  let enabled = false
  let lastMessageAt = 0
  const staleAfter = 120_000
  const errorMessage = ref<string | null>(null)

  const latestEvent = ref<RoomWebSocketEvent | null>(null)
  const latestBall = ref<PickedBall | null>(null)
  const latestCardChanges = ref<CardChanges | null>(null)
  const latestNewBingos = ref<BingoUpdate[]>([])
  const latestNewReaches = ref<ReachUpdate[]>([])
  const latestMessage = ref<Message | null>(null)
  const participants = ref<User[]>([])

  const roomState = ref<RoomState | null>(null)
  const settings = ref<GameSettings | null>(null)
  const pickState = ref<PickState | null>(null)
  const pickedBalls = ref<PickedBall[]>([])
  const bingoSummaries = ref<BingoSummary[]>([])
  const reachSummaries = ref<ReachSummary[]>([])
  const card = ref<Card | null>(null)
  const participantCount = ref<number | null>(null)
  const qrCodeVisible = ref<boolean | null>(null)

  const isConnected = computed(() => status.value === 'open')
  const canUseParticipantCard = computed(() => mode.value === 'participant' && card.value !== null)

  function isActiveConnection(options: ConnectOptions) {
    return (
      roomId.value === options.roomId &&
      mode.value === options.mode &&
      (status.value === 'connecting' || status.value === 'open')
    )
  }

  function resetRoomState() {
    participants.value = []
    latestEvent.value = null
    latestBall.value = null
    latestCardChanges.value = null
    latestNewBingos.value = []
    latestNewReaches.value = []
    latestMessage.value = null
    roomState.value = null
    settings.value = null
    pickState.value = null
    pickedBalls.value = []
    bingoSummaries.value = []
    reachSummaries.value = []
    card.value = null
    participantCount.value = null
    qrCodeVisible.value = null
  }

  function applyEvent(event: RoomWebSocketEvent) {
    latestEvent.value = event

    switch (event.type) {
      case 'ParticipantJoined':
        participantCount.value = event.body.participantCount
        participants.value.push(event.body.participant)
        break
      case 'Initialized': {
        const body = event.body as InitializedBody

        roomState.value = body.state
        settings.value = body.settings
        pickState.value = body.pickState
        latestBall.value = body.pickedBalls.at(-1) ?? null
        latestCardChanges.value = null
        latestNewBingos.value = []
        latestNewReaches.value = []
        pickedBalls.value = body.pickedBalls
        bingoSummaries.value = body.bingoSummaries
        reachSummaries.value = body.reachSummaries
        if (mode.value === 'participant') {
          card.value = (body as ParticipantInitializedBody).card ?? null
        } else {
          const displayBody = body as DisplayInitializedBody
          participants.value = displayBody.participants
          participantCount.value = displayBody.participantCount
          qrCodeVisible.value = displayBody.qrCodeVisible
        }
        break
      }
      case 'GameStarted': {
        const body = event.body as GameStartedBody

        roomState.value = 'playing'
        pickState.value = 'idle'
        if (mode.value === 'participant') {
          card.value = (body as ParticipantGameStartedBody).card
        } else {
          participantCount.value = (body as DisplayGameStartedBody).participantCount
        }
        break
      }
      case 'PickStarted':
        pickState.value = 'picking'
        break
      case 'PickCanceled':
        pickState.value = 'idle'
        break
      case 'PickFinished': {
        const body = event.body as PickFinishedBody

        latestBall.value = body.pickedBall
        latestNewBingos.value = body.newBingos
        latestNewReaches.value = body.newReaches
        pickState.value = body.pickState
        pickedBalls.value = body.pickedBalls
        bingoSummaries.value = body.bingoSummaries
        reachSummaries.value = body.reachSummaries
        if (mode.value === 'participant') {
          const participantBody = body as ParticipantPickFinishedBody
          card.value = participantBody.card
          latestCardChanges.value = participantBody.cardChanges
        } else {
          participantCount.value = (body as DisplayPickFinishedBody).participantCount
          latestCardChanges.value = null
        }
        break
      }
      case 'GameFinished': {
        const body = event.body as GameFinishedBody

        roomState.value = body.state
        pickState.value = body.pickState
        bingoSummaries.value = body.bingoSummaries
        reachSummaries.value = body.reachSummaries
        if (mode.value === 'participant') {
          card.value = (body as ParticipantGameFinishedBody).card
        } else {
          participantCount.value = (body as DisplayGameFinishedBody).participantCount
        }
        break
      }
      case 'ShowQRCode':
        qrCodeVisible.value = true
        break
      case 'HideQRCode':
        qrCodeVisible.value = false
        break
      case 'MessageCreated':
        latestMessage.value = event.body.message
        break
      case 'AllPicked':
        pickedBalls.value = event.body.pickedBalls
        pickState.value = 'exhausted'
        break
      case 'GameSettingsUpdated':
        settings.value = event.body.settings
        break
    }
  }

  function connect(options: ConnectOptions) {
    if (isActiveConnection(options)) {
      return
    }

    const changed = roomId.value !== options.roomId || mode.value !== options.mode
    disconnect()
    if (changed) {
      resetRoomState()
      reconnectAttempt.value = 0
    }
    enabled = true
    const connectionGeneration = generation
    installListeners()

    const wsUrl = createRoomWebSocketUrl(options.roomId, options.mode)

    roomId.value = options.roomId
    mode.value = options.mode
    status.value = 'connecting'
    errorMessage.value = null

    let nextSocket: WebSocket
    try {
      nextSocket = new WebSocket(wsUrl)
    } catch {
      status.value = 'error'
      scheduleReconnect()
      return
    }
    socket.value = nextSocket

    nextSocket.addEventListener('open', () => {
      if (generation !== connectionGeneration || socket.value !== nextSocket) {
        return
      }
      status.value = 'open'
      reconnectAttempt.value = 0
      lastMessageAt = Date.now()
    })

    nextSocket.addEventListener('message', (messageEvent: MessageEvent<string>) => {
      if (generation !== connectionGeneration || socket.value !== nextSocket) {
        return
      }

      lastMessageAt = Date.now()
      try {
        const event = JSON.parse(messageEvent.data) as RoomWebSocketEvent
        applyEvent(event)
      } catch (error) {
        errorMessage.value =
          error instanceof Error ? error.message : 'Failed to parse WebSocket event'
      }
    })

    nextSocket.addEventListener('error', () => {
      if (generation !== connectionGeneration || socket.value !== nextSocket) {
        return
      }
      status.value = 'error'
      errorMessage.value = 'WebSocket connection failed'
      nextSocket.close()
      scheduleReconnect()
    })

    nextSocket.addEventListener('close', () => {
      if (generation !== connectionGeneration || socket.value !== nextSocket) {
        return
      }
      status.value = 'closed'
      socket.value = null
      scheduleReconnect()
    })
  }

  function scheduleReconnect() {
    if (!enabled || reconnectTimer) return
    const delay = Math.min(10_000, 500 * 2 ** Math.min(reconnectAttempt.value, 5))
    reconnectAttempt.value += 1
    const scheduledGeneration = generation
    reconnectTimer = setTimeout(
      () => {
        reconnectTimer = undefined
        if (generation === scheduledGeneration) reconnectNow()
      },
      delay * (0.75 + Math.random() * 0.25),
    )
  }

  function reconnectNow() {
    if (!enabled || !roomId.value || !mode.value) return
    clearTimeout(reconnectTimer)
    reconnectTimer = undefined
    const attempt = reconnectAttempt.value
    const options = { roomId: roomId.value, mode: mode.value }
    // Replace even an apparently OPEN socket when its application state may be stale.
    disconnect()
    reconnectAttempt.value = attempt
    connect(options)
  }

  function resumeConnection() {
    if (document.visibilityState === 'hidden' || !enabled || !navigator.onLine) return
    if (
      sawOffline ||
      socket.value?.readyState !== WebSocket.OPEN ||
      Date.now() - lastMessageAt > staleAfter
    ) {
      sawOffline = false
      reconnectNow()
    }
  }

  function installListeners() {
    window.addEventListener('offline', markOffline)
    window.addEventListener('online', resumeConnection)
    document.addEventListener('visibilitychange', resumeConnection)
  }

  function markOffline() {
    sawOffline = true
  }

  function disconnect() {
    generation += 1
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

  onScopeDispose(disconnect)

  return {
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
    participants,
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
    isActiveConnection,
    connect,
    disconnect,
    resetRoomState,
  }
})
