<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref, watch } from 'vue'
import BingoCardPaper from '@/components/layouts/BingoCardPaper.vue'
import { useRoomWebSocketStore } from '@/stores/roomWebSocket'
import { useRoute } from 'vue-router'
import type { RoomCode, Card, RoomId } from '@/api/schema'
import { useRoomsStore } from '@/stores/rooms'
import { useCurrentUserStore } from '@/stores/currentUser'
import { storeToRefs } from 'pinia'

const route = useRoute()
const roomCode = route.params.roomCode as RoomCode | undefined
const roomId = ref<RoomId | null>(null)
const roomWebSocketStore = useRoomWebSocketStore()
const roomsStore = useRoomsStore()
const currentUserStore = useCurrentUserStore()
const { roomsByCode } = storeToRefs(roomsStore)
const { mode, roomId: connectedRoomId } = storeToRefs(roomWebSocketStore)

const { card, latestEvent } = storeToRefs(roomWebSocketStore)

const displayCard = ref<Card | null>(null)
const errorMessage = ref('')

onMounted(async () => {
  if (!roomCode) return

  await roomsStore.init()
  const room = roomsByCode.value.get(roomCode)
  roomId.value = room?.roomId ?? null

  if (!roomId.value) return

  if (connectedRoomId.value === roomId.value && mode.value === 'participant') {
    return
  }

  const isParticipant =
    room?.participants.some((participant) => participant.user.userId === currentUserStore.userId) ??
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
    <p v-if="errorMessage" class="text-red-500">{{ errorMessage }}</p>
    <BingoCardPaper v-if="displayCard" :card="displayCard" :size="90" />
  </div>
</template>

<style scoped>
.participant-room {
  display: flex;
  justify-content: center;
  align-items: center;
}
</style>
