<script setup lang="ts">
import { computed, ref, watch } from 'vue'

import type { CardCellState, PickedBall } from '@/api/schema'
import NumberBall from '@/components/layouts/NumberBall.vue'
import { getBallPalette } from '@/components/rooms/ballPalette'

const props = defineProps<{
  pickedBalls: PickedBall[]
  latestPickedBall?: PickedBall | null
}>()

type BallCell = {
  number: PickedBall
  state: Extract<CardCellState, 'open' | 'closed'>
  isLatest: boolean
  ballColor: string
  ringColor: string
  textColor: string
}

const seenBallNumbers = ref<Set<PickedBall>>(new Set())

watch(
  () => [props.pickedBalls, props.latestPickedBall] as const,
  ([pickedBalls, latestPickedBall]) => {
    if (pickedBalls.length === 0 && latestPickedBall == null) {
      seenBallNumbers.value = new Set()
      return
    }

    const nextSeenBallNumbers = new Set(seenBallNumbers.value)
    pickedBalls.forEach((ball) => nextSeenBallNumbers.add(ball))
    if (latestPickedBall != null) {
      nextSeenBallNumbers.add(latestPickedBall)
    }
    seenBallNumbers.value = nextSeenBallNumbers
  },
  { immediate: true },
)

const balls = computed<BallCell[]>(() =>
  Array.from({ length: 75 }, (_, index) => {
    const number = (index + 1) as PickedBall
    const palette = getBallPalette(number)
    const isPicked = seenBallNumbers.value.has(number)

    return {
      number,
      state: isPicked ? 'open' : 'closed',
      isLatest: props.latestPickedBall === number,
      ballColor: isPicked ? palette.picked : palette.waiting,
      ringColor: palette.ring,
      textColor: isPicked ? '#ffffff' : palette.text,
    }
  }),
)
</script>

<template>
  <div class="ball-state-grid-frame flex justify-center">
    <div class="ball-state-grid grid">
      <div
        v-for="ball in balls"
        :key="ball.number"
        class="ball-state-grid__cell relative grid place-items-center"
        :class="{ 'opacity-70': ball.state === 'closed' }"
      >
        <span
          v-if="ball.isLatest"
          class="pointer-events-none absolute -inset-1 rounded-full border-3"
          :style="{ borderColor: ball.ringColor }"
          aria-hidden="true"
        />
        <NumberBall
          class="ball-state-grid__number"
          :ball-color="ball.ballColor"
          :text-color="ball.textColor"
          :text="String(ball.number)"
          :size="42"
        />
      </div>
    </div>
  </div>
</template>

<style scoped>
.ball-state-grid-frame {
  box-sizing: border-box;
  width: 100%;
  max-width: 742px;
}

.ball-state-grid {
  width: fit-content;
  grid-template-columns: repeat(15, 42px);
  gap: 8px;
}

.ball-state-grid__cell {
  width: 42px;
  height: 42px;
  aspect-ratio: 1 / 1;
}

@media (max-width: 860px) {
  .ball-state-grid {
    width: 100%;
    grid-template-columns: repeat(15, minmax(18px, 1fr));
    gap: 4px;
  }

  .ball-state-grid__cell,
  .ball-state-grid__number {
    width: 100% !important;
    height: auto !important;
    min-width: 18px;
    min-height: 18px;
    aspect-ratio: 1 / 1;
  }

  .ball-state-grid__number {
    font-size: 2.2vw !important;
  }
}
</style>
