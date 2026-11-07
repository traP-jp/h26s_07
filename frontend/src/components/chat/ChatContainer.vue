<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import type { Uuid, Message, DateTime, WebSocketMode } from '@/api/schema'
import { apiClient } from '@/api/apiClient'
import MessageContainer from '@/components/chat/MessageContainer.vue'
import PostMessage from '@/components/chat/PostMessage.vue'
import { useRoomWebSocketStore } from '@/stores/roomWebSocket'
import { useRoomsStore } from '@/stores/rooms'
import { storeToRefs } from 'pinia'

const room = withDefaults(
  defineProps<{
    roomCode: string
    textarea?: boolean
    connect?: boolean
    variant?: 'default' | 'display' | 'participant'
  }>(),
  {
    textarea: false,
    connect: true,
    variant: 'default',
  },
)

const historyMessages = ref<Message[]>([])
const messages = computed(() => [
  ...historyMessages.value,
  ...(store.roomId === activeRoomId.value ? store.chatMessages : []),
])

const chatContainer = ref<HTMLElement | null>(null)
const participantInlineChatContainer = ref<HTMLElement | null>(null)
const participantDrawerChatContainer = ref<HTMLElement | null>(null)
const activeRoomId = ref<Uuid | null>(null)
const participantChatOpen = ref(false)
const participantChatClosing = ref(false)
let participantChatClosingTimer: ReturnType<typeof setTimeout> | undefined

const store = useRoomWebSocketStore()
const roomsStore = useRoomsStore()
const { activityLog } = storeToRefs(store)

const sortedMessages = (items: Message[]) =>
  [...items].sort((a, b) => {
    const diff = new Date(a.createdAt).getTime() - new Date(b.createdAt).getTime()
    if (diff !== 0) return diff
    return a.messageId.localeCompare(b.messageId)
  })

const mergeMessages = (items: Message[]) => {
  const messageById = new Map<string, Message>()
  for (const message of items) {
    messageById.set(message.messageId, message)
  }
  return sortedMessages([...messageById.values()])
}

const systemMessage = (id: string, content: string, createdAt: DateTime): Message => ({
  messageId: id as Uuid,
  content,
  author: { userId: '' },
  createdAt,
})

const summaryMessages = computed(() =>
  activityLog.value.map((entry) =>
    systemMessage(
      entry.id,
      entry.kind === 'reach'
        ? `${entry.user.userId} さんがリーチ！`
        : (entry.lineCount ?? 1) > 1
          ? `${entry.user.userId} さんが ${entry.lineCount} ライン目のビンゴ！`
          : `${entry.user.userId} さんが ${entry.bingoRank ?? 1} 番目にビンゴ！`,
      entry.createdAt,
    ),
  ),
)

const timelineMessages = computed(() =>
  mergeMessages([
    ...messages.value,
    ...summaryMessages.value,
    ...(store.roomId === activeRoomId.value && store.pickState === 'exhausted'
      ? [
          systemMessage(
            `exhausted:${activeRoomId.value}`,
            '球が枯渇しました！',
            summaryMessages.value.at(-1)?.createdAt ??
              roomsStore.rooms.find((item) => item.roomId === activeRoomId.value)?.createdAt ??
              '1970-01-01T00:00:00.000Z',
          ),
        ]
      : []),
  ]),
)

const latestMessagePreview = computed(() => {
  const latestMessage = timelineMessages.value.at(-1)
  if (!latestMessage) return 'チャット'
  if (latestMessage.author.userId === '') return latestMessage.content
  return `${latestMessage.author.userId}: ${latestMessage.content}`
})

const scrollToBottom = async () => {
  await nextTick()
  for (const element of [
    chatContainer.value,
    participantInlineChatContainer.value,
    participantDrawerChatContainer.value,
  ]) {
    if (!element) continue
    element.scrollTop = element.scrollHeight
  }
}

const closeParticipantChat = () => {
  participantChatOpen.value = false
}

