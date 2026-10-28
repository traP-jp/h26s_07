import { ref, watch, type Ref } from 'vue'

import { apiClient } from '@/api/apiClient'
import { useRoomWebSocketStore } from '@/stores/roomWebSocket'

type ControlPath =
  | '/api/rooms/{roomId}/control/start'
  | '/api/rooms/{roomId}/control/finish'
  | '/api/rooms/{roomId}/control/pick/start'
  | '/api/rooms/{roomId}/control/pick/finish'
  | '/api/rooms/{roomId}/control/pick/cancel'
  | '/api/rooms/{roomId}/control/qrcode/show'
  | '/api/rooms/{roomId}/control/qrcode/hide'

type RoomWebSocketStore = ReturnType<typeof useRoomWebSocketStore>

export function useControllerRoomControls(roomId: Ref<string>, connection: RoomWebSocketStore) {
  const errorMessage = ref('')
  const networkErrorMessage = '通信できませんでした。ネットワーク接続を確認してください。'
  const isDisabled = ref(false)
  const gameStarted = ref(true)
  const pickStarted = ref(false)
  const pickExhausted = ref(true)
  const qrCodeShowed = ref(false)
  const startModal = ref(false)
  const finishModal = ref(false)
  const finishedModal = ref(false)
  const ready = ref(false)
  let disposed = false

  function applyState(state: string, pickState: string, qrCodeVisible: boolean) {
    gameStarted.value = state !== 'waiting'
    pickStarted.value = pickState === 'picking'
    pickExhausted.value = pickState === 'exhausted'
    qrCodeShowed.value = qrCodeVisible
    finishedModal.value = state === 'finished'
    ready.value = true
  }

  watch(
    () =>
      [
        connection.roomState,
        connection.pickState,
        connection.qrCodeVisible,
        connection.initializationCount,
      ] as const,
    ([state, pick, qr]) => {
      if (
        connection.roomId === roomId.value &&
        connection.mode === 'display' &&
        state &&
        pick &&
        qr !== null
      ) {
        applyState(state, pick, qr)
      }
    },
  )

  // A fresh snapshot after reconnecting means the network is back; drop a stale
  // "network down" message so the operator is not misled.
  watch(
    () => connection.initializationCount,
    () => {
      if (errorMessage.value === networkErrorMessage) errorMessage.value = ''
    },
  )

  function hasWebSocketState() {
    return (
      connection.roomId === roomId.value &&
      connection.mode === 'display' &&
      connection.status === 'open' &&
      connection.roomState !== null
    )
  }

  // fetch() rejects with a TypeError ("Failed to fetch") when the network is down.
  const describeError = (error: unknown, fallback: string) =>
    error instanceof TypeError
      ? networkErrorMessage
      : error instanceof Error
        ? error.message
        : fallback

  async function getLatestInfo(afterControl = false) {
    try {
      const { data, error, response } = await apiClient.GET('/api/rooms/{roomId}', {
        params: { path: { roomId: roomId.value } },
      })
      if (response.status === 404) connection.markRoomDeleted()
      if (error || !data) throw new Error(error?.message ?? 'ルーム情報を取得できませんでした。')
      // REST seeds state before the socket is live; a late control response must not replace newer socket state.
      if (!disposed && !hasWebSocketState() && !(afterControl && connection.status === 'open')) {
        applyState(data.state, data.pickState, data.qrCodeVisible)
      }
      return data
    } catch (error) {
      errorMessage.value = describeError(error, '通信に失敗しました。')
      return null
    }
  }

  type RoomInfo = NonNullable<Awaited<ReturnType<typeof getLatestInfo>>>

  async function control(path: ControlPath, matches: (state: RoomInfo) => boolean) {
    if (isDisabled.value || !ready.value) return
    isDisabled.value = true
    errorMessage.value = ''
    let failure = ''
    try {
      const { error } = await apiClient.POST(path, { params: { path: { roomId: roomId.value } } })
      if (error) failure = error.message
    } catch (error) {
      failure = describeError(error, '操作に失敗しました。')
    } finally {
      const state = await getLatestInfo(true)
      if (failure && (!state || !matches(state))) errorMessage.value = failure
      startModal.value = false
      finishModal.value = false
      isDisabled.value = false
    }
  }

  const clickStartModal = () =>
    control('/api/rooms/{roomId}/control/start', (state) => state.state === 'playing')
  const clickFinishModal = () =>
    control('/api/rooms/{roomId}/control/finish', (state) => state.state === 'finished')
  const clickPickStart = () =>
    control('/api/rooms/{roomId}/control/pick/start', (state) => state.pickState === 'picking')
  const clickPickFinish = () =>
    control('/api/rooms/{roomId}/control/pick/finish', (state) => state.pickState !== 'picking')
  const clickPickCancel = () =>
    control('/api/rooms/{roomId}/control/pick/cancel', (state) => state.pickState === 'idle')
  const clickShowQr = () =>
    control('/api/rooms/{roomId}/control/qrcode/show', (state) => state.qrCodeVisible)
  const clickHideQr = () =>
    control('/api/rooms/{roomId}/control/qrcode/hide', (state) => !state.qrCodeVisible)

  return {
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
    dispose() {
      disposed = true
    },
    isDisposed() {
      return disposed
    },
  }
}
