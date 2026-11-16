<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import type { RoomId, ActivityLogEntry, Message, DateTime, WebSocketMode } from '@/api/schema'
import { apiClient } from '@/api/apiClient'
import MessageContainer from '@/components/chat/MessageContainer.vue'
import PostMessage from '@/components/chat/PostMessage.vue'
import { useRoomWebSocketStore } from '@/stores/roomWebSocket'
import { useRoomsStore } from '@/stores/rooms'
import { storeToRefs } from 'pinia'

const {
  roomCode,
  textarea = false,
  connect = true,
  variant = 'default',
} = defineProps<{
  roomCode: string
  textarea?: boolean
  connect?: boolean
  variant?: 'default' | 'display' | 'participant'
}>()

const historyMessages = ref<Message[]>([])
const historyStatus = ref<'loading' | 'ready' | 'error' | 'forbidden'>('loading')
const messages = computed(() => [
  ...historyMessages.value,
  ...(store.roomId === activeRoomId.value ? store.newChatMessages : []),
])

const chatContainer = ref<HTMLElement | null>(null)
const participantInlineChatContainer = ref<HTMLElement | null>(null)
const participantDrawerChatContainer = ref<HTMLElement | null>(null)
const activeRoomId = ref<RoomId | null>(null)
const participantChatOpen = ref(false)
const participantChatClosing = ref(false)
let participantChatClosingTimer: ReturnType<typeof setTimeout> | undefined
let historyRequestGeneration = 0
let isUnmounted = false

const store = useRoomWebSocketStore()
const roomsStore = useRoomsStore()
const { activityLog } = storeToRefs(store)

type NotificationType = 'bingo' | 'reach' | 'notice'
type TimelineMessage = Message & { notificationType?: NotificationType }

const sortedMessages = <T extends Message>(items: T[]) =>
  [...items].sort((a, b) => {
    const diff = new Date(a.createdAt).getTime() - new Date(b.createdAt).getTime()
    if (diff !== 0) return diff
    return a.messageId.localeCompare(b.messageId)
  })

const mergeMessages = <T extends Message>(items: T[]): T[] => {
  const messageById = new Map<string, T>()
  for (const message of items) {
    messageById.set(message.messageId, message)
  }
  return sortedMessages([...messageById.values()])
}

// messageId は重複排除と v-for の key に必要なだけで、表示には使わない。
const systemMessage = (
  content: string,
  createdAt: DateTime,
  notificationType: NotificationType,
  key: string,
): TimelineMessage => ({
  messageId: `${notificationType}:${key}`,
  content,
  author: { userId: '' },
  createdAt,
  notificationType,
})

type ActivityGroup = {
  key: string
  createdAt: DateTime
  entries: ActivityLogEntry[]
}

// 同じ抽選の記録は pickBall が同じになる。
const groupActivityEntriesByPick = (entries: ActivityLogEntry[]): ActivityGroup[] => {
  const groups = new Map<string, ActivityGroup>()
  for (const entry of entries) {
    const key = `ball:${entry.pickBall}`
    const group = groups.get(key)
    if (group) {
      group.entries.push(entry)
    } else {
      groups.set(key, { key, createdAt: entry.createdAt, entries: [entry] })
    }
  }
  return [...groups.values()]
}

const bingoMessages = (entries: ActivityLogEntry[]) =>
  groupActivityEntriesByPick(entries).map(({ key, createdAt, entries }) =>
    systemMessage(
      `${entries.map((entry) => entry.user.userId).join('、')} がビンゴしました！`,
      createdAt,
      'bingo',
      key,
    ),
  )

const reachMessages = (entries: ActivityLogEntry[]) =>
  groupActivityEntriesByPick(entries).map(({ key, createdAt, entries }) => {
    const content =
      entries.length >= 2
        ? `${entries[0]?.user.userId} と他 ${entries.length - 1} 人がリーチしました！`
        : `${entries[0]?.user.userId} がリーチしました！`

    return systemMessage(content, createdAt, 'reach', key)
  })

const summaryMessages = computed(() => {
  // lineCount は参加者の通算ライン数なので、1 がその人の初回ビンゴ。
  // ビンゴ通知は一人一回に絞り、同時に複数ラインが成立しても1件だけ出す。
  const bingoEntries = activityLog.value.filter(
    (entry) => entry.kind === 'bingo' && (entry.lineCount ?? 1) === 1,
  )
  const reachEntries = activityLog.value.filter((entry) => entry.kind === 'reach')
  return [...bingoMessages(bingoEntries), ...reachMessages(reachEntries)]
})

