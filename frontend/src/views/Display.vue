<template>
  <div class="flex min-h-screen items-center justify-center bg-[#f8fbff] px-4 py-6">
    <RoomStatsBar />
  </div>
</template>

<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from 'vue'
import { storeToRefs } from 'pinia'
import { useRoute } from 'vue-router'

import type { RoomCode, RoomId } from '@/api/schema'
import RoomStatsBar from '@/components/rooms/RoomStatsBar.vue'
import { useRoomsStore } from '@/stores/rooms'
import { useRoomWebSocketStore } from '@/stores/roomWebSocket'

const route = useRoute()
const roomsStore = useRoomsStore()
const roomWebSocketStore = useRoomWebSocketStore()
const { mode, roomId: connectedRoomId } = storeToRefs(roomWebSocketStore)
const { roomsByCode } = storeToRefs(roomsStore)

const roomCode = route.params.roomCode as RoomCode | undefined
const roomId = ref<RoomId | null>(null)

onMounted(async () => {
  if (!roomCode) return

  await roomsStore.init()
  roomId.value = roomsByCode.value.get(roomCode)?.roomId ?? null

  if (!roomId.value) return

  if (connectedRoomId.value === roomId.value && mode.value === 'display') {
    return
  }

  roomWebSocketStore.connect({ roomId: roomId.value, mode: 'display' })
})

onBeforeUnmount(() => {
  if (roomId.value && connectedRoomId.value === roomId.value) {
    roomWebSocketStore.disconnect()
  }
})
</script>
