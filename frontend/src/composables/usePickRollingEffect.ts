import { computed, onBeforeUnmount, ref, watch } from 'vue'
import type { Ref } from 'vue'
import { storeToRefs } from 'pinia'

import type { PickState, PickedBall, RoomState } from '@/api/schema'
import { getBallPalette } from '@/components/display/ballPalette'
import { useAudioUnlock, useSoundEffect } from '@/composables/useSoundEffect'
import { useRoomWebSocketStore } from '@/stores/roomWebSocket'

function useRollingBallAnimation() {
  const rollingPickedBall = ref<PickedBall | null>(null)
  let rollingTimerId: number | null = null

  function stopRollingBall(): void {
    if (rollingTimerId == null) return

    window.clearInterval(rollingTimerId)
    rollingTimerId = null
  }

  function startRollingBall(): void {
    stopRollingBall()
    rollingTimerId = window.setInterval(() => {
      rollingPickedBall.value = Math.floor(Math.random() * 75 + 1) as PickedBall
    }, 60)
  }

  return { rollingPickedBall, startRollingBall, stopRollingBall }
}

function useBallDisplay(
  roomState: Readonly<Ref<RoomState | null>>,
  pickState: Readonly<Ref<PickState | null>>,
  rollingPickedBall: Readonly<Ref<PickedBall | null>>,
  latestBall: Readonly<Ref<PickedBall | null>>,
) {
  const displayBall = computed(() => {
    if (roomState.value == null || roomState.value === 'waiting') return null
    if (pickState.value === 'picking') return rollingPickedBall.value ?? latestBall.value
    return latestBall.value
  })

  const displayBallText = computed(() =>
    displayBall.value == null ? '?' : String(displayBall.value),
  )

  const displayBallPalette = computed(() => {
    const palette = getBallPalette(displayBall.value)

    if (displayBall.value == null) return palette

    // 選ばれたボールは文字色を白にする
    return {
      ...palette,
      text: '#ffffff',
    }
  })

  return { displayBallPalette, displayBallText }
}

export function usePickRollingEffect() {
  const roomWebSocketStore = useRoomWebSocketStore()
  const { latestBall, latestEvent, pickState, roomState, status } = storeToRefs(roomWebSocketStore)
  const { rollingPickedBall, startRollingBall, stopRollingBall } = useRollingBallAnimation()
  const { displayBallPalette, displayBallText } = useBallDisplay(
    roomState,
    pickState,
    rollingPickedBall,
    latestBall,
  )
  const drumroll = useSoundEffect('drumroll', { loop: true })
  const cymbal = useSoundEffect('cymbal')
  const { audioUnlocked, unlockAudio } = useAudioUnlock()

  watch(
    [pickState, status, audioUnlocked],
    ([pick, connectionStatus]) => {
      if (pick === 'picking' && connectionStatus === 'open') {
        startRollingBall()
        drumroll.play()
      } else {
        stopRollingBall()
        drumroll.stop()
      }
    },
    { immediate: true },
  )

  watch(latestEvent, (event) => {
    if (event?.type === 'PickFinished') {
      drumroll.stop()
      cymbal.play()
    }
    if (event && ['PickCanceled', 'GameFinished', 'AllPicked'].includes(event.type)) {
      stopRollingBall()
      drumroll.stop()
    }
  })

  onBeforeUnmount(() => {
    stopRollingBall()
    drumroll.stop()
  })

  return {
    audioUnlocked,
    unlockAudio,
    displayBallPalette,
    displayBallText,
    pickState,
  }
}
