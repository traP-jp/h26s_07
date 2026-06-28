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

const addSpecialMessage = (id: Uuid, content: string, createdAt: DateTime, userId?: string) => {
  messages.value.push({
    messageId: `${id}-${messages.value.length}` as Uuid,
    content: content,
    author: { userId: userId ?? '' },
    createdAt: createdAt,
  })
  void scrollToBottom()
}

const store = useRoomWebSocketStore()
const roomsStore = useRoomsStore()

const notificationType = (message: Message) => {
  if (message.messageId.startsWith('newBingos')) return 'bingo'
  if (message.messageId.startsWith('newReaches')) return 'reach'
  if (message.messageId.startsWith('allPicked')) return 'notice'
  return undefined
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
      if (newValue.length >= 1) {
        addSpecialMessage(
          'newBingos',
          `${newValue.map((bingo) => bingo.user.userId).join('，')} がビンゴしました！`,
          'ima',
          newValue[0]?.user.userId,
        )
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
          newValue[0]?.user.userId,
        )
      } else if (newValue.length == 1) {
        addSpecialMessage(
          'newReaches',
          `${newValue[0]?.user.userId} がリーチしました！`,
          'ima',
          newValue[0]?.user.userId,
        )
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
    <div class="chat-container__messages">
      <div v-for="message in messages" :key="message.messageId">
        <MessageContainer
          :user-id="message.author.userId"
          :content="message.content"
          :special="message.createdAt == 'ima'"
          :notification-type="notificationType(message)"
        ></MessageContainer>
      </div>
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

.chat-container__messages {
  display: flex;
  min-height: 100%;
  flex-direction: column;
  justify-content: flex-end;
}

#chatContainer.chat-container--display {
  flex: 1 1 auto;
  height: 100%;
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
