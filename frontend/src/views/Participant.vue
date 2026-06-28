<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import BingoCardPaper from '@/components/layouts/BingoCardPaper.vue'
import ChatContainer from '@/components/layouts/ChatContainer.vue'
import { useRoomWebSocketStore } from '@/stores/roomWebSocket'
import { useRoute } from 'vue-router'
import type { RoomCode, Card, RoomId } from '@/api/schema'
import { useRoomsStore } from '@/stores/rooms'
import { storeToRefs } from 'pinia'

const route = useRoute()
const roomCode = route.params.roomCode as RoomCode | undefined
const roomId = ref<RoomId | null>(null)
const roomWebSocketStore = useRoomWebSocketStore()
const roomsStore = useRoomsStore()
const { roomsByCode } = storeToRefs(roomsStore)
const { mode, roomId: connectedRoomId, roomState } = storeToRefs(roomWebSocketStore)

const { card, latestEvent } = storeToRefs(roomWebSocketStore)

const displayCard = ref<Card | null>(null)
const isGameWaiting = computed(() => roomState.value === 'waiting')

onMounted(async () => {
  if (!roomCode) return

  await roomsStore.init()
  roomId.value = roomsByCode.value.get(roomCode)?.roomId ?? null

  if (!roomId.value) return

  if (connectedRoomId.value === roomId.value && mode.value === 'participant') {
    return
  }

  roomWebSocketStore.connect({ roomId: roomId.value, mode: 'participant' })
})
onBeforeUnmount(() => {
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
        :card="displayCard"
        cell-size="var(--participant-card-cell-size)"
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
  </div>
</template>

<style scoped>
.participant-room {
  box-sizing: border-box;
  position: fixed;
  inset: 0;
  display: grid;
  width: 100%;
  height: 100dvh;
  overflow: hidden;
  grid-template-columns: minmax(0, 1fr) minmax(300px, 360px);
  gap: 18px;
  padding: 18px;
  background:
    radial-gradient(circle at 12% 18%, rgb(176 226 255 / 0.75), transparent 34%),
    radial-gradient(circle at 82% 16%, rgb(255 197 223 / 0.7), transparent 32%),
    radial-gradient(circle at 24% 82%, rgb(194 255 231 / 0.65), transparent 36%),
    radial-gradient(circle at 78% 76%, rgb(211 198 255 / 0.7), transparent 34%),
    linear-gradient(135deg, #f7fbff 0%, #fff8fd 48%, #f4fff9 100%);
}

.participant-room__card {
  --participant-card-cell-size: clamp(70px, min(14.8vw, 14dvh), 112px);

  display: flex;
  min-width: 0;
  min-height: 0;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  padding-bottom: 0;
}

.participant-room__chat {
  min-width: 0;
  height: 100%;
  min-height: 0;
  overflow: hidden;
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

@media (max-width: 639px) {
  .participant-room {
    display: flex;
    height: 100dvh;
    align-items: center;
    justify-content: center;
    padding: 8px 8px 68px;
  }

  .participant-room__card {
    --participant-card-cell-size: clamp(48px, min(15.5vw, 11dvh), 72px);

    width: 100%;
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
    height: auto;
    min-height: 0;
  }
}
</style>
