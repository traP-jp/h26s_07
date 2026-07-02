<script setup lang="ts">
import type { Card, CardCell as CardCellType, CardChanges, Line } from '@/api/schema'
import CardCell from '@/components/bingo/CardCell.vue'
import { computed, ref, watch } from 'vue'

const props = withDefaults(
  defineProps<{
    card?: Card | null
    cardChanges?: CardChanges | null
    placeholder?: boolean
  }>(),
  { cardChanges: null, placeholder: false },
)

type CardEffectMode = 'bingo' | 'reach'
type LineEffectItem = {
  id: string
  mode: CardEffectMode
  line: Line
  style: Record<string, string>
}

const cardNo = computed(() => props.card?.cardNumber ?? 'not assigned')
const title = ['B', 'I', 'N', 'G', 'O']
const placeholderCells: CardCellType[] = Array.from({ length: 25 }, (_, index) => ({
  index,
  number: index === 12 ? null : 0,
  displayText: index === 12 ? 'FREE' : '?',
  cellState: index === 12 ? 'open' : 'closed',
}))
const cells = computed(() => props.card?.cells ?? placeholderCells)
const effectSerial = ref(0)

const activeEffect = computed<{ mode: CardEffectMode; line: Line } | null>(() => {
  const bingoLine = props.cardChanges?.newBingoLines.at(-1)
  if (bingoLine) return { mode: 'bingo', line: bingoLine }

  const reachLine = props.cardChanges?.newReachLines.at(-1)
  if (reachLine) return { mode: 'reach', line: reachLine }

  return null
})

const activeEffectLine = computed(() => activeEffect.value?.line ?? [])
const activeEffectKey = computed(() => {
  const effect = activeEffect.value
  return effect ? `${effect.mode}-${effect.line.join('-')}-${effectSerial.value}` : 'none'
})

const persistentLineEffects = computed<LineEffectItem[]>(() => {
  const bingoLines = props.card?.bingoLines ?? []
  const bingoLineKeys = new Set(bingoLines.map(lineKey))
  const reachLines = (props.card?.reachLines ?? []).filter(
    (line) => !bingoLineKeys.has(lineKey(line)),
  )

  return [
    ...bingoLines.map((line, index) => ({
      id: `bingo-${index}-${lineKey(line)}`,
      mode: 'bingo' as const,
      line,
      style: createLineEffectStyle(line),
    })),
    ...reachLines.map((line, index) => ({
      id: `reach-${index}-${lineKey(line)}`,
      mode: 'reach' as const,
      line,
      style: createLineEffectStyle(line),
    })),
  ]
})

function createLineEffectStyle(line: Line): Record<string, string> {
  if (!line || line.length < 2) return {}

  const start = cellCenter(line[0]!)
  const end = cellCenter(line[line.length - 1]!)
  const dx = end.x - start.x
  const dy = end.y - start.y
  const length = Math.hypot(dx, dy)
  const angle = Math.atan2(dy, dx)

  return {
    '--line-x': `${start.x}%`,
    '--line-y': `${start.y}%`,
    '--line-length': `${length}%`,
    '--line-angle': `${angle}rad`,
  }
}

function lineKey(line: Line) {
  return line.join('-')
}

const celebrationParticles = Array.from({ length: 18 }, (_, index) => ({
  id: index,
  style: {
    '--particle-x': `${Math.cos((index / 18) * Math.PI * 2) * (42 + (index % 3) * 8)}px`,
    '--particle-y': `${Math.sin((index / 18) * Math.PI * 2) * (30 + (index % 4) * 7)}px`,
    '--particle-delay': `${index * 26}ms`,
  },
}))

const isInLine = (line: Line[] | undefined, cellIndex: number) => {
  return line?.some((candidate) => candidate.includes(cellIndex)) ?? false
}

const isInActiveEffectLine = (cellIndex: number) => activeEffectLine.value.includes(cellIndex)

