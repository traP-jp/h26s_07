<template>
  <section class="room-stats-bar" aria-label="ルーム状況">
    <UBadge v-for="stat in stats" :key="stat.label" class="room-stats-bar__item" color="primary">
      <span class="room-stats-bar__label">{{ stat.label }}：</span>
      <span class="room-stats-bar__value">{{ stat.value }}人</span>
    </UBadge>
  </section>
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
  min-height: 52px;
  width: min(100%, 848px);
  align-items: center;
  justify-content: center;
  gap: clamp(16px, 3.2vw, 28px);
  border-radius: 14px;
  background: rgb(245 250 255 / 0.88);
  padding: 8px clamp(18px, 4vw, 42px);
  box-shadow: 0 12px 36px rgb(53 89 138 / 0.12);
  backdrop-filter: blur(14px);
}

.room-stats-bar__item {
  min-width: 112px;
  justify-content: center;
  border: 0;
  background: transparent;
  padding: 0;
  font-size: 22px;
  font-weight: 800;
  line-height: 1.2;
  letter-spacing: 0;
  color: #1d5aa5;
}

.room-stats-bar__label,
.room-stats-bar__value {
  white-space: nowrap;
}

@media (max-width: 520px) {
  .room-stats-bar {
    gap: 10px;
    padding-inline: 12px;
  }

  .room-stats-bar__item {
    min-width: 0;
    font-size: 16px;
  }
}
</style>
