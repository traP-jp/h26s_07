<script setup lang="ts">
import RoomDeletedNotice from '@/components/common/RoomDeletedNotice.vue'
import ConnectionBadge from '@/components/common/ConnectionBadge.vue'
import { useControllerRoomControls } from '@/composables/useControllerRoomControls'
import { useControllerShortcuts } from '@/composables/useControllerShortcuts'
import { useRoomsStore } from '@/stores/rooms'
import { useRoomWebSocketStore } from '@/stores/roomWebSocket'
import { ref, onMounted, onBeforeUnmount } from 'vue'

const roomsStore = useRoomsStore()
const connection = useRoomWebSocketStore()
const props = defineProps<{ roomCode: string }>()
const loading = ref(true)
const roomId = ref('')
const {
  clickFinishModal,
  clickHideQr,
  clickPickCancel,
  clickPickFinish,
  clickPickStart,
  clickShowQr,
  clickStartModal,
  errorMessage,
  finishModal,
  finishedModal,
  gameStarted,
  getLatestInfo,
  isDisabled,
  pickExhausted,
  pickStarted,
  qrCodeShowed,
  ready,
  startModal,
  dispose,
  isDisposed,
} = useControllerRoomControls(roomId, connection)

const clickGameStart = () => {
  startModal.value = true
}
const clickGameFinish = () => {
  finishModal.value = true
}
const cancelGameStart = () => {
  startModal.value = false
}
const cancelGameFinish = () => {
  finishModal.value = false
}

const shortcuts = useControllerShortcuts({
  isDisabled,
  loading,
  ready,
  startModal,
  finishModal,
  finishedModal,
  gameStarted,
  pickStarted,
  pickExhausted,
  qrCodeShowed,
  clickShowQr,
  clickHideQr,
  clickPickStart,
  clickPickFinish,
  clickPickCancel,
})

const savedStyles: Array<{
  element: HTMLElement
  overflow: string
  overscrollBehavior: string
  touchAction: string
}> = []
onMounted(async () => {
  for (const element of [document.documentElement, document.body]) {
    savedStyles.push({
      element,
      overflow: element.style.overflow,
      overscrollBehavior: element.style.overscrollBehavior,
      touchAction: element.style.touchAction,
    })
    Object.assign(element.style, {
      overflow: 'hidden',
      overscrollBehavior: 'none',
      touchAction: 'manipulation',
    })
  }
  shortcuts.attach()
  try {
    roomId.value = (await roomsStore.getRoomIdByCode(props.roomCode)) ?? ''
    if (isDisposed()) return
    if (!roomId.value) throw new Error('コードに対応するルームが見つかりませんでした。')
    connection.connect({ roomId: roomId.value, mode: 'display' })
    await getLatestInfo()
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : '通信に失敗しました。'
  } finally {
    loading.value = false
  }
})
onBeforeUnmount(() => {
  dispose()
  shortcuts.detach()
  for (const { element, ...style } of savedStyles) Object.assign(element.style, style)
  if (connection.roomId === roomId.value && connection.mode === 'display') connection.disconnect()
})
</script>

