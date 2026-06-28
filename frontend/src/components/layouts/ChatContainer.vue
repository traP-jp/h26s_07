<script setup lang="ts">
import { nextTick, onMounted, ref, watch } from 'vue'
import type { Uuid, Message, DateTime, WebSocketMode } from '@/api/schema'
import { apiClient } from '@/api/apiClient'
import { useRoomWebSocketStore } from '@/stores/roomWebSocket'
import { useRoomsStore } from '@/stores/rooms'

const room = withDefaults(
  defineProps<{
    roomCode: string
    textarea?: boolean
    connect?: boolean
    variant?: 'default' | 'display'
  }>(),
  {
    textarea: false,
    connect: true,
    variant: 'default',
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

const addUserMessage = async (m: Message) => {
  messages.value.push(m)
  void scrollToBottom()
}

const loadMessages = async (roomId: Uuid) => {
  const response = await apiClient.GET('/api/rooms/{roomId}/chats', {
    params: { path: { roomId } },
  })

  if (response.data) {
    messages.value = response.data
    void scrollToBottom()
  }
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

const notificationType = (message: Message) => {
  if (message.author.userId !== '') return undefined
  if (message.messageId.startsWith('newBingos')) return 'bingo'
  if (message.messageId.startsWith('newReaches')) return 'reach'
  return 'notice'
}

onMounted(async () => {
  const roomId = await roomsStore.getRoomIdByCode(room.roomCode)
  if (!roomId) return

  await loadMessages(roomId)

  if (!room.connect) return

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
          `${newValue.map((bingo) => bingo.user.userId).join('、')} がビンゴしました！`,
          'ima',
        )
      } else if (newValue.length == 1) {
        addSpecialMessage('newBingos', `${newValue[0]?.user.userId} がビンゴしました！`, 'ima')
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
          `${newValue[0]?.user.userId} と他 ${newValue.length - 1} 人がリーチしました！`,
          'ima',
        )
      } else if (newValue.length == 1) {
        addSpecialMessage('newReaches', `${newValue[0]?.user.userId} がリーチしました！`, 'ima')
      }
    }
  },
)
</script>

<template>
  <div
    ref="chatContainer"
    id="chatContainer"
    :class="{ 'chat-container--display': room.variant === 'display' }"
  >
    <div v-for="message in messages" :key="message.messageId">
      <MessageContainer
        :user-id="message.author.userId"
        :content="message.content"
        :notification-type="notificationType(message)"
      ></MessageContainer>
    </div>
  </div>
  <div v-if="room.textarea" style="height: 30px">
    <PostMessage :room-code="room.roomCode"></PostMessage>
  </div>
</template>

<style>
#chatContainer {
  height: calc(100% - 50px);
  overflow-y: scroll;
  scrollbar-width: none;
}

.chat-container--display {
  flex: 1 1 auto;
  min-height: 0;
  padding: 8px 0 14px;
}

.chat-container--display .message {
  padding-right: 12px;
  padding-left: 12px;
}

.chat-container--display .nakami:not(.special) {
  background: rgb(255 255 255 / 0.86);
  border-color: rgb(56 114 177 / 0.28);
  color: #24364d;
  font-size: 14px;
}

.chat-container--display .nakami {
  opacity: 0.88;
}

.chat-container--display .nakami.special {
  font-size: 16px;
  line-height: 1.25;
}

.chat-container::-webkit-scrollbar {
  display: none;
}
</style>
