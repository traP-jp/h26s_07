<template>
  <div class="room-stats-bar">
    <div v-for="stat in stats" :key="stat.label" class="room-stats-bar__item">
      <span class="room-stats-bar__label">{{ stat.label }}</span>
      <span class="room-stats-bar__value">
        <span>{{ stat.value }}</span>
        <span class="room-stats-bar__unit">人</span>
      </span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { storeToRefs } from 'pinia'

import { useRoomWebSocketStore } from '@/stores/roomWebSocket'

const roomWebSocketStore = useRoomWebSocketStore()
const { bingoSummaries, participantCount, reachSummaries } = storeToRefs(roomWebSocketStore)

const stats = computed(() => [
  {
    label: '参加',
    value: participantCount.value ?? 0,
  },
  {
    label: 'ビンゴ',
    value: bingoSummaries.value.length,
  },
  {
    label: 'リーチ',
    value: reachSummaries.value.length,
  },
])
</script>

<style scoped>
.room-stats-bar {
  display: flex;
  box-sizing: border-box;
  width: 100%;
  max-width: 360px;
  height: 65px;
  align-items: center;
  justify-content: center;
  gap: 26px;
  padding: 7px 36px;
  border: 1px solid rgb(255 255 255 / 0.46);
  border-radius: 999px;
  background: rgb(255 255 255 / 0.28);
  box-shadow:
    inset 0 1px 0 rgb(255 255 255 / 0.58),
    0 16px 42px rgb(73 41 142 / 0.12);
  backdrop-filter: blur(12px) saturate(1.18);
}

.room-stats-bar__item {
  box-sizing: border-box;
  flex: 1 1 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 3px;
  min-width: 62px;
  color: #1f4f8f;
  line-height: 1;
}

.room-stats-bar__label,
.room-stats-bar__value {
  white-space: nowrap;
}

.room-stats-bar__label {
  color: #43678f;
  font-size: 0.75em;
  font-weight: 700;
}

.room-stats-bar__value {
  display: inline-flex;
  align-items: baseline;
  gap: 1px;
  min-width: 2.4em;
  justify-content: center;
  color: #1f4f8f;
  font-size: 1.8125em;
  font-weight: 800;
  letter-spacing: 0;
}

.room-stats-bar__unit {
  font-size: 0.68em;
}
</style>
