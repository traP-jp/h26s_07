<template>
  <div class="room-stats-bar" aria-label="ルーム状況">
    <UBadge v-for="stat in stats" :key="stat.label" class="room-stats-bar__item" color="primary">
      <span class="room-stats-bar__text">{{ stat.label }}：</span>
      <span class="room-stats-bar__text">{{ stat.value }}人</span>
    </UBadge>
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
    label: '参加者',
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
  min-height: 46px;
  width: 100%;
  max-width: 742px;
  align-items: center;
  justify-content: center;
  gap: 3%;
  border-radius: 12px;
  background: rgb(245 250 255 / 0.9);
  padding: 6px 3%;
  backdrop-filter: blur(8px);
}

.room-stats-bar__item {
  min-width: 96px;
  justify-content: center;
  border: 0;
  background: transparent;
  padding: 0;
  font-size: 20px;
  font-weight: 800;
  line-height: 1.25;
  letter-spacing: 0;
  color: #1d5aa5;
}

.room-stats-bar__text {
  white-space: nowrap;
}

@media (max-width: 520px) {
  .room-stats-bar__item {
    min-width: 0;
    font-size: 16px;
  }
}
</style>
