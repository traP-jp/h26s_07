import { onBeforeUnmount, onMounted, ref } from 'vue'

import cymbalUrl from '@/assets/sounds/cymbal.mp3'
import drumrollUrl from '@/assets/sounds/drumroll.mp3'

const soundEffectUrls = {
  cymbal: cymbalUrl,
  drumroll: drumrollUrl,
} as const

const audioUnlocked = ref(false)
const sounds = new Set<HTMLAudioElement>()
let unlocking = false

export function useAudioUnlock() {
  async function unlockAudio() {
    if (audioUnlocked.value || unlocking) return
    unlocking = true
    try {
      await Promise.all(
        [...sounds].map(async (audio) => {
          audio.muted = true
          try {
            await audio.play()
            audio.pause()
            audio.currentTime = 0
          } finally {
            audio.muted = false
          }
        }),
      )
      audioUnlocked.value = sounds.size > 0
    } catch {
      // Leave the unlock button available for another gesture.
    } finally {
      unlocking = false
    }
  }
  onMounted(() => {
    window.addEventListener('pointerdown', unlockAudio)
    window.addEventListener('keydown', unlockAudio)
  })
  onBeforeUnmount(() => {
    window.removeEventListener('pointerdown', unlockAudio)
    window.removeEventListener('keydown', unlockAudio)
  })
  return { audioUnlocked, unlockAudio }
}

export type SoundEffectName = keyof typeof soundEffectUrls

type UseSoundEffectOptions = {
  loop?: boolean
}

export function useSoundEffect(name: SoundEffectName, options: UseSoundEffectOptions = {}) {
  let audio: HTMLAudioElement | null = null

  function play() {
    if (!audio) return
    // A looping sound that is already playing keeps going instead of restarting.
    if (audio.loop && !audio.paused) return

    audio.currentTime = 0
    // Try even before an unlock gesture: kiosk browsers that allow autoplay
    // should not need a click. Only an autoplay refusal re-shows the button;
    // AbortError from a quick stop() must not.
    void audio
      .play()
      .then(() => {
        audioUnlocked.value = true
      })
      .catch((error: unknown) => {
        if (error instanceof DOMException && error.name === 'NotAllowedError') {
          audioUnlocked.value = false
        }
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
    // First play primes this element lazily, preserving the page's unlock state.
    audio.preload = 'none'
    sounds.add(audio)
  })

  onBeforeUnmount(() => {
    stop()
    if (audio) sounds.delete(audio)
    audio = null
  })

  return {
    play,
    stop,
  }
}
