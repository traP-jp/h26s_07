<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import BingoCardPaper from '@/components/bingo/BingoCardPaper.vue'
import ChatContainer from '@/components/chat/ChatContainer.vue'
import { useRoomWebSocketStore } from '@/stores/roomWebSocket'
import { useRoute } from 'vue-router'
import type { RoomCode, Card, RoomId, UserId } from '@/api/schema'
import { useRoomsStore } from '@/stores/rooms'
import { useCurrentUserStore } from '@/stores/currentUser'
import { storeToRefs } from 'pinia'
import { Fireworks } from 'fireworks-js'

const route = useRoute()
const roomCode = route.params.roomCode as RoomCode | undefined
const roomId = ref<RoomId | null>(null)
const roomWebSocketStore = useRoomWebSocketStore()
const roomsStore = useRoomsStore()
const currentUserStore = useCurrentUserStore()
const { roomsByCode } = storeToRefs(roomsStore)
const { mode, roomId: connectedRoomId, roomState } = storeToRefs(roomWebSocketStore)

const { card, latestCardChanges, latestEvent } = storeToRefs(roomWebSocketStore)

const displayCard = ref<Card | null>(null)
const fireworksOverlay = ref<HTMLElement | null>(null)
const isGameWaiting = computed(() => roomState.value === 'waiting')
let fireworks: Fireworks | undefined
let fireworksStopTimer: ReturnType<typeof setTimeout> | undefined
const errorMessage = ref('')
const currentUserId = ref<UserId>()

onMounted(async () => {
  if (!roomCode) return

  currentUserId.value = await currentUserStore.getUserId()

  await roomsStore.init()
  const room = roomsByCode.value.get(roomCode)
  roomId.value = room?.roomId ?? null

  if (!roomId.value) return

  if (connectedRoomId.value === roomId.value && mode.value === 'participant') {
    return
  }

  const isParticipant =
    room?.participants.some((participant) => participant.user.userId === currentUserId.value) ??
    false

  if (!isParticipant) {
    if (room?.state !== 'waiting') {
      errorMessage.value = 'このルームには参加できません。'
      return
    }

    await roomsStore.joinRoom(roomId.value)
  }

  roomWebSocketStore.connect({ roomId: roomId.value, mode: 'participant' })
})
onBeforeUnmount(() => {
  stopBingoFireworks(true)

  if (roomId.value && connectedRoomId.value === roomId.value) {
    roomWebSocketStore.disconnect()
  }
})

watch(latestEvent, async (event) => {
  if (!event) return

  switch (event.type) {
    case 'Initialized':
    case 'GameStarted':
    case 'PickFinished':
    case 'GameFinished':
      displayCard.value = card.value
      break

    default:
      break
  }
})

watch(
  () => latestCardChanges.value?.newBingoLines,
  async (newBingoLines) => {
    if (!newBingoLines?.length) return

    await nextTick()
    playBingoFireworks()
  },
)

function playBingoFireworks() {
  if (window.matchMedia('(prefers-reduced-motion: reduce)').matches) return
  if (!fireworksOverlay.value) return

  stopBingoFireworks(true)

  fireworks = new Fireworks(fireworksOverlay.value, {
    autoresize: true,
    opacity: 0.78,
    acceleration: 1.03,
    friction: 0.965,
    gravity: 1.18,
    particles: 120,
    explosion: 8,
    intensity: 46,
    traceLength: 3.1,
    traceSpeed: 8,
    rocketsPoint: { min: 8, max: 92 },
    hue: { min: 185, max: 345 },
    delay: { min: 8, max: 20 },
    brightness: { min: 84, max: 100 },
    decay: { min: 0.011, max: 0.022 },
    flickering: 42,
    lineWidth: {
      explosion: { min: 1.8, max: 4.4 },
      trace: { min: 1.4, max: 3.4 },
    },
    lineStyle: 'round',
    mouse: {
      click: false,
      move: false,
      max: 0,
    },
    sound: {
      enabled: false,
      files: [],
      volume: { min: 0, max: 0 },
    },
  })

  fireworks.start()
  fireworks.launch(12)

  fireworksStopTimer = setTimeout(() => {
    fireworks?.launch(9)
    fireworksStopTimer = setTimeout(() => {
      fireworks?.launch(6)
      fireworksStopTimer = setTimeout(() => stopBingoFireworks(), 1900)
    }, 520)
  }, 420)
}

