<script setup lang="ts">
import { nextTick, onMounted, ref, watch } from 'vue'
import type { Uuid, Message, DateTime, WebSocketMode } from '@/api/schema'
import { useRoomWebSocketStore } from '@/stores/roomWebSocket'
import { useRoomsStore } from '@/stores/rooms'

const room = withDefaults(
  defineProps<{
    roomCode: string
    textarea?: boolean
    connect?: boolean
  }>(),
  {
    textarea: false,
    connect: true,
  },
)

const messages = ref<Message[]>([])
const chatContainer = ref<HTMLElement | null>(null)

const scrollToBottom = async () => {
  await nextTick()
  const element = chatContainer.value
  if (!element) return

  element.scrollTop = element.scrollHeight
}

const addUserMessage = (m: Message) => {
  messages.value.push(m)
  void scrollToBottom()
}

const addSpecialMessage = (id: Uuid, content: string, createdAt: DateTime) => {
  messages.value.push({
    messageId: `${id}-${messages.value.length}` as Uuid,
    content: content,
    author: { userId: '' },
    createdAt: createdAt,
  })
  void scrollToBottom()
}

const store = useRoomWebSocketStore()
const roomsStore = useRoomsStore()

onMounted(async () => {
  if (!room.connect) return

  const roomId = await roomsStore.getRoomIdByCode(room.roomCode)
  if (!roomId) return

  const mode: WebSocketMode = room.textarea ? 'participant' : 'display'
  if (store.isActiveConnection({ roomId, mode })) return

  store.connect({ roomId, mode })
})

watch(
  () => store.latestMessage,
  (newValue) => {
    if (newValue) {
      addUserMessage(newValue)
    }
  },
)
watch(
  () => store.pickState,
  (newValue) => {
    if (newValue == 'exhausted') {
      addSpecialMessage('allPicked', '球が枯渇しました！', 'ima')
    }
  },
)
watch(
  () => store.latestNewBingos,
  (newValue) => {
    if (newValue) {
      if (newValue.length >= 2) {
        addSpecialMessage(
          'newBingos',
          `${newValue[0]?.user} と他 ${newValue.length - 1} 人がビンゴしました！`,
          'ima',
        )
      } else {
        addSpecialMessage('newBingos', `${newValue[0]?.user} がビンゴしました！`, 'ima')
      }
    }
  },
)
watch(
  () => store.latestNewReaches,
  (newValue) => {
    if (newValue) {
      if (newValue.length >= 2) {
        addSpecialMessage(
          'newReaches',
          `${newValue[0]?.user} と他 ${newValue.length - 1} 人がリーチしました！`,
          'ima',
        )
      } else {
        addSpecialMessage('newReaches', `${newValue[0]?.user} がリーチしました！`, 'ima')
      }
    }
  },
)
</script>

<template>
  <div ref="chatContainer" class="chat-container">
    <div v-for="message in messages" :key="message.messageId">
      <MessageContainer
        :user-id="message.author.userId"
        :content="message.content"
      ></MessageContainer>
    </div>
  </div>
  <div v-if="room.textarea">
    <PostMessage :room-code="room.roomCode"></PostMessage>
  </div>
</template>

<style scoped>
.chat-container {
  min-height: calc(100% - 50px);
  overflow: auto;
  scrollbar-width: none;
}
.chat-container::-webkit-scrollbar {
  display: none;
}
</style>