const isReachMissingCell = (cell: CardCellType) => {
  const effect = activeEffect.value
  return effect?.mode === 'reach' && effect.line.includes(cell.index) && cell.cellState === 'closed'
}

const activeLineOrder = (cellIndex: number) => {
  const index = activeEffectLine.value.indexOf(cellIndex)
  return index === -1 ? 0 : index
}

const gridCellClass = (cell: CardCellType) => ({
  'grid-cell--line-bingo': isInLine(props.card?.bingoLines, cell.index),
  'grid-cell--line-reach': isInLine(props.card?.reachLines, cell.index),
  'grid-cell--active-line': isInActiveEffectLine(cell.index),
  'grid-cell--active-bingo':
    activeEffect.value?.mode === 'bingo' && isInActiveEffectLine(cell.index),
  'grid-cell--active-reach':
    activeEffect.value?.mode === 'reach' && isInActiveEffectLine(cell.index),
  'grid-cell--reach-missing': isReachMissingCell(cell),
})

const gridCellStyle = (cell: CardCellType) => ({
  '--shine-delay': `${activeLineOrder(cell.index) * 110}ms`,
})

function cellCenter(index: number) {
  const column = Math.floor(index / 5)
  const row = index % 5

  return {
    x: ((column + 0.5) / 5) * 100,
    y: ((row + 0.5) / 5) * 100,
  }
}

watch(
  () => props.cardChanges,
  () => {
    effectSerial.value += 1
  },
)
</script>

<template>
  <div
    class="bingo-paper"
    :class="{ 'bingo-paper--placeholder': props.placeholder || !props.card }"
  >
    <div class="bingo-grid">
      <div class="bingo-title">
        <div v-for="letter in title" :key="letter" class="bingo-title-cell">
          {{ letter }}
        </div>
      </div>

      <div class="bingo-grid__cells">
        <div
          v-for="cell in cells"
          :key="cell.index"
          class="grid-cell"
          :class="gridCellClass(cell)"
          :style="gridCellStyle(cell)"
        >
          <CardCell :cell="cell" :style="{ width: '80%', height: '80%' }" />
        </div>

        <div
          v-for="lineEffect in persistentLineEffects"
          :key="`${lineEffect.id}-${effectSerial}`"
          class="line-effect"
          :class="`line-effect--${lineEffect.mode}`"
          :style="lineEffect.style"
        ></div>
      </div>
    </div>

    <div
      v-if="activeEffect?.mode === 'bingo'"
      :key="`celebration-${activeEffectKey}`"
      class="bingo-celebration"
    >
      <span
        v-for="particle in celebrationParticles"
        :key="particle.id"
        class="bingo-celebration__particle"
        :style="particle.style"
      ></span>
      <span class="bingo-celebration__text">BINGO!!</span>
    </div>

    <div class="bingo-footer">
      <span class="card-no">Card No. {{ cardNo }}</span>
    </div>

    <div v-if="props.placeholder || !props.card" class="bingo-placeholder-message">
      カードはまだ配られていません
    </div>
  </div>
</template>

<style scoped>
.bingo-paper {
  position: relative;
  box-sizing: border-box;
  --cell-size: calc(100cqw / 5.8);
  padding: 3% 2.5% 1%;
  background: #c32020;
  box-shadow: 0 4px 14px rgb(0 0 0 / 18%);
  container-type: inline-size;
}

.bingo-paper--placeholder .grid-cell {
  background: #f3f5f8;
}

.bingo-paper--placeholder :deep(.number-ball) {
  opacity: 0.35;
}

.bingo-grid {
  display: grid;
  gap: 0.5em;
}

.bingo-title {
  display: grid;
  user-select: none;
  grid-template-columns: repeat(5, 1fr);
}

.bingo-title-cell {
  text-align: center;
  color: #ffffff;
  font-weight: 900;
  letter-spacing: 0.05em;
  line-height: 1;
  font-size: calc(var(--cell-size) * 0.7);
}

