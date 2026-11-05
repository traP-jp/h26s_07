import { computed, ref, watch, type Ref } from 'vue'
import { storeToRefs } from 'pinia'

import type { Message, Uuid } from '@/api/schema'
import { apiClient } from '@/api/apiClient'
import { useRoomWebSocketStore } from '@/stores/roomWebSocket'
import { useRoomsStore } from '@/stores/rooms'

type RoomWebSocketStore = ReturnType<typeof useRoomWebSocketStore>
type RoomsStore = ReturnType<typeof useRoomsStore>

type ChatTimelineOptions = {
  activeRoomId: Ref<Uuid | null>
  store: RoomWebSocketStore
  roomsStore: RoomsStore
  onHistoryLoaded: () => void
}

function sortMessages(messages: Message[]) {
  return [...messages].sort((left, right) => {
    const diff = new Date(left.createdAt).getTime() - new Date(right.createdAt).getTime()
    if (diff !== 0) return diff
    return left.messageId.localeCompare(right.messageId)
  })
}

function mergeMessages(messages: Message[]) {
  const messageById = new Map<string, Message>()
  for (const message of messages) {
    messageById.set(message.messageId, message)
  }
  return sortMessages([...messageById.values()])
}

function createSystemMessage(
  id: string,
  content: string,
  createdAt: Message['createdAt'],
): Message {
  return {
    messageId: id as Uuid,
    content,
    author: { userId: '' },
    createdAt,
  }
}

export function useChatTimeline({
  activeRoomId,
  store,
  roomsStore,
  onHistoryLoaded,
}: ChatTimelineOptions) {
  const historyMessages = ref<Message[]>([])
  const { activityLog, initializationCount } = storeToRefs(store)

  const messages = computed(() => [
    ...historyMessages.value,
    ...(store.roomId === activeRoomId.value ? store.chatMessages : []),
  ])

  const summaryMessages = computed(() =>
    activityLog.value.map((entry) =>
      createSystemMessage(
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

  const timelineMessages = computed(() => {
    const exhaustedNotice =
      store.roomId === activeRoomId.value && store.pickState === 'exhausted'
        ? createSystemMessage(
            `exhausted:${activeRoomId.value}`,
            '球が枯渇しました！',
            // AllPicked has no timestamp, so anchor the stable notice to the latest known activity time.
            summaryMessages.value.at(-1)?.createdAt ??
              roomsStore.rooms.find((item) => item.roomId === activeRoomId.value)?.createdAt ??
              '1970-01-01T00:00:00.000Z',
          )
        : null

    return mergeMessages([
      ...messages.value,
      ...summaryMessages.value,
      ...(exhaustedNotice ? [exhaustedNotice] : []),
    ])
  })

  const latestMessagePreview = computed(() => {
    const latestMessage = timelineMessages.value.at(-1)
    if (!latestMessage) return 'チャット'
    if (latestMessage.author.userId === '') return latestMessage.content
    return `${latestMessage.author.userId}: ${latestMessage.content}`
  })

  async function loadMessages(roomId: Uuid) {
    const response = await apiClient.GET('/api/rooms/{roomId}/chats', {
      params: { path: { roomId } },
    })

    // Guests receive 403 on the initial fetch; WebSocket chat still supplies later messages.
    if (response.response.status === 403) return

    if (response.data && activeRoomId.value === roomId) {
      historyMessages.value = mergeMessages([...historyMessages.value, ...response.data])
      onHistoryLoaded()
    }
  }

  watch(
    initializationCount,
    () => {
      if (activeRoomId.value && store.roomId === activeRoomId.value) {
        void loadMessages(activeRoomId.value)
      }
    },
    { flush: 'sync' },
  )

  const notificationType = (message: Message) => {
    if (message.author.userId !== '') return undefined
    return activityLog.value.find((entry) => entry.id === message.messageId)?.kind ?? 'notice'
  }

  return { latestMessagePreview, loadMessages, notificationType, timelineMessages }
}
