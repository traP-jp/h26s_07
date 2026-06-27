<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import type { Room } from '@/api/schema'

import { apiClient } from '@/api/apiClient'
import { useCurrentUserStore } from '@/stores/currentUser'

const currentUserStore = useCurrentUserStore()

const loading = ref(true)
const errorMessage = ref('')
const rooms = ref(new Array<Room>())

onMounted(async () => {
  loading.value = true

  const { data, error } = await apiClient.GET('/api/rooms')
  if (error) {
    errorMessage.value = error.message
  } else {
    rooms.value = data
  }

  loading.value = false
})

const joinedRooms = computed(() => {
  return rooms.value.filter((room) =>
    room.participants.some((participant) => participant.user.userId === currentUserStore.userId),
  )
})

const adminRooms = computed(() => {
  return rooms.value.filter((room) =>
    room.settings.admins.some((admin) => admin.userId === currentUserStore.userId),
  )
})

const waitingRooms = computed(() => {
  return rooms.value.filter(
    (room) =>
      room.state == 'waiting' &&
      !room.participants.some((participant) => participant.user.userId === currentUserStore.userId),
  )
})
</script>

<template>
  <UContainer>
    <div v-if="loading">読み込み中...</div>
    <div v-else-if="errorMessage">
      ルーム情報の取得に失敗しました。 <br />
      {{ errorMessage }}
    </div>
    <template v-else>
      <h2 class="text-3xl font-bold mb-2">あなたが参加しているルーム</h2>
      <div class="grid grid-cols-1 gap-2 md:grid-cols-2 mb-6">
        <template v-for="room in joinedRooms" :key="room.roomCode">
          <Card
            :title="room.settings.name"
            :description="room.settings.description"
            :room-code="room.roomCode"
          ></Card>
        </template>
      </div>

      <h2 class="text-3xl font-bold mb-2">あなたが管理しているルーム</h2>
      <div class="grid grid-cols-1 gap-2 md:grid-cols-2 mb-6">
        <template v-for="room in adminRooms" :key="room.roomCode">
          <Card
            :title="room.settings.name"
            :description="room.settings.description"
            :room-code="room.roomCode"
            isAdmin
          ></Card>
        </template>
      </div>

      <h2 class="text-3xl font-bold mb-2">参加者を募集しているルーム</h2>
      <div class="grid grid-cols-1 gap-2 md:grid-cols-2 mb-6">
        <template v-for="room in waitingRooms" :key="room.roomCode">
          <Card
            :title="room.settings.name"
            :description="room.settings.description"
            :room-code="room.roomCode"
          ></Card>
        </template>
      </div>
    </template>
  </UContainer>
</template>
