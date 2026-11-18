import { computed, onBeforeUnmount, ref, watch } from 'vue'
import { storeToRefs } from 'pinia'

import type { PickedBall } from '@/api/schema'
import { getBallPalette } from '@/components/display/ballPalette'
import { useSoundEffect } from '@/composables/useSoundEffect'
import { useRoomWebSocketStore } from '@/stores/roomWebSocket'

function createRandomPickedBall() {
  return Math.floor(Math.random() * 75 + 1) as PickedBall
}

export function usePickRollingEffect() {
  const roomWebSocketStore = useRoomWebSocketStore()
  const { latestBall, latestEvent, pickState, roomState } = storeToRefs(roomWebSocketStore)
  const rollingPickedBall = ref<PickedBall | null>(null)
  const drumroll = useSoundEffect('drumroll', { loop: true })
  const cymbal = useSoundEffect('cymbal')
  let rollingTimerId: number | null = null

  const displayBall = computed(() => {
    if (roomState.value == null || roomState.value === 'waiting') {
      return null
    }

    if (pickState.value === 'picking') {
      return rollingPickedBall.value ?? latestBall.value
    }

    return latestBall.value
  })

  const displayBallText = computed(() => {
    if (displayBall.value == null) {
      return '?'
    }

    return String(displayBall.value)
  })

  const displayBallPalette = computed(() => {
    const palette = getBallPalette(displayBall.value)

    if (displayBall.value == null) {
      return palette
    }

    // 選ばれたボールは文字色を白にする
    return {
      ...palette,
      text: '#ffffff',
    }
  })

  function stopRollingBall() {
    if (rollingTimerId == null) return

    window.clearInterval(rollingTimerId)
    rollingTimerId = null
  }

  function startRollingBall() {
    stopRollingBall()

    rollingTimerId = window.setInterval(() => {
      rollingPickedBall.value = createRandomPickedBall()
    }, 60)
  }

  watch(
    pickState,
    (state) => {
      stopRollingBall()
      if (state === 'picking') {
        startRollingBall()
        drumroll.play()
      } else {
        rollingPickedBall.value = null
        drumroll.stop()
      }
    },
    { immediate: true },
  )

  watch(latestEvent, (event) => {
    if (event?.type === 'PickFinished') {
      cymbal.play()
    }
  })

  onBeforeUnmount(() => {
    stopRollingBall()
    drumroll.stop()
  })

  return {
    displayBallPalette,
    displayBallText,
    pickState,
  }
}
