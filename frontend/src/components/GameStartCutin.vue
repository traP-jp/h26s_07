<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, type CSSProperties } from 'vue'

type TextAnimation = {
  play: () => void
  pause?: () => void
  cancel?: () => void
}

type TextSplitter = {
  words: Element[]
  revert?: () => void
}

type Props = {
  topText?: string
  bottomText?: string
  topColor?: string
  bottomColor?: string
  textColor?: string
  duration?: number
  easing?: string
  textStartDelay?: number
  textHoldDelay?: number
  textDuration?: number
  textStagger?: number
  fadeDuration?: number
  rotateFrom?: number
  panelOpacity?: number
}

const props = withDefaults(defineProps<Props>(), {
  topText: 'GAME',
  bottomText: 'START',
  topColor: '#000000',
  bottomColor: '#000000',
  textColor: '#ffffff',
  duration: 2000,
  easing: 'cubic-bezier(0.4, 0, 0.2, 1)',
  textStartDelay: 1150,
  textHoldDelay: 1100,
  textDuration: 650,
  textStagger: 180,
  fadeDuration: 500,
  rotateFrom: -30,
  panelOpacity: 0.8,
})
const emit = defineEmits<{
  complete: []
}>()

const root = ref<HTMLElement | null>(null)
let textAnimation: TextAnimation | undefined
let textSplitters: TextSplitter[] = []
let textTimer: number | undefined
let completeTimer: number | undefined

const stageStyle = computed(
  () =>
    ({
      '--motion-duration': `${props.duration}ms`,
      '--motion-easing': props.easing,
      '--transition__duration': 'var(--motion-duration)',
      '--transition__easing': 'var(--motion-easing)',
      '--top-color': props.topColor,
      '--bottom-color': props.bottomColor,
      '--text-color': props.textColor,
      '--fade-duration': `${props.fadeDuration}ms`,
      '--rotate-from': `${props.rotateFrom}deg`,
      '--panel-opacity': String(props.panelOpacity),
    }) as CSSProperties,
)

onMounted(async () => {
  const { animate, splitText, stagger } = await import('animejs')

  await nextTick()

  if (!root.value) return

  const labels = Array.from(root.value.querySelectorAll<HTMLElement>('.wipe-label'))

  textSplitters = labels.map((label) =>
    splitText(label, {
      words: { wrap: 'clip' },
      accessible: false,
    }),
  )

  const words = textSplitters.flatMap((splitter) => splitter.words)

  textAnimation = animate(words, {
    y: [{ to: ['-100%', '0%'] }, { to: '100%', delay: props.textHoldDelay, ease: 'in(3)' }],
    duration: props.textDuration,
    ease: 'out(3)',
    delay: stagger(props.textStagger),
    loop: false,
    autoplay: false,
    onComplete: () => {
      root.value?.querySelector('.tilt')?.classList.add('is-fading')
      completeTimer = window.setTimeout(() => {
        emit('complete')
      }, props.fadeDuration)
    },
  }) as TextAnimation

  textTimer = window.setTimeout(() => {
    labels.forEach((label) => label.classList.add('is-visible'))
    textAnimation?.play()
  }, props.textStartDelay)
})

onBeforeUnmount(() => {
  window.clearTimeout(textTimer)
  window.clearTimeout(completeTimer)
  textAnimation?.cancel?.()
  textAnimation?.pause?.()
  textSplitters.forEach((splitter) => splitter.revert?.())
})
</script>

<template>
  <main ref="root" class="cutin-stage" :style="stageStyle">
    <div class="tilt">
      <div class="wipe wipe-red" transition-style="in:wipe:left">
        <span class="wipe-label label-red">{{ topText }}</span>
      </div>
      <div class="wipe wipe-blue" transition-style="in:wipe:right">
        <span class="wipe-label label-blue">{{ bottomText }}</span>
      </div>
    </div>
  </main>
</template>

<style scoped>
.cutin-stage {
  position: fixed;
  inset: 0;
  overflow: hidden;
}

.tilt {
  position: absolute;
  top: 50%;
  left: 50%;
  width: 180vmax;
  height: 180vmax;
  animation: rotate-in var(--motion-duration) var(--motion-easing) both;
  opacity: var(--panel-opacity);
  transform-origin: center;
  transition: opacity var(--fade-duration) ease;
  will-change: transform;
}

.tilt.is-fading {
  opacity: 0;
}

.wipe {
  position: absolute;
  left: 0;
  width: 100%;
  height: calc(50% + 4px);
}

.wipe-red {
  top: -2px;
  background: var(--top-color);
}

.wipe-blue {
  bottom: -2px;
  background: var(--bottom-color);
}

.wipe-label {
  position: absolute;
  left: 50%;
  display: block;
  opacity: 0;
  color: var(--text-color);
  font-family:
    ui-sans-serif,
    system-ui,
    -apple-system,
    BlinkMacSystemFont,
    'Segoe UI',
    sans-serif;
  font-size: clamp(72px, 18vw, 220px);
  font-weight: 900;
  letter-spacing: 0;
  line-height: 1;
  transform: translateX(-50%);
  transition: opacity 180ms ease;
  white-space: nowrap;
}

.wipe-label.is-visible {
  opacity: 1;
}

.label-red {
  bottom: 3vmin;
}

.label-blue {
  top: 3vmin;
}

@keyframes rotate-in {
  0% {
    transform: translate(calc(-50% + 20vmin), calc(-50% - 20vmin)) rotate(var(--rotate-from));
  }

  60%,
  100% {
    transform: translate(-50%, -50%) rotate(0deg);
  }
}
</style>
