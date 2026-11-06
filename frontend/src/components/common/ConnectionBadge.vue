<script setup lang="ts">
import { useRoomWebSocketStore } from '@/stores/roomWebSocket'
const connection = useRoomWebSocketStore()
</script>

<template>
  <div
    v-if="connection.status !== 'open'"
    role="status"
    class="fixed top-2 right-2 z-50 rounded-full bg-amber-100 px-3 py-1 text-xs text-amber-900"
  >
    再接続中… <span v-if="connection.reconnectAttempt">({{ connection.reconnectAttempt }})</span>
  </div>
</template>