const timelineMessages = computed<TimelineMessage[]>(() =>
  mergeMessages<TimelineMessage>([
    ...messages.value,
    ...summaryMessages.value,
    ...(store.roomId === activeRoomId.value && store.pickState === 'exhausted'
      ? [
          systemMessage(
            '球が枯渇しました！',
            summaryMessages.value.at(-1)?.createdAt ??
              roomsStore.rooms.find((item) => item.roomId === activeRoomId.value)?.createdAt ??
              '1970-01-01T00:00:00.000Z',
            'notice',
            `exhausted:${activeRoomId.value}`,
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

const loadMessages = async (roomId: RoomId) => {
  if (isUnmounted) return

  const requestGeneration = ++historyRequestGeneration
  const isCurrentRequest = () =>
    !isUnmounted && requestGeneration === historyRequestGeneration && activeRoomId.value === roomId

  historyStatus.value = 'loading'
  try {
    const response = await apiClient.GET('/api/rooms/{roomId}/chats', {
      params: { path: { roomId } },
    })

    if (response.response.status === 403) {
      if (isCurrentRequest()) {
        historyMessages.value = []
        historyStatus.value = 'forbidden'
      }
      return
    }

    if (response.data && isCurrentRequest()) {
      // 履歴APIは全件返すため、再取得しても同じIDは重ねない。
      historyMessages.value = mergeMessages([...historyMessages.value, ...response.data])
      historyStatus.value = 'ready'
      void scrollToBottom()
    } else if (isCurrentRequest()) {
      historyStatus.value = 'error'
    }
  } catch {
    if (isCurrentRequest()) historyStatus.value = 'error'
  }
}

const initializeChat = async () => {
  if (isUnmounted) return

  historyStatus.value = 'loading'
  try {
    const roomId = await roomsStore.getRoomIdByCode(roomCode)
    if (isUnmounted) return
    if (!roomId) throw new Error('Room not found')
    activeRoomId.value = roomId

    if (connect) {
      const mode: WebSocketMode = textarea ? 'participant' : 'display'
      if (!store.isActiveConnection({ roomId, mode })) store.connect({ roomId, mode })
    }
    await loadMessages(roomId)
  } catch {
    if (!isUnmounted) historyStatus.value = 'error'
  }
}

const retryLoadMessages = () => {
  if (activeRoomId.value) {
    void loadMessages(activeRoomId.value)
  } else {
    void initializeChat()
  }
}

onMounted(() => void initializeChat())

watch(
  () => store.latestEvent,
  (event) => {
    // 初期化後にも履歴を取り直し、初回取得との間の投稿や切断中の投稿を補う。
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
  isUnmounted = true
  historyRequestGeneration += 1
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
  <div v-if="variant === 'participant'" class="participant-chat">
    <section class="participant-chat__desktop">
      <div class="participant-chat__header">Chat</div>
      <div ref="participantInlineChatContainer" class="chat-container chat-container--participant">
        <div class="chat-container__messages">
          <div v-if="historyStatus === 'forbidden'" role="status">
            過去のチャット履歴を表示する権限がありません
          </div>
          <div v-else-if="historyStatus === 'error'" role="status">
            <span>チャットを読み込めませんでした</span>
            <button type="button" @click="retryLoadMessages">再読み込み</button>
          </div>
          <div
            v-else-if="historyStatus === 'loading' && timelineMessages.length === 0"
            role="status"
          >
            チャットを読み込んでいます…
          </div>
          <div v-for="message in timelineMessages" :key="message.messageId">
            <MessageContainer
              :user-id="message.author.userId"
              :content="message.content"
              :notification-type="message.notificationType"
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
              <div v-if="historyStatus === 'forbidden'" role="status">
                過去のチャット履歴を表示する権限がありません
              </div>
              <div v-else-if="historyStatus === 'error'" role="status">
                <span>チャットを読み込めませんでした</span>
                <button type="button" @click="retryLoadMessages">再読み込み</button>
              </div>
              <div
                v-else-if="historyStatus === 'loading' && timelineMessages.length === 0"
                role="status"
              >
                チャットを読み込んでいます…
              </div>
              <div v-for="message in timelineMessages" :key="message.messageId">
                <MessageContainer
                  :user-id="message.author.userId"
                  :content="message.content"
                  :notification-type="message.notificationType"
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
      :class="{ 'chat-container--display': variant === 'display' }"
    >
      <div class="chat-container__messages">
        <div v-if="historyStatus === 'forbidden'" role="status">
          過去のチャット履歴を表示する権限がありません
        </div>
        <div v-else-if="historyStatus === 'error'" role="status">
          <span>チャットを読み込めませんでした</span>
          <button type="button" @click="retryLoadMessages">再読み込み</button>
        </div>
        <div v-else-if="historyStatus === 'loading' && timelineMessages.length === 0" role="status">
          チャットを読み込んでいます…
        </div>
        <div v-for="message in timelineMessages" :key="message.messageId">
          <MessageContainer
            :user-id="message.author.userId"
            :content="message.content"
            :notification-type="message.notificationType"
          ></MessageContainer>
        </div>
      </div>
    </div>

    <div v-if="textarea && activeRoomId" style="height: 30px">
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
