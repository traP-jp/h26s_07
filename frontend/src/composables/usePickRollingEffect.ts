import { computed, onBeforeUnmount, ref, watch } from 'vue'
import { useMediaQuery } from '@vueuse/core'
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
  const prefersReducedMotion = useMediaQuery('(prefers-reduced-motion: reduce)')
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
    [pickState, prefersReducedMotion],
    ([state, reduceMotion]) => {
      stopRollingBall()
      if (state === 'picking') {
        if (reduceMotion) {
          rollingPickedBall.value = null
        } else {
          startRollingBall()
        }
      } else {
        rollingPickedBall.value = null
      }
    },
    { immediate: true },
  )

  watch(
    pickState,
    (state) => {
      if (state === 'picking') {
        drumroll.play()
      } else {
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
