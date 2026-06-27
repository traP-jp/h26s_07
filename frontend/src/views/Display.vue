<template>
  <div class="display-page">
    <div class="display-page__latest-ball" aria-label="直近の抽選番号">
      <NumberBall
        class="display-page__latest-number"
        :ball-color="displayBallColor"
        :text-color="displayBallTextColor"
        :text="displayBallText"
        :size="260"
      />
    </div>
    <BallStateGrid :picked-balls="pickedBalls" :latest-picked-ball="latestPickedBall" />
    <RoomStatsBar />
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { storeToRefs } from 'pinia'
import { useRoute } from 'vue-router'

import type { RoomCode, RoomId } from '@/api/schema'
import type { PickedBall } from '@/api/schema'
import NumberBall from '@/components/layouts/NumberBall.vue'
import BallStateGrid from '@/components/rooms/BallStateGrid.vue'
import { getBallPalette } from '@/components/rooms/ballPalette'
import RoomStatsBar from '@/components/rooms/RoomStatsBar.vue'
import { useRoomsStore } from '@/stores/rooms'
import { useRoomWebSocketStore } from '@/stores/roomWebSocket'

const route = useRoute()
const roomsStore = useRoomsStore()
const roomWebSocketStore = useRoomWebSocketStore()
const {
  latestPickedBall,
  mode,
  pickState,
  pickedBalls,
  roomId: connectedRoomId,
  roomState,
} = storeToRefs(roomWebSocketStore)
const roomCode = route.params.roomCode as RoomCode | undefined
const roomId = ref<RoomId | null>(null)
const rollingPickedBall = ref<PickedBall | null>(null)
let rollingTimerId: number | null = null

const displayPickedBall = computed(() => {
  if (pickState.value === 'picking') {
    return rollingPickedBall.value
  }

  return latestPickedBall.value
})

const displayBallText = computed(() => {
  if (roomState.value == null || roomState.value === 'waiting' || displayPickedBall.value == null) {
    return '?'
  }

  return String(displayPickedBall.value)
})

const displayBallColor = computed(() => {
  if (displayPickedBall.value == null || displayBallText.value === '?') {
    return '#f1f6fb'
  }

  return getBallPalette(displayPickedBall.value).picked
})

const displayBallTextColor = computed(() => {
  if (displayBallText.value === '?') {
    return '#9aa8b7'
  }

  return '#ffffff'
})

function stopRollingPickedBall() {
  if (rollingTimerId == null) return

  window.clearInterval(rollingTimerId)
  rollingTimerId = null
}

watch(
  pickState,
  (nextPickState) => {
    stopRollingPickedBall()

    if (nextPickState !== 'picking') {
      rollingPickedBall.value = null
      return
    }

    rollingPickedBall.value = Math.floor(Math.random() * 75 + 1) as PickedBall
    rollingTimerId = window.setInterval(() => {
      rollingPickedBall.value = Math.floor(Math.random() * 75 + 1) as PickedBall
    }, 60)
  },
  { immediate: true },
)

onMounted(async () => {
  if (!roomCode) return

  const room = await roomsStore.getRoomByCode(roomCode)
  roomId.value = room?.roomId ?? null

  if (!roomId.value) return

  if (connectedRoomId.value === roomId.value && mode.value === 'display') {
    return
  }

  roomWebSocketStore.connect({ roomId: roomId.value, mode: 'display' })
})

onBeforeUnmount(() => {
  stopRollingPickedBall()

  if (roomId.value && connectedRoomId.value === roomId.value) {
    roomWebSocketStore.disconnect()
  }
})
</script>

<style scoped>
.display-page {
  display: flex;
  height: 100vh;
  flex-direction: column;
  align-items: center;
  justify-content: flex-start;
  overflow: hidden;
  padding: 3% 3% 2%;
  gap: 6%;
}

.display-page__latest-ball {
  width: 100%;
  height: 34%;
  display: grid;
  place-items: center;
}

.display-page__latest-number {
  width: auto !important;
  height: 90% !important;
  aspect-ratio: 1 / 1;
  font-size: 12dvh !important;
  box-shadow:
    0 10px 15px -3px rgb(0 0 0 / 0.1),
    0 4px 6px -4px rgb(0 0 0 / 0.1);
}
</style>
