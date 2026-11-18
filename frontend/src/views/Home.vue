<script setup lang="ts">
import { ref, computed } from 'vue'
import type { Room } from '@/api/schema'

import { useCurrentUserStore } from '@/stores/currentUser'
import { useRoomsStore } from '@/stores/rooms'
import { storeToRefs } from 'pinia'

const currentUserStore = useCurrentUserStore()
const roomsStore = useRoomsStore()
const { rooms } = storeToRefs(roomsStore)

const currentUserId = currentUserStore.userId

const joinModalOpen = ref(false)
const selectedRoom = ref<Room | null>(null)

const joinedRooms = computed(() => {
  return rooms.value.filter((room) =>
    room.participants.some((participant) => participant.user.userId === currentUserId),
  )
})

const adminRooms = computed(() => {
  return rooms.value.filter((room) =>
    room.settings.admins.some((admin) => admin.userId === currentUserId),
  )
})

const waitingRooms = computed(() => {
  return rooms.value.filter(
    (room) =>
      room.state == 'waiting' &&
      !room.participants.some((participant) => participant.user.userId === currentUserId),
  )
})

function openJoinModal(room: Room) {
  selectedRoom.value = room
  joinModalOpen.value = true
}

function closeJoinModal() {
  joinModalOpen.value = false
  selectedRoom.value = null
}

function confirmJoinRoom() {
  if (!selectedRoom.value) return

  const room = selectedRoom.value

  joinModalOpen.value = false
  selectedRoom.value = null

  window.open(`/${room.roomCode}/participant`, '_blank', 'noopener,noreferrer')
}
</script>

<template>
  <UContainer class="pt-6">
    <h2 class="text-3xl font-bold mb-4">
      <UIcon name="i-lucide-zap" class="text-yellow-400 inline" />
      あなたが参加しているルーム
    </h2>

    <div class="grid grid-cols-1 gap-4 md:grid-cols-2 mb-8">
      <template v-for="room in joinedRooms" :key="room.roomCode">
        <Card
          :title="room.settings.name"
          :description="room.settings.description"
          :room-code="room.roomCode"
        ></Card>
      </template>
    </div>

    <h2 class="text-3xl font-bold mb-4">あなたが管理しているルーム</h2>

    <div class="grid grid-cols-1 gap-4 md:grid-cols-2 mb-8">
      <template v-for="room in adminRooms" :key="room.roomCode">
        <Card
          :title="room.settings.name"
          :description="room.settings.description"
          :room-code="room.roomCode"
          is-Admin
        ></Card>
      </template>
    </div>

    <h2 class="text-3xl font-bold mb-4">参加者を募集しているルーム</h2>

    <div class="grid grid-cols-1 gap-4 md:grid-cols-2 mb-8">
      <template v-for="room in waitingRooms" :key="room.roomCode">
        <Card
          :title="room.settings.name"
          :description="room.settings.description"
          :room-code="room.roomCode"
          is-joinable
          @join="openJoinModal(room)"
        ></Card>
      </template>
    </div>

    <UModal v-model:open="joinModalOpen">
      <template #content>
        <UCard>
          <template #header>
            <h3 class="text-xl font-bold">参加しますか？</h3>
          </template>

          <div class="space-y-2">
            <p>「{{ selectedRoom?.settings.name }}」に参加しますか？</p>
          </div>

          <template #footer>
            <div class="flex justify-end gap-2">
              <UButton color="neutral" variant="soft" @click="closeJoinModal"> しない </UButton>

              <UButton color="primary" @click="confirmJoinRoom"> する </UButton>
            </div>
          </template>
        </UCard>
      </template>
    </UModal>
  </UContainer>
</template>
