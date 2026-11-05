<script setup lang="ts">
import { useRoomWebSocketStore } from '@/stores/roomWebSocket'
const connection = useRoomWebSocketStore()
</script>

<template>
  <UModal
    :open="connection.roomDeleted"
    :dismissible="false"
    :close="false"
    title="このルームは削除されました"
  >
    <template #footer><UButton to="/" color="neutral">トップへ戻る</UButton></template>
  </UModal>
</template>
