import { nextTick, ref, watch } from 'vue'
import { storeToRefs } from 'pinia'

import { useRoomWebSocketStore } from '@/stores/roomWebSocket'

export function useDisplayEffects() {
  const roomWebSocketStore = useRoomWebSocketStore()
  const { bingoSummaries, latestEvent } = storeToRefs(roomWebSocketStore)

  const showCutin = ref(false)
  const cutinKey = ref(0)
  const topText = ref('GAME')
  const bottomText = ref('START')
  let cutinGeneration = 0

  async function playCutin() {
    const generation = ++cutinGeneration
    showCutin.value = false
    await nextTick()
    if (generation !== cutinGeneration) return

    cutinKey.value += 1
    showCutin.value = true
  }

  function resetDisplayEffects() {
    cutinGeneration += 1
    showCutin.value = false
    topText.value = 'GAME'
    bottomText.value = 'START'
  }

  function handleCutinComplete() {
    showCutin.value = false
  }

  watch(latestEvent, (event) => {
    if (event == null) {
      return
    }

    if (event.type === 'GameStarted') {
      topText.value = 'GAME'
      bottomText.value = 'START'
      void playCutin()
    }
  })

  watch(bingoSummaries, (summaries = [], previousSummaries = []) => {
    if (latestEvent.value?.type === 'Initialized') {
      return
    }

    const increasedBingoCount = summaries.length - previousSummaries.length

    if (increasedBingoCount < 1) {
      return
    }

    topText.value = String(increasedBingoCount) + ' players'
    bottomText.value = 'BINGO!!!'
    void playCutin()
  })

  return {
    bottomText,
    cutinKey,
    handleCutinComplete,
    resetDisplayEffects,
    showCutin,
    topText,
  }
}
