<template>
  <div class="page">
    <main ref="displayPageElement" class="display-page" aria-labelledby="display-page-title">
      <h1 id="display-page-title" class="sr-only">ビンゴ抽選画面</h1>
      <Iridescence
        class="display-page__background"
        :color="[1, 1, 1]"
        :mouse-react="false"
        :amplitude="0.1"
        :speed="0.1"
      />
      <div class="display-page__stage">
        <div v-if="isActiveRoom" class="display-page__content">
          <p class="sr-only" role="status" aria-live="polite" aria-atomic="true">
            {{ latestDrawAnnouncement }}
          </p>
          <div class="display-page__latest-ball">
            <NumberBall
              class="display-page__latest-number"
              :ball-color="displayBallPalette.picked"
              :text-color="displayBallPalette.text"
              :text="displayBallText"
            />
          </div>
          <div v-if="isGameWaiting" class="display-page__waiting-panel">
            <p class="display-page__waiting-title">ゲームはまだ始まっていません</p>
            <p class="display-page__waiting-text">参加者の準備ができるまでお待ちください</p>
            <p role="status" aria-live="polite" aria-atomic="true">
              参加者: {{ roomWebSocketStore.participantCount ?? 0 }} 人
            </p>
            <div class="max-h-24 overflow-auto" role="list" aria-label="参加者一覧">
              <span
                v-for="(user, index) in roomWebSocketStore.participants"
                :key="user.userId"
                role="listitem"
              >
                {{ user.userId
                }}<span v-if="index < roomWebSocketStore.participants.length - 1" aria-hidden="true"
                  >、</span
                >
              </span>
            </div>
          </div>
          <BallStateGrid v-else :picked-balls="pickedBalls" :latest-ball="latestBall" />
          <RoomStatsBar />
        </div>
        <div class="display-page__chat">
          <div class="display-page__chat-header">Chat</div>
          <ChatContainer
            :key="props.roomCode"
            :room-code="props.roomCode"
            :textarea="false"
            :connect="false"
            variant="display"
          />
        </div>
      </div>
      <UButton
        class="display-page__fullscreen-button"
        color="neutral"
        variant="soft"
        size="xl"
        :icon="isFullscreen ? 'i-lucide-minimize' : 'i-lucide-expand'"
        :aria-label="isFullscreen ? '全画面表示を終了' : '全画面表示'"
        @click="toggleFullscreen"
      />
      <DisplayParticipantQrCode :room-code="props.roomCode" :open="qrCodeVisible ?? false" />
      <ConnectionBadge />

      <GameStartCutin
        v-if="showCutin"
        :key="cutinKey"
        @complete="handleCutinComplete"
        class="z-1"
        :bottom-text="bottomText"
        :top-text="topText"
      />
    </main>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { storeToRefs } from 'pinia'

import Iridescence from '@/components/backgrounds/Iridescence.vue'
import ConnectionBadge from '@/components/common/ConnectionBadge.vue'
import NumberBall from '@/components/common/NumberBall.vue'
import BallStateGrid from '@/components/display/BallStateGrid.vue'
import DisplayParticipantQrCode from '@/components/display/DisplayParticipantQrCode.vue'
import RoomStatsBar from '@/components/display/RoomStatsBar.vue'
import ChatContainer from '@/components/chat/ChatContainer.vue'
import { useDisplayEffects } from '@/composables/useDisplayEffects'
import { useNumberBallFavicon } from '@/composables/useNumberBallFavicon'
import { usePickRollingEffect } from '@/composables/usePickRollingEffect'
import { useRoomsStore } from '@/stores/rooms'
import { useRoomWebSocketStore } from '@/stores/roomWebSocket'

import GameStartCutin from '@/components/GameStartCutin.vue'

const roomsStore = useRoomsStore()
const roomWebSocketStore = useRoomWebSocketStore()
const { latestBall, latestEvent, mode, pickedBalls, qrCodeVisible, roomState } =
  storeToRefs(roomWebSocketStore)
const { isActiveRoom } = storeToRefs(roomsStore)
const props = defineProps<{ roomCode: string }>()
const displayPageElement = ref<HTMLElement | null>(null)
const isFullscreen = ref(false)
const latestDrawAnnouncement = ref('')
const { displayBallPalette, displayBallText } = usePickRollingEffect()
const { bottomText, cutinKey, handleCutinComplete, showCutin, topText } = useDisplayEffects()
useNumberBallFavicon()

const isGameWaiting = computed(() => roomState.value === 'waiting')

watch(latestEvent, (event) => {
  if (isActiveRoom.value && event?.type === 'PickFinished') {
    latestDrawAnnouncement.value = `数字 ${event.body.pickedBall} が抽選されました。`
  }
})

function syncFullscreenState() {
  isFullscreen.value = document.fullscreenElement === displayPageElement.value
}

async function toggleFullscreen() {
  if (document.fullscreenElement) {
    await document.exitFullscreen()
    return
  }

  await displayPageElement.value?.requestFullscreen()
}

watch(
  () => props.roomCode,
  (roomCode) => {
    latestDrawAnnouncement.value = ''
    void roomsStore.connectToRoom(roomCode, { mode: 'display' })
  },
  { immediate: true },
)

onMounted(() => {
  document.addEventListener('fullscreenchange', syncFullscreenState)
})

