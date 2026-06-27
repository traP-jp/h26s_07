import { computed, ref, shallowRef } from 'vue'
import { defineStore } from 'pinia'

import type {
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
  const errorMessage = ref<string | null>(null)

  const latestEvent = ref<RoomWebSocketEvent | null>(null)
  const latestPickedBall = ref<PickedBall | null>(null)
  const latestCardChanges = ref<CardChanges | null>(null)
  const latestNewBingos = ref<BingoUpdate[]>([])
  const latestNewReaches = ref<ReachUpdate[]>([])
  const latestMessage = ref<Message | null>(null)

  const roomState = ref<RoomState | null>(null)
  const settings = ref<GameSettings | null>(null)
  const pickState = ref<PickState | null>(null)
  const pickedBalls = ref<PickedBall[]>([])
  const bingoSummaries = ref<BingoSummary[]>([])
  const reachSummaries = ref<ReachSummary[]>([])
  const card = ref<Card | null>(null)
  const participantCount = ref<number | null>(null)
  const qrCodeVisible = ref<boolean | null>(null)
  const messages = ref<Message[]>([])

  const isConnected = computed(() => status.value === 'open')
  const canUseParticipantCard = computed(() => mode.value === 'participant' && card.value !== null)

  function resetRoomState() {
    latestEvent.value = null
    latestPickedBall.value = null
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
    messages.value = []
  }

  function applyEvent(event: RoomWebSocketEvent) {
    latestEvent.value = event

    switch (event.type) {
      case 'Initialized': {
        const body = event.body as InitializedBody

        roomState.value = body.state
        settings.value = body.settings
        pickState.value = body.pickState
        pickedBalls.value = body.pickedBalls
        bingoSummaries.value = body.bingoSummaries
        reachSummaries.value = body.reachSummaries
        if (mode.value === 'participant') {
          card.value = (body as ParticipantInitializedBody).card ?? null
        } else {
          const displayBody = body as DisplayInitializedBody
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

        latestPickedBall.value = body.pickedBall
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
        messages.value = [...messages.value, event.body.message]
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
    disconnect()
    resetRoomState()

    const wsUrl = createRoomWebSocketUrl(options.roomId, options.mode)

    roomId.value = options.roomId
    mode.value = options.mode
    status.value = 'connecting'
    errorMessage.value = null

    const nextSocket = new WebSocket(wsUrl)
    socket.value = nextSocket

    nextSocket.addEventListener('open', () => {
      if (socket.value !== nextSocket) {
        return
      }
      status.value = 'open'
    })

    nextSocket.addEventListener('message', (messageEvent: MessageEvent<string>) => {
      if (socket.value !== nextSocket) {
        return
      }

      try {
        const event = JSON.parse(messageEvent.data) as RoomWebSocketEvent
        applyEvent(event)
      } catch (error) {
        errorMessage.value =
          error instanceof Error ? error.message : 'Failed to parse WebSocket event'
      }
    })

    nextSocket.addEventListener('error', () => {
      if (socket.value !== nextSocket) {
        return
      }
      status.value = 'error'
      errorMessage.value = 'WebSocket connection failed'
    })

    nextSocket.addEventListener('close', () => {
      if (socket.value !== nextSocket) {
        return
      }
      status.value = 'closed'
      socket.value = null
    })
  }

  function disconnect() {
    const currentSocket = socket.value
    if (currentSocket) {
      socket.value = null
      currentSocket.close()
    }
    status.value = 'closed'
  }

  return {
    socket,
    roomId,
    mode,
    status,
    errorMessage,
    latestEvent,
    latestPickedBall,
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
    messages,
    isConnected,
    canUseParticipantCard,
    connect,
    disconnect,
    resetRoomState,
  }
})