const loadMessages = async (roomId: Uuid) => {
  try {
    const response = await apiClient.GET('/api/rooms/{roomId}/chats', {
      params: { path: { roomId } },
    })

    if (response.response.status === 403) return

    if (response.data && activeRoomId.value === roomId) {
      historyMessages.value.push(...response.data)
      void scrollToBottom()
    }
  } catch {
    // 履歴取得に失敗しても、WebSocket の新着メッセージは表示する。
  }
}

const notificationType = (message: Message) => {
  if (message.author.userId !== '') return undefined
  return activityLog.value.find((entry) => entry.id === message.messageId)?.kind ?? 'notice'
}

onMounted(async () => {
  const roomId = await roomsStore.getRoomIdByCode(room.roomCode)
  if (!roomId) return
  activeRoomId.value = roomId

  if (room.connect) {
    const mode: WebSocketMode = room.textarea ? 'participant' : 'display'
    if (!store.isActiveConnection({ roomId, mode })) store.connect({ roomId, mode })
  }
  await loadMessages(roomId)
})

watch(
  () => store.latestEvent,
  (event) => {
    // 再接続後に、切断中の投稿をチャット履歴から補う。
    if (
      event?.type === 'Initialized' &&
      activeRoomId.value &&
      store.roomId === activeRoomId.value
    ) {
      void loadMessages(activeRoomId.value)
    }
  },
  { flush: 'sync' },
)

onBeforeUnmount(() => {
  if (participantChatClosingTimer) clearTimeout(participantChatClosingTimer)
})

watch(
  () => timelineMessages.value.length,
  () => {
    void scrollToBottom()
  },
)

watch(participantChatOpen, (open) => {
  if (participantChatClosingTimer) clearTimeout(participantChatClosingTimer)

  if (open) {
    participantChatClosing.value = false
    void scrollToBottom()
    return
  }

  participantChatClosing.value = true
  participantChatClosingTimer = setTimeout(() => {
    participantChatClosing.value = false
  }, 360)
})
</script>

<template>
  <div v-if="room.variant === 'participant'" class="participant-chat">
    <section class="participant-chat__desktop">
      <div class="participant-chat__header">Chat</div>
      <div ref="participantInlineChatContainer" class="chat-container chat-container--participant">
        <div class="chat-container__messages">
          <div v-for="message in timelineMessages" :key="message.messageId">
            <MessageContainer
              :user-id="message.author.userId"
              :content="message.content"
              :notification-type="notificationType(message)"
            ></MessageContainer>
          </div>
        </div>
      </div>
      <PostMessage v-if="activeRoomId" :room-id="activeRoomId" variant="light"></PostMessage>
    </section>

    <UDrawer
      v-model:open="participantChatOpen"
      direction="bottom"
      :handle="true"
      :ui="{
        content: 'bg-white p-0 sm:hidden h-[82dvh] max-h-[82dvh]',
      }"
    >
      <button
        class="participant-chat__mobile-peek"
        :class="{ 'participant-chat__mobile-peek--closing': participantChatClosing }"
        type="button"
      >
        <span class="participant-chat__mobile-handle"></span>
        <span class="participant-chat__mobile-peek-row">
          <span class="participant-chat__mobile-label">Chat</span>
          <span class="participant-chat__mobile-message">{{ latestMessagePreview }}</span>
        </span>
      </button>

      <template #content>
        <section class="participant-chat__drawer">
          <div class="participant-chat__drawer-header">
            <span>Chat</span>
            <UButton
              icon="i-lucide-x"
              color="neutral"
              variant="ghost"
              size="sm"
              @click="closeParticipantChat"
            />
          </div>
          <div
            ref="participantDrawerChatContainer"
            class="chat-container chat-container--participant"
          >
            <div class="chat-container__messages">
              <div v-for="message in timelineMessages" :key="message.messageId">
                <MessageContainer
                  :user-id="message.author.userId"
                  :content="message.content"
                  :notification-type="notificationType(message)"
                ></MessageContainer>
              </div>
            </div>
          </div>
          <PostMessage v-if="activeRoomId" :room-id="activeRoomId" variant="light"></PostMessage>
        </section>
      </template>
    </UDrawer>
  </div>

  <template v-else>
    <div
      ref="chatContainer"
      class="chat-container"
      :class="{ 'chat-container--display': room.variant === 'display' }"
    >
      <div class="chat-container__messages">
        <div v-for="message in timelineMessages" :key="message.messageId">
          <MessageContainer
            :user-id="message.author.userId"
            :content="message.content"
            :notification-type="notificationType(message)"
          ></MessageContainer>
        </div>
      </div>
    </div>

    <div v-if="room.textarea && activeRoomId" style="height: 30px">
      <PostMessage :room-id="activeRoomId"></PostMessage>
    </div>
  </template>