.bingo-grid__cells {
  position: relative;
  display: grid;
  aspect-ratio: 1 / 1;
  grid-template-columns: repeat(5, 1fr);
  grid-template-rows: repeat(5, 1fr);
  grid-auto-flow: column;
  border-top: 4px solid #c32020;
  border-left: 4px solid #c32020;
  background: #c32020;
}

.grid-cell {
  position: relative;
  z-index: 1;
  background: #ffffff;
  border-right: 4px solid #c32020;
  border-bottom: 4px solid #c32020;

  display: flex;
  align-items: center;
  justify-content: center;
}

.grid-cell--line-reach {
  background: #fff8fb;
}

.grid-cell--line-bingo {
  background: #fff9fb;
}

.grid-cell--active-line::before {
  position: absolute;
  inset: 5%;
  z-index: 0;
  border-radius: 16px;
  opacity: 0;
  pointer-events: none;
  content: '';
}

.grid-cell--active-reach::before {
  background: radial-gradient(circle, rgb(255 211 235 / 0.76), rgb(255 244 185 / 0.24) 64%);
  animation: reach-cell-glow 1250ms ease-out var(--shine-delay) both;
}

.grid-cell--active-bingo::before {
  background: radial-gradient(circle, rgb(255 216 236 / 0.95), rgb(255 239 183 / 0.34) 68%);
  animation: bingo-cell-spark 1450ms ease-out var(--shine-delay) both;
}

.grid-cell--active-line :deep(.ball) {
  position: relative;
  z-index: 1;
}

.grid-cell--reach-missing :deep(.ball) {
  animation: reach-ball-wobble 1500ms ease-in-out infinite;
}

.grid-cell--line-reach:not(.grid-cell--line-bingo) :deep(.ball) {
  box-shadow:
    0 3px 9px rgb(0 0 0 / 0.08),
    0 0 0 calc(var(--cell-size) * 0.025) rgb(255 193 224 / 0.18),
    0 0 12px rgb(255 185 105 / 0.22);
  animation: reach-ball-aura 1700ms ease-in-out infinite;
}

.grid-cell--line-bingo :deep(.ball) {
  animation: bingo-ball-shimmer 1800ms ease-in-out infinite;
}

.line-effect {
  position: absolute;
  top: var(--line-y);
  left: var(--line-x);
  z-index: 2;
  width: var(--line-length);
  height: calc(var(--cell-size) * 0.13);
  border-radius: 999px;
  opacity: 0;
  pointer-events: none;
  transform: translateY(-50%) rotate(var(--line-angle));
  transform-origin: 0 50%;
}

.line-effect::before {
  position: absolute;
  inset: 0;
  border-radius: inherit;
  content: '';
  transform: scaleX(0);
  transform-origin: 0 50%;
}

.line-effect::after {
  position: absolute;
  top: 50%;
  left: 0;
  width: calc(var(--cell-size) * 0.2);
  height: calc(var(--cell-size) * 0.2);
  border-radius: 999px;
  content: '';
  transform: translate(-50%, -50%);
}

.line-effect--reach {
  filter: drop-shadow(0 0 10px rgb(255 181 92 / 0.58));
  animation: line-effect-hold 1700ms ease-in-out infinite;
}

