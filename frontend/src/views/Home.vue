<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import type { Room, UserId } from '@/api/schema'

import { useRoomsStore } from '@/stores/rooms'
import { useCurrentUserStore } from '@/stores/currentUser'

const roomsStore = useRoomsStore()
const currentUserStore = useCurrentUserStore()

const loading = ref(true)
const errorMessage = ref('')
const showFinished = ref(false)
const visibleRooms = computed(() =>
  rooms.value.filter((room) => showFinished.value || room.state !== 'finished'),
)
const rooms = computed(() => roomsStore.rooms)
const currentUserId = ref<UserId>()

const deleteModalOpen = ref(false)
const deletingRoom = ref<Room | null>(null)
const deleting = ref(false)
const deleteError = ref('')
function openDeleteModal(room: Room) {
  deletingRoom.value = room
  deleteError.value = ''
  deleteModalOpen.value = true
}
async function confirmDeleteRoom() {
  if (!deletingRoom.value) return
  deleting.value = true
  deleteError.value = ''
  try {
    await roomsStore.deleteRoom(deletingRoom.value.roomId)
    deleteModalOpen.value = false
    deletingRoom.value = null
  } catch (error) {
    deleteError.value = error instanceof Error ? error.message : 'ルームの削除に失敗しました。'
  } finally {
    deleting.value = false
  }
}

const joinModalOpen = ref(false)
const selectedRoom = ref<Room | null>(null)

onMounted(async () => {
  loading.value = true

  currentUserId.value = await currentUserStore.getUserId()

  try {
    await roomsStore.init(true)
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : 'ルーム情報の取得に失敗しました。'
  }

  loading.value = false
})

const joinedRooms = computed(() => {
  return visibleRooms.value.filter((room) =>
    room.participants.some((participant) => participant.user.userId === currentUserId.value),
  )
})

const adminRooms = computed(() => {
  return visibleRooms.value.filter((room) =>
    room.settings.admins.some((admin) => admin.userId === currentUserId.value),
  )
})

const waitingRooms = computed(() => {
  return visibleRooms.value.filter(
    (room) =>
      room.state == 'waiting' &&
      !room.participants.some((participant) => participant.user.userId === currentUserId.value),
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
    <div v-if="loading">読み込み中...</div>

    <div v-else-if="errorMessage">
      ルーム情報の取得に失敗しました。 <br />
      {{ errorMessage }}
    </div>

    <template v-else>
      <UCheckbox v-model="showFinished" label="終了したルームも表示" class="mb-4" />
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
        <div v-for="room in adminRooms" :key="room.roomCode" class="space-y-2">
          <Card
            :title="room.settings.name"
            :description="room.settings.description"
            :room-code="room.roomCode"
            is-Admin
          >
          </Card>
          <UButton
            color="error"
            variant="soft"
            icon="i-lucide-trash-2"
            @click="openDeleteModal(room)"
            >ルームを削除</UButton
          >
        </div>
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
    </template>

    <UModal
      v-model:open="deleteModalOpen"
      title="ルームを削除しますか？この操作は取り消せません"
      :description="deletingRoom?.settings.name"
      :dismissible="!deleting"
      :close="!deleting"
    >
      <template #body
        ><p v-if="deleteError" class="text-error">{{ deleteError }}</p></template
      >
      <template #footer>
        <UButton
          color="neutral"
          variant="outline"
          :disabled="deleting"
          @click="deleteModalOpen = false"
          >キャンセル</UButton
        >
        <UButton color="error" :loading="deleting" @click="confirmDeleteRoom">削除</UButton>
      </template>
    </UModal>
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
