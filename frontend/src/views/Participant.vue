<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import ConnectionBadge from '@/components/common/ConnectionBadge.vue'
import BingoCardPaper from '@/components/bingo/BingoCardPaper.vue'
import ChatContainer from '@/components/chat/ChatContainer.vue'
import { useRoomWebSocketStore } from '@/stores/roomWebSocket'
import { useRoute } from 'vue-router'
import type { RoomCode, Card, RoomId, UserId } from '@/api/schema'
import { useRoomsStore } from '@/stores/rooms'
import { useCurrentUserStore } from '@/stores/currentUser'
import { storeToRefs } from 'pinia'

const route = useRoute()
const roomCode = route.params.roomCode as RoomCode | undefined
const roomId = ref<RoomId | null>(null)
const roomWebSocketStore = useRoomWebSocketStore()
const roomsStore = useRoomsStore()
const currentUserStore = useCurrentUserStore()
const { roomId: connectedRoomId, roomState } = storeToRefs(roomWebSocketStore)

const { card, latestCardChanges } = storeToRefs(roomWebSocketStore)
const { latestBall } = storeToRefs(roomWebSocketStore)

const displayCard = ref<Card | null>(null)
const isGameWaiting = computed(() => roomState.value === 'waiting')
const errorMessage = ref('')
const currentUserId = ref<UserId>()

let disposed = false
onMounted(async () => {
  try {
    if (!roomCode) return

    currentUserId.value = await currentUserStore.getUserId()

    const room = await roomsStore.getRoomByCode(roomCode)
    roomId.value = room?.roomId ?? null

    if (!roomId.value) return

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

    if (disposed) return
    roomWebSocketStore.connect({ roomId: roomId.value, mode: 'participant' })
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : 'ルームに参加できませんでした。'
  }
})
onBeforeUnmount(() => {
  disposed = true

  if (roomId.value && connectedRoomId.value === roomId.value) {
    roomWebSocketStore.disconnect()
  }
})

watch(
  card,
  (value) => {
    displayCard.value = value
  },
  { immediate: true },
)
</script>

<template>
  <div class="participant-room">
    <ConnectionBadge />
    <UAlert v-if="errorMessage" :description="errorMessage" color="error" />
    <div class="participant-room__card">
      <div v-if="isGameWaiting" class="participant-room__waiting">
        <p class="participant-room__waiting-title">ゲームはまだ始まっていません</p>
        <p class="participant-room__waiting-text">開始されるまでお待ちください</p>
      </div>
      <BingoCardPaper
        v-else
        class="participant-room__bingo-card"
        :card="displayCard"
        :card-changes="latestCardChanges"
        :draw-key="latestBall"
        :placeholder="displayCard === null"
      />
    </div>
    <ChatContainer
      v-if="roomCode"
      class="participant-room__chat"
      :room-code="roomCode"
      textarea
      :connect="false"
      variant="participant"
    />
  </div>
</template>

<style scoped>
.participant-room {
  display: flex;
  height: 100dvh;
  padding: 18px;

  background: #e9e7df;
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
  width: clamp(300px, 28vw, 360px);
  flex-shrink: 0;
  height: 100%;
}

.participant-room__waiting {
  width: 60%;
  padding: 30px;
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

  .participant-room__bingo-card {
    width: min(94cqw, 68cqh);
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

@media (max-width: 639px) and (orientation: landscape) {
  .participant-room__bingo-card {
    width: min(94cqw, 76cqh);
  }
}
</style>