.line-effect--reach::before {
  background: linear-gradient(90deg, rgb(255 205 232 / 0), #ffc1dc 34%, #fff0a8 68%);
  animation: line-sweep-loop 1700ms ease-in-out infinite;
}

.line-effect--reach::after {
  background: radial-gradient(circle, #ffffff 0 22%, #fff0a8 42%, rgb(255 193 220 / 0) 72%);
  animation: line-comet-loop 1700ms ease-in-out infinite;
}

.line-effect--bingo {
  filter: drop-shadow(0 0 14px rgb(255 192 226 / 0.72));
  animation: line-effect-hold 1550ms ease-in-out infinite;
}

.line-effect--bingo::before {
  background: linear-gradient(90deg, rgb(255 206 235 / 0), #ffcde8 28%, #fff0a8 62%, #c9f7ff);
  animation: line-sweep-loop 1550ms cubic-bezier(0.2, 0.8, 0.2, 1) infinite;
}

.line-effect--bingo::after {
  background: radial-gradient(
    circle,
    #ffffff 0 18%,
    #fff0a8 36%,
    #ffcde8 56%,
    rgb(255 205 232 / 0) 76%
  );
  animation: line-comet-loop 1550ms cubic-bezier(0.2, 0.8, 0.2, 1) infinite;
}

.bingo-celebration {
  position: absolute;
  top: 50%;
  left: 50%;
  z-index: 4;
  display: grid;
  width: min(82%, calc(var(--cell-size) * 4.7));
  aspect-ratio: 1 / 0.58;
  place-items: center;
  pointer-events: none;
  transform: translate(-50%, -50%);
  animation: celebration-pop 1800ms ease-out both;
}

.bingo-celebration__text {
  position: relative;
  z-index: 2;
  padding: calc(var(--cell-size) * 0.12) calc(var(--cell-size) * 0.22);
  border-radius: calc(var(--cell-size) * 0.2);
  background: rgb(255 255 255 / 0.84);
  box-shadow: 0 10px 28px rgb(255 154 209 / 0.24);
  color: #f06fb1;
  font-size: clamp(24px, calc(var(--cell-size) * 0.46), 54px);
  font-weight: 950;
  line-height: 1;
  text-shadow:
    0 2px 0 #ffffff,
    0 0 18px rgb(255 214 239 / 0.9);
  animation: bingo-text-bounce 1500ms ease-out both;
}

.bingo-celebration__particle {
  position: absolute;
  top: 50%;
  left: 50%;
  z-index: 1;
  width: calc(var(--cell-size) * 0.12);
  height: calc(var(--cell-size) * 0.12);
  border-radius: 999px;
  background: #ffcde8;
  opacity: 0;
  transform: translate(-50%, -50%);
  animation: pastel-firework 980ms ease-out var(--particle-delay) both;
}

.bingo-celebration__particle:nth-child(3n + 1) {
  background: #b8e8ff;
}

.bingo-celebration__particle:nth-child(3n + 2) {
  background: #d7fff2;
}

.bingo-celebration__particle:nth-child(3n + 3) {
  background: #fff0a8;
}

.bingo-footer {
  margin-top: 8px;
  color: #ffffff;
  font-style: italic;
}

.card-no {
  display: block;
  overflow-wrap: anywhere;
  line-height: 1.2;
  font-size: calc(var(--cell-size) * 0.18);
  text-align: right;
}

.bingo-placeholder-message {
  position: absolute;
  top: 50%;
  left: 50%;
  width: 82%;
  padding: 10px 12px;
  transform: translate(-50%, -50%);
  border-radius: 10px;
  background: rgb(255 255 255 / 0.92);
  box-shadow: 0 10px 28px rgb(0 0 0 / 0.16);
  color: #1f3556;
  font-size: clamp(14px, calc(var(--cell-size) * 0.25), 20px);
  font-weight: 800;
  line-height: 1.4;
  text-align: center;
}

@keyframes reach-cell-glow {
  0% {
    opacity: 0;
    transform: scale(0.86);
  }
  42% {
    opacity: 0.82;
    transform: scale(1.02);
  }
  100% {
    opacity: 0.16;
    transform: scale(1);
  }
}

@keyframes bingo-cell-spark {
  0% {
    opacity: 0;
    transform: scale(0.82);
  }
  38% {
    opacity: 1;
    transform: scale(1.08);
  }
  100% {
    opacity: 0.28;
    transform: scale(1);
  }
}

@keyframes reach-ball-aura {
  0%,
  100% {
    box-shadow:
      0 3px 9px rgb(0 0 0 / 0.08),
      0 0 0 calc(var(--cell-size) * 0.02) rgb(255 193 224 / 0.14),
      0 0 10px rgb(255 185 105 / 0.18);
    transform: scale(1);
  }
  45% {
    box-shadow:
      0 3px 9px rgb(0 0 0 / 0.08),
      0 0 0 calc(var(--cell-size) * 0.055) rgb(255 193 224 / 0.22),
      0 0 16px rgb(255 185 105 / 0.26);
    transform: scale(1.025);
  }
}

@keyframes reach-ball-wobble {
  0%,
  100% {
    box-shadow:
      0 3px 9px rgb(0 0 0 / 0.08),
      0 0 0 calc(var(--cell-size) * 0.025) rgb(255 193 224 / 0.16);
    transform: scale(1) rotate(0);
  }
  35% {
    box-shadow:
      0 3px 9px rgb(0 0 0 / 0.08),
      0 0 0 calc(var(--cell-size) * 0.065) rgb(255 193 224 / 0.24),
      0 0 18px rgb(255 185 105 / 0.28);
    transform: scale(1.04) rotate(-1.5deg);
  }
  65% {
    transform: scale(1.02) rotate(1.5deg);
  }
}

@keyframes bingo-ball-shimmer {
  0%,
  100% {
    box-shadow:
      0 3px 9px rgb(0 0 0 / 0.08),
      0 0 0 calc(var(--cell-size) * 0.02) rgb(255 240 168 / 0.16),
      0 0 10px rgb(255 122 182 / 0.18);
  }
  48% {
    box-shadow:
      0 3px 9px rgb(0 0 0 / 0.08),
      0 0 0 calc(var(--cell-size) * 0.045) rgb(255 240 168 / 0.24),
      0 0 16px rgb(255 122 182 / 0.26);
  }
}

@keyframes line-effect-hold {
  0%,
  100% {
    opacity: 0.4;
  }
  18%,
  72% {
    opacity: 1;
  }
}

@keyframes line-sweep-loop {
  0% {
    transform: scaleX(0);
  }
  58%,
  100% {
    transform: scaleX(1);
  }
}

@keyframes line-comet-loop {
  0% {
    left: 0;
    opacity: 0;
  }
  12% {
    opacity: 0.45;
  }
  62% {
    left: 100%;
    opacity: 0.45;
  }
  100% {
    left: 100%;
    opacity: 0;
  }
}
@keyframes celebration-pop {
  0% {
    opacity: 0;
    transform: translate(-50%, -50%) scale(0.88);
  }
  12%,
  76% {
    opacity: 1;
  }
  100% {
    opacity: 0;
    transform: translate(-50%, -50%) scale(1.02);
  }
}

@keyframes bingo-text-bounce {
  0% {
    transform: scale(0.7) rotate(-2deg);
  }
  26% {
    transform: scale(1.06) rotate(1deg);
  }
  48%,
  100% {
    transform: scale(1) rotate(0);
  }
}

@keyframes pastel-firework {
  0% {
    opacity: 0;
    transform: translate(-50%, -50%) scale(0.4);
  }
  18% {
    opacity: 0.95;
  }
  100% {
    opacity: 0;
    transform: translate(calc(-50% + var(--particle-x)), calc(-50% + var(--particle-y))) scale(0.95);
  }
}

@media (prefers-reduced-motion: reduce) {
  .grid-cell--active-reach::before,
  .grid-cell--active-bingo::before,
  .grid-cell--reach-missing :deep(.ball),
  .line-effect,
  .line-effect::before,
  .line-effect::after,
  .bingo-celebration,
  .bingo-celebration__text,
  .bingo-celebration__particle {
    animation-duration: 1ms;
    animation-iteration-count: 1;
  }
}
</style>
