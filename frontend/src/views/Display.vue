<template>
  <div class="display-page">
    <Iridescence
      class="display-page__background"
      :color="[1, 1, 1]"
      :mouse-react="false"
      :amplitude="0.1"
      :speed="0.1"
    />
    <div class="display-page__content">
      <div class="display-page__latest-ball" aria-label="直近の抽選番号">
        <NumberBall
          class="display-page__latest-number"
          :ball-color="displayBallColor"
          :text-color="displayBallTextColor"
          :text="displayBallText"
          :size="260"
        />
      </div>
      <div v-if="isGameWaiting" class="display-page__waiting-panel" role="status">
        <p class="display-page__waiting-title">ゲームはまだ始まっていません</p>
        <p class="display-page__waiting-text">参加者の準備ができるまでお待ちください</p>
      </div>
      <BallStateGrid v-else :picked-balls="pickedBalls" :latest-picked-ball="latestPickedBall" />
      <RoomStatsBar />
    </div>
    <DisplayParticipantQrCode :room-code="props.roomCode" :open="qrCodeVisible ?? false" />
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { storeToRefs } from 'pinia'

import type { RoomId } from '@/api/schema'
import type { PickedBall } from '@/api/schema'
import Iridescence from '@/components/backgrounds/Iridescence.vue'
import NumberBall from '@/components/layouts/NumberBall.vue'
import BallStateGrid from '@/components/display/BallStateGrid.vue'
import { getBallPalette } from '@/components/display/ballPalette'
import DisplayParticipantQrCode from '@/components/display/DisplayParticipantQrCode.vue'
import RoomStatsBar from '@/components/display/RoomStatsBar.vue'
import { useSoundEffect } from '@/composables/useSoundEffect'
import { useRoomsStore } from '@/stores/rooms'
import { useRoomWebSocketStore } from '@/stores/roomWebSocket'

const roomsStore = useRoomsStore()
const roomWebSocketStore = useRoomWebSocketStore()
const {
  latestEvent,
  latestPickedBall,
  mode,
  pickState,
  pickedBalls,
  qrCodeVisible,
  roomId: connectedRoomId,
  roomState,
} = storeToRefs(roomWebSocketStore)
const props = defineProps<{ roomCode: string }>()
const roomId = ref<RoomId | null>(null)
const rollingPickedBall = ref<PickedBall | null>(null)
const drumroll = useSoundEffect('drumroll', { loop: true })
const cymbal = useSoundEffect('cymbal')
let rollingTimerId: number | null = null

const isGameWaiting = computed(() => roomState.value === 'waiting')

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
      drumroll.stop()
      return
    }

    drumroll.play()
    rollingPickedBall.value = Math.floor(Math.random() * 75 + 1) as PickedBall
    rollingTimerId = window.setInterval(() => {
      rollingPickedBall.value = Math.floor(Math.random() * 75 + 1) as PickedBall
    }, 60)
  },
  { immediate: true },
)

watch(latestEvent, (event) => {
  if (!event) return
  if (event.type !== 'PickFinished') return

  cymbal.play()
})

onMounted(async () => {
  if (!props.roomCode) return

  const room = await roomsStore.getRoomByCode(props.roomCode)
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
  position: relative;
  isolation: isolate;
  height: 100vh;
  overflow: hidden;
}

.display-page__background {
  position: absolute;
  inset: 0;
  z-index: -2;
  opacity: 0.6;
  pointer-events: none;
}

.display-page__content {
  position: relative;
  z-index: 1;
  display: flex;
  height: 100vh;
  flex-direction: column;
  align-items: center;
  justify-content: flex-start;
  overflow: hidden;
  padding: 3% 3% 2%;
  gap: 6%;
}

.display-page__content::before {
  position: absolute;
  inset: 0;
  z-index: -1;
  background: rgb(255 255 255 / 0.26);
  backdrop-filter: blur(1px) saturate(1.15);
  content: '';
}

.display-page__latest-ball {
  width: 100%;
  height: 34%;
  display: grid;
  place-items: center;
  container-type: size;
}

.display-page__latest-number {
  width: min(90cqw, 90cqh) !important;
  height: auto !important;
  opacity: 0.8;
  aspect-ratio: 1 / 1;
  font-size: min(14dvh, 38cqw) !important;
  box-shadow:
    0 10px 15px -3px rgb(0 0 0 / 0.1),
    0 4px 6px -4px rgb(0 0 0 / 0.1);
}

.display-page__waiting-panel {
  box-sizing: border-box;
  width: 100%;
  max-width: 774px;
  min-height: 186px;
  display: grid;
  align-content: center;
  justify-items: center;
  gap: 10px;
  padding: 24px;
  border: 1px solid rgb(255 255 255 / 0.48);
  border-radius: 18px;
  background: rgb(255 255 255 / 0.42);
  box-shadow: 0 18px 45px rgb(20 55 110 / 0.16);
  backdrop-filter: blur(14px) saturate(1.25);
  text-align: center;
}

.display-page__waiting-title {
  margin: 0;
  font-size: 32px;
  font-weight: 800;
  line-height: 1.25;
  color: #1f4f8f;
}

.display-page__waiting-text {
  margin: 0;
  font-size: 18px;
  font-weight: 700;
  line-height: 1.4;
  color: #43678f;
}

@media (max-width: 520px) {
  .display-page__waiting-panel {
    min-height: 132px;
    padding: 18px;
    border-radius: 14px;
  }

  .display-page__waiting-title {
    font-size: 22px;
  }

  .display-page__waiting-text {
    font-size: 14px;
  }
}
</style>