</template>

<style>
.chat-container {
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

.chat-container.chat-container--display {
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

.participant-chat {
  height: 100%;
  min-width: 0;
  min-height: 0;
  overflow: hidden;
}

.participant-chat__desktop {
  box-sizing: border-box;
  display: flex;
  width: 100%;
  height: 100%;
  min-height: 0;
  flex-direction: column;
  overflow: hidden;
  border: 1px solid rgb(35 63 105 / 0.12);
  border-radius: 12px;
  background: rgb(255 255 255 / 0.82);
  box-shadow: 0 14px 34px rgb(24 47 85 / 0.12);
}

.participant-chat__header,
.participant-chat__drawer-header {
  display: flex;
  min-height: 48px;
  align-items: center;
  justify-content: space-between;
  padding: 0 16px;
  border-bottom: 1px solid rgb(35 63 105 / 0.1);
  color: #1f3556;
  font-size: 16px;
  font-weight: 800;
}

.chat-container--participant {
  flex: 1 1 auto;
  height: auto;
  min-height: 0;
  overflow-y: auto;
  padding: 8px 0;
}

.chat-container--participant .nakami:not(.special) {
  font-size: 14px;
}

.chat-container--participant .nakami.special {
  font-size: 16px;
  line-height: 1.25;
}

.participant-chat__mobile-peek {
  position: fixed;
  right: 0;
  bottom: 0;
  left: 0;
  z-index: 20;
  display: none;
  height: calc(58px + env(safe-area-inset-bottom));
  min-width: 0;
  flex-direction: column;
  align-items: stretch;
  gap: 6px;
  padding: 8px 16px calc(8px + env(safe-area-inset-bottom));
  overflow: hidden;
  border: 0;
  border-top: 1px solid rgb(35 63 105 / 0.14);
  border-radius: 18px 18px 0 0;
  background: rgb(255 255 255 / 0.97);
  box-shadow: 0 -10px 30px rgb(24 47 85 / 0.18);
  color: #233f69;
  text-align: left;
}

.participant-chat__mobile-handle {
  width: 42px;
  height: 5px;
  flex: 0 0 auto;
  align-self: center;
  border-radius: 999px;
  background: rgb(35 63 105 / 0.24);
}

.participant-chat__mobile-peek-row {
  display: flex;
  min-width: 0;
  align-items: center;
  gap: 10px;
  transition: opacity 120ms ease;
}

.participant-chat__mobile-peek--closing .participant-chat__mobile-peek-row {
  opacity: 0;
}

.participant-chat__mobile-label {
  flex: 0 0 auto;
  font-size: 13px;
  font-weight: 900;
}

.participant-chat__mobile-message {
  min-width: 0;
  overflow: hidden;
  font-size: 14px;
  font-weight: 600;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.participant-chat__drawer {
  display: flex;
  width: 100%;
  height: 82dvh;
  min-height: 0;
  flex-direction: column;
  padding-bottom: env(safe-area-inset-bottom);
  background: #fff;
}

@media (max-width: 639px) {
  .participant-chat__desktop {
    display: none;
  }

  .participant-chat__mobile-peek {
    display: flex;
  }

  .chat-container--participant .nakami.special {
    font-size: 15px;
  }
}
</style>