onBeforeUnmount(() => {
  document.removeEventListener('fullscreenchange', syncFullscreenState)

  if (mode.value === 'display') {
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
  background: #ffffff;
}

.display-page:fullscreen {
  background: #ffffff;
}

.display-page:fullscreen::backdrop {
  background: #ffffff;
}

.display-page__background {
  position: absolute;
  inset: 0;
  z-index: -2;
  opacity: 0.6;
  pointer-events: none;
}

.display-page__stage {
  position: relative;
  z-index: 1;
  box-sizing: border-box;
  display: grid;
  grid-template-columns: minmax(0, 1fr) clamp(310px, 27vw, 440px);
  gap: 20px;
  height: 100vh;
  padding: 2%;
  overflow: hidden;
}

.display-page__stage::before {
  position: absolute;
  inset: 0;
  z-index: 0;
  background: rgb(255 255 255 / 0.26);
  backdrop-filter: blur(1px) saturate(1.15);
  content: '';
}

.display-page__content {
  position: relative;
  z-index: 1;
  display: flex;
  min-width: 0;
  min-height: 0;
  flex-direction: column;
  align-items: center;
  justify-content: flex-start;
  overflow: hidden;
  padding: 1% 1% 0;
  gap: 6%;
}

.display-page__latest-ball {
  width: 100%;
  height: 34%;
  display: grid;
  place-items: center;
  container-type: size;
}

.display-page__latest-number {
  width: min(90cqw, 90cqh);
  height: auto;
  opacity: 0.8;
  aspect-ratio: 1 / 1;
  font-size: min(14dvh, 38cqw);
  box-shadow:
    0 10px 15px -3px rgb(0 0 0 / 0.1),
    0 4px 6px -4px rgb(0 0 0 / 0.1);
}

.display-page__chat {
  box-sizing: border-box;
  position: relative;
  z-index: 1;
  display: flex;
  min-width: 0;
  min-height: 0;
  flex-direction: column;
  overflow: hidden;
  border: 1px solid rgb(255 255 255 / 0.5);
  border-radius: 18px;
  background: rgb(255 255 255 / 0.46);
  box-shadow: 0 18px 45px rgb(20 55 110 / 0.16);
  backdrop-filter: blur(14px) saturate(1.25);
}

.display-page__chat-header {
  position: absolute;
  top: 0;
  right: 0;
  left: 0;
  z-index: 2;
  padding: 16px 18px 12px;
  border-bottom: 1px solid rgb(255 255 255 / 0.48);
  background: rgb(255 255 255 / 0.82);
  backdrop-filter: blur(10px) saturate(1.2);
  color: #1f4f8f;
  font-size: 18px;
  font-weight: 800;
  line-height: 1;
}

.display-page__fullscreen-button {
  position: absolute;
  top: 18px;
  left: 18px;
  z-index: 4;
  box-shadow: 0 12px 28px rgb(20 55 110 / 0.16);
  backdrop-filter: blur(12px) saturate(1.2);
  opacity: 0.6;
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

@media (max-width: 720px) {
  .display-page__stage {
    grid-template-columns: minmax(0, 1fr);
    grid-template-rows: minmax(0, 1.25fr) minmax(0, 1fr);
    gap: 12px;
    padding: 12px;
  }

  .display-page__content {
    gap: 3%;
  }

  .display-page__latest-ball {
    height: 28%;
  }

  :deep(.ball-state-grid__cell),
  :deep(.ball-state-grid__number) {
    min-width: 0;
    min-height: 0;
  }

  :deep(.ball-state-grid) {
    grid-template-columns: repeat(15, minmax(12px, 1fr));
    gap: 2px;
  }

  .display-page__waiting-title {
    font-size: 24px;
  }

  .display-page__waiting-text {
    font-size: 16px;
  }

  .room-stats-bar {
    height: 56px;
    gap: 14px;
    padding: 7px 20px;
  }
}

@media (max-width: 720px) and (max-height: 500px) {
  .display-page__stage {
    grid-template-columns: minmax(0, 1fr) minmax(220px, 38vw);
    grid-template-rows: minmax(0, 1fr);
    gap: 12px;
    padding: 12px;
  }

  .display-page__content {
    gap: 3%;
  }

  .display-page__latest-ball {
    height: 28%;
  }

  .display-page__waiting-panel {
    min-height: 0;
    gap: 6px;
    padding: 12px;
  }

  .display-page__waiting-title {
    font-size: 20px;
  }

  .display-page__waiting-text {
    font-size: 14px;
  }

  .display-page__waiting-panel > div {
    max-height: 42px;
  }
}

@media (max-width: 720px) and (min-height: 501px) and (max-height: 650px) {
  .display-page__stage {
    grid-template-rows: minmax(0, 1.2fr) minmax(0, 0.8fr);
  }

  .display-page__content {
    gap: 1%;
  }

  .display-page__latest-ball {
    height: 12%;
  }

  .display-page__latest-number {
    font-size: min(6dvh, 38cqw);
  }

  :deep(.ball-state-grid) {
    width: min(100%, 84vw);
  }

  .display-page__waiting-panel {
    min-height: 0;
    gap: 4px;
    padding: 8px;
  }

  .display-page__waiting-title {
    font-size: 18px;
  }

  .display-page__waiting-text {
    font-size: 12px;
  }

  .display-page__waiting-panel > div {
    max-height: 24px;
  }

  .room-stats-bar {
    height: 48px;
  }
}
</style>
