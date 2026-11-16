import { nextTick, ref, watch } from 'vue'
import { storeToRefs } from 'pinia'

import { useRoomWebSocketStore } from '@/stores/roomWebSocket'

export function useDisplayEffects() {
  const roomWebSocketStore = useRoomWebSocketStore()
  const { bingoParticipantCount, latestEvent } = storeToRefs(roomWebSocketStore)

  const showCutin = ref(false)
  const cutinKey = ref(0)
  const topText = ref('GAME')
  const bottomText = ref('START')

  async function playCutin() {
    showCutin.value = false
    await nextTick()

    cutinKey.value += 1
    showCutin.value = true
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

  watch(bingoParticipantCount, (count, previousCount = 0) => {
    if (latestEvent.value?.type === 'Initialized') {
      return
    }

    const increasedBingoCount = count - previousCount

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
    showCutin,
    topText,
  }
}
