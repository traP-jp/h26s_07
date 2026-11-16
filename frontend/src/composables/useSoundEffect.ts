import { onBeforeUnmount, onMounted } from 'vue'

import cymbalUrl from '@/assets/sounds/cymbal.mp3'
import drumrollUrl from '@/assets/sounds/drumroll.mp3'

const soundEffectUrls = {
  cymbal: cymbalUrl,
  drumroll: drumrollUrl,
} as const

export type SoundEffectName = keyof typeof soundEffectUrls

type UseSoundEffectOptions = {
  loop?: boolean
}

export function useSoundEffect(name: SoundEffectName, options: UseSoundEffectOptions = {}) {
  let audio: HTMLAudioElement | null = null

  function play() {
    if (!audio) return

    audio.currentTime = 0
    void audio.play().catch(() => {
      // ページ上でユーザー操作が行われるまでは、ブラウザーの自動再生制限で再生できない場合がある。
    })
  }

  function stop() {
    if (!audio) return

    audio.pause()
    audio.currentTime = 0
  }

  onMounted(() => {
    audio = new Audio(soundEffectUrls[name])
    audio.loop = options.loop ?? false
  })

  onBeforeUnmount(() => {
    stop()
    audio = null
  })

  return {
    play,
    stop,
  }
}