function stopBingoFireworks(dispose = false) {
  if (fireworksStopTimer) {
    clearTimeout(fireworksStopTimer)
    fireworksStopTimer = undefined
  }

  fireworks?.stop(dispose)
  fireworks = undefined
}
</script>

<template>
  <div class="participant-room">
    <div class="participant-room__card">
      <div v-if="isGameWaiting" class="participant-room__waiting" role="status">
        <p class="participant-room__waiting-title">ゲームはまだ始まっていません</p>
        <p class="participant-room__waiting-text">開始されるまでお待ちください</p>
      </div>
      <BingoCardPaper
        v-else
        class="participant-room__bingo-card"
        :card="displayCard"
        :card-changes="latestCardChanges"
        :placeholder="displayCard === null"
      />
    </div>
    <ChatContainer
      v-if="roomCode"
      class="participant-room__chat"
      :room-code="roomCode"
      textarea
      variant="participant"
    />

    <div ref="fireworksOverlay" class="participant-room__fireworks" aria-hidden="true"></div>
  </div>
</template>

<style scoped>
.participant-room {
  display: flex;
  height: 100dvh;
  padding: 18px;

  background:
    radial-gradient(circle at 12% 18%, rgb(176 226 255 / 0.75), transparent 34%),
    radial-gradient(circle at 82% 16%, rgb(255 197 223 / 0.7), transparent 32%),
    radial-gradient(circle at 24% 82%, rgb(194 255 231 / 0.65), transparent 36%),
    radial-gradient(circle at 78% 76%, rgb(211 198 255 / 0.7), transparent 34%),
    linear-gradient(135deg, #f7fbff 0%, #fff8fd 48%, #f4fff9 100%);
}

.participant-room__card {
  display: flex;
  flex: 1;
  align-items: center;
  justify-content: center;
  container-type: size;
}

.participant-room__bingo-card {
  width: min(85cqw, 80cqh);
}

.participant-room__chat {
  width: 360px;
  flex-shrink: 0;
  height: 100%;
}

.participant-room__waiting {
  box-sizing: border-box;
  width: min(100%, 420px);
  padding: 28px 24px;
  border: 1px solid rgb(255 255 255 / 0.7);
  border-radius: 14px;
  background: rgb(255 255 255 / 0.82);
  box-shadow: 0 14px 40px rgb(20 55 110 / 0.14);
  color: #1f3556;
  text-align: center;
  backdrop-filter: blur(12px) saturate(1.15);
}

.participant-room__waiting-title {
  margin: 0;
  font-size: 22px;
  font-weight: 900;
  line-height: 1.35;
}

.participant-room__waiting-text {
  margin: 10px 0 0;
  color: #50627d;
  font-size: 15px;
  font-weight: 700;
  line-height: 1.45;
}

.participant-room__fireworks {
  position: fixed;
  inset: 0;
  z-index: 30;
  overflow: hidden;
  mix-blend-mode: screen;
  pointer-events: none;
}

.participant-room__fireworks :deep(canvas) {
  display: block;
  width: 100% !important;
  height: 100% !important;
}

@media (max-width: 639px) {
  .participant-room {
    height: 100dvh;
    align-items: center;
    justify-content: center;
    gap: 0;
    padding: 8px 8px 68px;
  }

  .participant-room__card {
    width: 100%;
    height: 100%;
  }

  .participant-room__waiting {
    width: min(100%, 340px);
    padding: 22px 18px;
  }

  .participant-room__waiting-title {
    font-size: 18px;
  }

  .participant-room__waiting-text {
    font-size: 13px;
  }

  .participant-room__chat {
    width: auto;
    height: auto;
    min-height: 0;
  }
}
</style>