<template>
  <UModal
    v-model:open="startModal"
    :dismissible="false"
    title="本当にゲームを開始しますか？"
    description="ゲームを一度開始すると、以降、新たにプレイヤーがゲームに参加することはできません。"
    :ui="{ footer: 'justify-end' }"
    :close="false"
  >
    <template #footer>
      <UButton
        label="やっぱ開始しない"
        color="neutral"
        variant="outline"
        @click="cancelGameStart"
      />
      <UButton label="本当に開始する！" color="neutral" @click="clickStartModal()" />
    </template>
  </UModal>

  <UModal
    v-model:open="finishModal"
    :dismissible="false"
    title="本当にゲームを終了しますか？"
    description="ゲームを一度終了すると、再開することはできません。新たな球の抽選やチャットへの投稿はできなくなります。"
    :ui="{ footer: 'justify-end' }"
    :close="false"
  >
    <template #footer>
      <UButton
        label="やっぱ終了しない"
        color="neutral"
        variant="outline"
        @click="cancelGameFinish"
      />
      <UButton label="本当に終了する！" color="neutral" @click="clickFinishModal()" />
    </template>
  </UModal>

  <UModal
    :open="finishedModal && !connection.roomDeleted"
    :dismissible="false"
    title="このゲームは終了しました。"
    :ui="{ footer: 'justify-end' }"
    :close="false"
  >
    <template #footer>
      <UButton label="トップへ戻る" color="neutral" to="/" />
    </template>
  </UModal>

  <UContainer class="pt-6 touch-manipulation overscroll-none">
    <RoomDeletedNotice />
    <ConnectionBadge />
    <UAlert
      v-if="errorMessage"
      :description="errorMessage"
      color="error"
      close
      @update:open="errorMessage = ''"
      class="mb-2"
    />
    <div v-if="loading">読み込み中...</div>
    <template v-else>
      <h2 class="text-3xl font-bold mb-6">ルーム{{ props.roomCode }}の操作</h2>
      <p class="mb-4">参加者: {{ connection.participantCount ?? 0 }} 人</p>

      <p class="mb-2 text-xs text-muted">
        Space / Enter / ← → / PageUp・Down: 抽選開始・終了　Esc / Backspace: キャンセル　Q: QR切替
      </p>
      <div class="grid h-[65dvh] grid-rows-5 overflow-hidden gap-4">
        <template v-if="!gameStarted">
          <UButton
            variant="solid"
            class="row-span-4 min-h-0 h-full w-full overflow-hidden rounded-2xl text-6xl font-extrabold leading-none grid place-items-center"
            @click="clickGameStart()"
            :disabled="isDisabled || !ready || finishedModal"
          >
            <span class="flex flex-col items-center justify-center gap-4">
              <UIcon name="i-lucide-power" class="size-32" />
              <div>
                <span class="whitespace-nowrap">ゲーム</span>
                <span class="whitespace-nowrap">を</span>
                <span class="whitespace-nowrap">開始</span>
              </div>
            </span>
          </UButton>
        </template>

        <template v-else>
          <template v-if="!pickStarted">
            <UButton
              variant="solid"
              class="row-span-3 min-h-0 h-full w-full overflow-hidden rounded-2xl text-6xl font-extrabold leading-none grid place-items-center"
              @click="clickPickStart()"
              v-if="!pickExhausted"
              :disabled="isDisabled || !ready || finishedModal"
            >
              <span class="flex flex-col items-center justify-center gap-4">
                <UIcon name="i-lucide-circle-play" class="size-32" />
                <div>
                  <span class="whitespace-nowrap">抽選</span>
                  <span class="whitespace-nowrap">を</span>
                  <span class="whitespace-nowrap">開始</span>
                </div>
              </span>
            </UButton>
            <UButton
              variant="solid"
              class="row-span-3 min-h-0 h-full w-full overflow-hidden rounded-2xl text-6xl font-extrabold leading-none grid place-items-center"
              disabled
              v-else
            >
              <span class="flex flex-col items-center justify-center gap-4">
                <UIcon name="i-lucide-x" class="size-32" />
                <div>
                  <span class="whitespace-nowrap">抽選</span>
                  <span class="whitespace-nowrap">を</span>
                  <span class="whitespace-nowrap">開始</span>
                </div>
              </span>
            </UButton>
          </template>

          <template v-else>
            <UButton
              variant="solid"
              class="row-span-2 min-h-0 h-full w-full overflow-hidden rounded-2xl text-[32px] font-extrabold leading-none grid place-items-center"
              @click="clickPickFinish()"
              :disabled="isDisabled || !ready || finishedModal"
            >
              <span class="flex flex-col items-center justify-center gap-4">
                <UIcon name="i-lucide-circle-pause" class="size-24" />
                <div>
                  <span class="whitespace-nowrap">抽選</span>
                  <span class="whitespace-nowrap">を</span>
                  <span class="whitespace-nowrap">終了</span>
                </div>
              </span>
            </UButton>

            <UButton
              variant="outline"
              class="min-h-0 h-full w-full overflow-hidden rounded-2xl text-[32px] font-extrabold leading-none grid place-items-center"
              @click="clickPickCancel()"
              :disabled="isDisabled || !ready || finishedModal"
            >
              <span class="flex flex-col items-center justify-center gap-4">
                <div>
                  <span class="whitespace-nowrap">抽選</span>
                  <span class="whitespace-nowrap">を</span>
                  <span class="whitespace-nowrap">キャンセル</span>
                </div>
              </span>
            </UButton>
          </template>

          <UButton
            variant="outline"
            class="min-h-0 h-full w-full overflow-hidden rounded-2xl text-[32px] font-extrabold leading-none grid place-items-center"
            @click="clickGameFinish()"
            :disabled="isDisabled || !ready || finishedModal"
          >
            <span class="flex flex-col items-center justify-center gap-4">
              <div>
                <span class="whitespace-nowrap">ゲーム</span>
                <span class="whitespace-nowrap">を</span>
                <span class="whitespace-nowrap">終了</span>
              </div>
            </span>
          </UButton>
        </template>

        <template v-if="qrCodeShowed">
          <UButton
            variant="soft"
            class="min-h-0 h-full w-full overflow-hidden rounded-2xl text-[32px] font-extrabold leading-none grid place-items-center"
            @click="clickHideQr()"
            :disabled="isDisabled || !ready || finishedModal"
          >
            <span class="flex flex-col items-center justify-center gap-4">
              <div>
                <UIcon
                  name="i-lucide-qr-code"
                  class="size-8 inline-flex shrink-0 translate-y-[-0.1em] mr-1"
                />
                <span class="whitespace-nowrap">QR</span>
                <span class="whitespace-nowrap">コード</span>
                <span class="whitespace-nowrap">を</span>
                <span class="whitespace-nowrap">非表示</span>
              </div>
            </span>
          </UButton>
        </template>
        <template v-else>
          <UButton
            variant="soft"
            class="min-h-0 h-full w-full overflow-hidden rounded-2xl text-[32px] font-extrabold leading-none grid place-items-center"
            @click="clickShowQr()"
            :disabled="isDisabled || !ready || finishedModal"
          >
            <span class="flex flex-col items-center justify-center gap-4">
              <div>
                <UIcon
                  name="i-lucide-qr-code"
                  class="size-8 inline-flex shrink-0 translate-y-[-0.1em] mr-1"
                />
                <span class="whitespace-nowrap">QR</span>
                <span class="whitespace-nowrap">コード</span>
                <span class="whitespace-nowrap">を</span>
                <span class="whitespace-nowrap">表示</span>
              </div>
            </span>
          </UButton>
        </template>
      </div>
    </template>
  </UContainer>
</template>
