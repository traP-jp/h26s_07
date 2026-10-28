<script setup lang="ts">
import RoomDeletedNotice from '@/components/common/RoomDeletedNotice.vue'
import { ref, onMounted, onBeforeUnmount, watch } from 'vue'
import { apiClient } from '@/api/apiClient'
import { useRoomsStore } from '@/stores/rooms'
import { useRoomWebSocketStore } from '@/stores/roomWebSocket'
import ConnectionBadge from '@/components/common/ConnectionBadge.vue'

const roomsStore = useRoomsStore()
const connection = useRoomWebSocketStore()
const props = defineProps<{ roomCode: string }>()
const loading = ref(true)
const errorMessage = ref('')
let roomId = ''
const isDisabled = ref(false)
const gameStarted = ref(true)
const pickStarted = ref(false)
const pickExhausted = ref(true)
const qrCodeShowed = ref(false)
const startModal = ref(false)
const finishModal = ref(false)
const finishedModal = ref(false)
const ready = ref(false)

function applyState(state: string, pickState: string, qrCodeVisible: boolean) {
  gameStarted.value = state !== 'waiting'
  pickStarted.value = pickState === 'picking'
  pickExhausted.value = pickState === 'exhausted'
  qrCodeShowed.value = qrCodeVisible
  finishedModal.value = state === 'finished'
  ready.value = true
}

watch(
  () => [connection.roomState, connection.pickState, connection.qrCodeVisible] as const,
  ([state, pick, qr]) => {
    if (
      connection.roomId === roomId &&
      connection.mode === 'display' &&
      state &&
      pick &&
      qr !== null
    ) {
      applyState(state, pick, qr)
    }
  },
)

async function getLatestInfo() {
  try {
    const { data, error, response } = await apiClient.GET('/api/rooms/{roomId}', {
      params: { path: { roomId } },
    })
    if (response.status === 404) connection.markRoomDeleted()
    if (error || !data) throw new Error(error?.message ?? 'ルーム情報を取得できませんでした。')
    applyState(data.state, data.pickState, data.qrCodeVisible)
    return data
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : '通信に失敗しました。'
    return null
  }
}

type ControlPath =
  | '/api/rooms/{roomId}/control/start'
  | '/api/rooms/{roomId}/control/finish'
  | '/api/rooms/{roomId}/control/pick/start'
  | '/api/rooms/{roomId}/control/pick/finish'
  | '/api/rooms/{roomId}/control/pick/cancel'
  | '/api/rooms/{roomId}/control/qrcode/show'
  | '/api/rooms/{roomId}/control/qrcode/hide'
type RoomInfo = NonNullable<Awaited<ReturnType<typeof getLatestInfo>>>

async function control(path: ControlPath, matches: (state: RoomInfo) => boolean) {
  if (isDisabled.value || !ready.value) return
  isDisabled.value = true
  errorMessage.value = ''
  let failure = ''
  try {
    const { error } = await apiClient.POST(path, { params: { path: { roomId } } })
    if (error) failure = error.message
  } catch (error) {
    failure = error instanceof Error ? error.message : '操作に失敗しました。'
  } finally {
    const state = await getLatestInfo()
    if (failure && (!state || !matches(state))) errorMessage.value = failure
    startModal.value = false
    finishModal.value = false
    isDisabled.value = false
  }
}

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
const clickStartModal = () =>
  control('/api/rooms/{roomId}/control/start', (s) => s.state === 'playing')
const clickFinishModal = () =>
  control('/api/rooms/{roomId}/control/finish', (s) => s.state === 'finished')
const clickPickStart = () =>
  control('/api/rooms/{roomId}/control/pick/start', (s) => s.pickState === 'picking')
const clickPickFinish = () =>
  control('/api/rooms/{roomId}/control/pick/finish', (s) => s.pickState !== 'picking')
const clickPickCancel = () =>
  control('/api/rooms/{roomId}/control/pick/cancel', (s) => s.pickState === 'idle')
const clickShowQr = () => control('/api/rooms/{roomId}/control/qrcode/show', (s) => s.qrCodeVisible)
const clickHideQr = () =>
  control('/api/rooms/{roomId}/control/qrcode/hide', (s) => !s.qrCodeVisible)

let lastKeyAt = 0
function onKey(event: KeyboardEvent) {
  if (
    event.isComposing ||
    event.ctrlKey ||
    event.metaKey ||
    event.altKey ||
    (event.target instanceof Element &&
      event.target.closest('input, textarea, select, [contenteditable="true"]'))
  )
    return
  const key = event.key.toLowerCase()
  const toggle = [' ', 'enter', 'pagedown', 'arrowright', 'pageup', 'arrowleft'].includes(key)
  const cancel = ['escape', 'backspace'].includes(key)
  if (!toggle && !cancel && key !== 'q') return
  if (
    startModal.value ||
    finishModal.value ||
    finishedModal.value ||
    document.querySelector('[role="dialog"]')
  )
    return
  event.preventDefault()
  if (
    event.repeat ||
    Date.now() - lastKeyAt < 400 ||
    isDisabled.value ||
    loading.value ||
    !ready.value
  )
    return
  lastKeyAt = Date.now()
  if (key === 'q') void (qrCodeShowed.value ? clickHideQr() : clickShowQr())
  else if (gameStarted.value && !pickExhausted.value) {
    if (cancel && pickStarted.value) void clickPickCancel()
    else if (toggle) void (pickStarted.value ? clickPickFinish() : clickPickStart())
  }
}

const savedStyles: Array<{
  element: HTMLElement
  overflow: string
  overscrollBehavior: string
  touchAction: string
}> = []
let disposed = false
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
  window.addEventListener('keydown', onKey)
  try {
    roomId = (await roomsStore.getRoomIdByCode(props.roomCode)) ?? ''
    if (disposed) return
    if (!roomId) throw new Error('コードに対応するルームが見つかりませんでした。')
    connection.connect({ roomId, mode: 'display' })
    await getLatestInfo()
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : '通信に失敗しました。'
  } finally {
    loading.value = false
  }
})
onBeforeUnmount(() => {
  disposed = true
  window.removeEventListener('keydown', onKey)
  for (const { element, ...style } of savedStyles) Object.assign(element.style, style)
  if (connection.roomId === roomId && connection.mode === 'display') connection.disconnect()
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
