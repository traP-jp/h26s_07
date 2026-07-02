import { watch } from 'vue'
import { storeToRefs } from 'pinia'

import { getBallPalette } from '@/components/display/ballPalette'
import { useRoomWebSocketStore } from '@/stores/roomWebSocket'

type NumberBallFavicon = {
  ballColor: string
  text: string
  textColor: string
}

export function useNumberBallFavicon() {
  const roomWebSocketStore = useRoomWebSocketStore()
  const { latestBall } = storeToRefs(roomWebSocketStore)

  watch(
    latestBall,
    (pickedBall) => {
      if (pickedBall == null) {
        return
      }

      updateNumberBallFavicon({
        ballColor: getBallPalette(pickedBall).picked,
        text: String(pickedBall),
        textColor: '#ffffff',
      })
    },
    { immediate: true },
  )
}

function updateNumberBallFavicon({ ballColor, text, textColor }: NumberBallFavicon) {
  const icon = document.querySelector<HTMLLinkElement>('link[rel~="icon"]')
  if (!icon) {
    return
  }

  icon.type = 'image/svg+xml'
  icon.href = createNumberBallFaviconDataUrl({ ballColor, text, textColor })
}

function createNumberBallFaviconDataUrl({ ballColor, text, textColor }: NumberBallFavicon) {
  const fontSize = text.length >= 2 ? 26 : 32
  const svg = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 64 64">
<circle cx="32" cy="32" r="29" fill="${ballColor}"/>
<text x="32" y="33" fill="${textColor}" font-family="system-ui, sans-serif" font-size="${fontSize}"
font-weight="900" text-anchor="middle" dominant-baseline="middle">${text}
</text>
</svg>`

  return `data:image/svg+xml,${encodeURIComponent(svg)}`
}
