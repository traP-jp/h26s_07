<script setup lang="ts">
import type { Card, CardCell as CardCellType, CardChanges, Line } from '@/api/schema'
import CardCell from '@/components/bingo/CardCell.vue'
import { computed, ref, watch } from 'vue'

const props = withDefaults(
  defineProps<{
    card?: Card | null
    cardChanges?: CardChanges | null
    drawKey?: number | null
    placeholder?: boolean
  }>(),
  { cardChanges: null, drawKey: null, placeholder: false },
)

type CardEffectMode = 'bingo' | 'reach'
const cardNo = computed(() => props.card?.cardNumber ?? 'not assigned')
const title = ['B', 'I', 'N', 'G', 'O']
const placeholderCells: CardCellType[] = Array.from({ length: 25 }, (_, index) => ({
  index,
  number: index === 12 ? null : 0,
  displayText: index === 12 ? 'FREE' : '?',
  cellState: index === 12 ? 'open' : 'closed',
}))
const cells = computed(() => props.card?.cells ?? placeholderCells)
const accessibleRows = computed(() =>
  Array.from({ length: 5 }, (_, row) =>
    Array.from({ length: 5 }, (_, column) => cells.value[column * 5 + row]).filter(
      (cell): cell is CardCellType => cell !== undefined,
    ),
  ),
)
const cellStateLabel = (cell: CardCellType) => {
  switch (cell.cellState) {
    case 'open':
      return '抽選済み'
    case 'reach':
      return 'リーチ'
    case 'bingo':
      return 'ビンゴ'
    case 'closed':
    default:
      return '未抽選'
  }
}

// A reconnect can restore the last card snapshot. Only a changed draw key should replay motion.
const activeDrawKey = ref<number | null>(null)
watch(
  () => props.drawKey,
  (drawKey, previousDrawKey) => {
    if (drawKey != null && drawKey !== previousDrawKey) activeDrawKey.value = drawKey
  },
)

const activeEffect = computed<{ mode: CardEffectMode; line: Line } | null>(() => {
  if (props.drawKey == null || activeDrawKey.value !== props.drawKey) return null
  const bingoLine = props.cardChanges?.newBingoLines.at(-1)
  if (bingoLine) return { mode: 'bingo', line: bingoLine }
  const reachLine = props.cardChanges?.newReachLines.at(-1)
  if (reachLine) return { mode: 'reach', line: reachLine }
  return null
})
const activeLine = computed(() => activeEffect.value?.line ?? [])
const activeEffectKey = computed(() =>
  activeEffect.value
    ? `${props.drawKey}-${activeEffect.value.mode}-${activeLine.value.join('-')}`
    : 'none',
)

const isInLine = (lines: Line[] | undefined, cellIndex: number) =>
  lines?.some((line) => line.includes(cellIndex)) ?? false

const gridCellClass = (cell: CardCellType) => ({
  'grid-cell--line-bingo': isInLine(props.card?.bingoLines, cell.index),
  'grid-cell--line-reach': isInLine(props.card?.reachLines, cell.index),
  'grid-cell--active-line': activeLine.value.includes(cell.index),
  'grid-cell--active-bingo':
    activeEffect.value?.mode === 'bingo' && activeLine.value.includes(cell.index),
  'grid-cell--active-reach':
    activeEffect.value?.mode === 'reach' && activeLine.value.includes(cell.index),
  'grid-cell--reach-missing':
    isInLine(props.card?.reachLines, cell.index) && cell.cellState === 'closed',
  'grid-cell--new-draw':
    activeDrawKey.value === props.drawKey &&
    (props.cardChanges?.openedCellIndices.includes(cell.index) ?? false),
})

const gridCellStyle = (cell: CardCellType) => {
  const lineIndex = activeLine.value.indexOf(cell.index)
  return { '--cell-delay': `${lineIndex > 0 ? lineIndex * 55 : 0}ms` }
}

function cellCenter(index: number) {
  const column = Math.floor(index / 5)
  const row = index % 5
  return { x: ((column + 0.5) / 5) * 100, y: ((row + 0.5) / 5) * 100 }
}

const activeLineStyle = computed(() => {
  const line = activeLine.value
  if (line.length < 2) return {}
  const start = cellCenter(line[0]!)
  const end = cellCenter(line[line.length - 1]!)
  return {
    '--line-x': `${start.x}%`,
    '--line-y': `${start.y}%`,
    '--line-length': `${Math.hypot(end.x - start.x, end.y - start.y)}%`,
    '--line-angle': `${Math.atan2(end.y - start.y, end.x - start.x)}rad`,
  }
})
</script>

<template>
  <div class="bingo-card-root">
    <table v-if="card" class="bingo-screen-reader-table">
      <caption>
        ビンゴカード、カード番号
        {{
          cardNo
        }}
      </caption>
      <thead>
        <tr>
          <th v-for="letter in title" :key="letter" scope="col">{{ letter }}</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="(row, rowIndex) in accessibleRows" :key="rowIndex">
          <td v-for="cell in row" :key="cell.index">
            {{ cell.displayText }}、{{ cellStateLabel(cell) }}
          </td>
        </tr>
      </tbody>
    </table>

    <div
      :key="activeEffectKey"
      class="bingo-paper"
      :class="{
        'bingo-paper--won': card?.bingoLines.length,
        'bingo-paper--fluttering': activeEffect !== null,
        'bingo-paper--fluttering-bingo': activeEffect?.mode === 'bingo',
      }"
      :aria-hidden="card ? 'true' : undefined"
    >
      <div class="bingo-title" aria-label="Bingo">
        <div v-for="letter in title" :key="letter" class="bingo-title-cell">{{ letter }}</div>
      </div>

      <div class="bingo-cells">
        <div
          v-for="cell in cells"
          :key="cell.index"
          class="grid-cell"
          :class="gridCellClass(cell)"
          :style="gridCellStyle(cell)"
        >
          <CardCell :cell="cell" :style="{ width: '78%', height: '78%' }" />
        </div>
        <div
          v-if="activeEffect"
          :key="activeEffectKey"
          class="line-effect"
          :class="`line-effect--${activeEffect.mode}`"
          :style="activeLineStyle"
        ></div>
      </div>

      <div
        v-if="activeEffect?.mode === 'bingo'"
        :key="`celebration-${activeEffectKey}`"
        class="bingo-celebration"
      >
        BINGO
      </div>

      <div class="card-no">Card No. {{ cardNo }}</div>

      <div v-if="props.placeholder || !props.card" class="bingo-placeholder-message">
        カードはまだ配られていません
      </div>
    </div>

    <div
      v-if="activeEffect?.mode === 'bingo'"
      class="bingo-screen-reader-status"
      aria-live="polite"
    >
      BINGO
    </div>
  </div>
</template>

<style scoped>
.bingo-card-root {
  position: relative;
  width: 100%;
}

.bingo-screen-reader-table,
.bingo-screen-reader-status {
  position: fixed;
  top: 0;
  left: 0;
  width: 1px;
  height: 1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  clip-path: inset(50%);
  white-space: nowrap;
}

.bingo-paper {
  position: relative;
  width: 100%;
  box-sizing: border-box;
  padding: 4.1% 3.5% 2%;
  border: 1px solid #81271f;
  border-radius: 10px;
  background: #a9362c;
  box-shadow:
    0 14px 32px rgb(41 29 25 / 18%),
    inset 0 0 0 4px rgb(255 240 218 / 18%);
  container-type: inline-size;
}

.bingo-title {
  display: grid;
  margin-bottom: 2.2%;
  grid-template-columns: repeat(5, 1fr);
}

.bingo-title-cell {
  color: #fff4df;
  font-size: 11cqw;
  font-weight: 900;
  line-height: 1;
  text-align: center;
  text-shadow: 0 1px 0 rgb(72 22 18 / 28%);
  user-select: none;
}

.bingo-cells {
  position: relative;
  display: grid;
  aspect-ratio: 1;
  overflow: hidden;
  border: 3px solid #8f2b23;
  border-radius: 5px;
  background: #8f2b23;
  grid-template-columns: repeat(5, 1fr);
  grid-template-rows: repeat(5, 1fr);
  grid-auto-flow: column;
}

.grid-cell {
  position: relative;
  z-index: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 1px solid #c9bca8;
  background: #fffaf0;
}

.grid-cell--line-reach:not(.grid-cell--line-bingo) {
  background: #fff0dd;
}

.grid-cell--line-bingo {
  background: #f5e6b7;
}

.grid-cell--active-line::before {
  position: absolute;
  inset: 8%;
  z-index: 0;
  border-radius: 50%;
  background: #ebc66f;
  content: '';
  opacity: 0;
  pointer-events: none;
  animation: cell-settle 520ms cubic-bezier(0.16, 0.84, 0.26, 1) var(--cell-delay, 0ms) both;
}

.grid-cell--active-reach::before {
  background: #edc47d;
}

.grid-cell--active-bingo::before {
  background: #e1b955;
}

.grid-cell--new-draw :deep(.ball) {
  animation: ball-land 480ms cubic-bezier(0.16, 0.84, 0.24, 1) both;
}

.grid-cell--reach-missing :deep(.ball) {
  outline: 2px dashed #b14a4b;
  outline-offset: 2px;
}

.line-effect {
  position: absolute;
  top: var(--line-y);
  left: var(--line-x);
  z-index: 3;
  width: var(--line-length);
  height: 1.3cqw;
  border-radius: 999px;
  background: #bd6348;
  opacity: 0;
  pointer-events: none;
  transform: translateY(-50%) rotate(var(--line-angle));
  transform-origin: 0 50%;
}

.line-effect--reach {
  background: #bb6a4f;
  animation: line-draw 560ms cubic-bezier(0.18, 0.82, 0.22, 1) 480ms both;
}

.line-effect--bingo {
  height: 1.8cqw;
  background: #b08a34;
  animation: line-draw 620ms cubic-bezier(0.18, 0.82, 0.22, 1) 480ms both;
}

.bingo-celebration {
  position: absolute;
  top: 50%;
  left: 50%;
  z-index: 5;
  padding: 2.2cqw 5.3cqw;
  border: 2px solid #9d7623;
  border-radius: 999px;
  background: #fff8df;
  box-shadow: 0 5px 16px rgb(54 39 19 / 20%);
  color: #87372b;
  font-size: clamp(20px, 7.2cqw, 52px);
  font-weight: 950;
  letter-spacing: 0.06em;
  line-height: 1;
  pointer-events: none;
  transform: translate(-50%, -50%);
  animation: bingo-stamp 900ms cubic-bezier(0.16, 0.84, 0.24, 1) 980ms both;
}

.bingo-paper--won {
  box-shadow:
    0 14px 32px rgb(41 29 25 / 18%),
    inset 0 0 0 4px rgb(255 240 218 / 28%);
}

.bingo-paper--fluttering {
  transform-origin: 52% 48%;
  animation: paper-flutter 980ms cubic-bezier(0.24, 0.72, 0.22, 1) 420ms both;
  transform-style: preserve-3d;
}

.bingo-paper--fluttering::after {
  position: absolute;
  bottom: 1.4%;
  left: 3.5%;
  z-index: 4;
  width: 12cqw;
  aspect-ratio: 1;
  border-bottom-left-radius: 2px;
  background: linear-gradient(135deg, #fff9ec 5%, #f4e9d4 54%, #b7a78d 100%);
  clip-path: polygon(0 0, 0 100%, 100% 100%);
  content: '';
  filter: drop-shadow(2px -1px 2px rgb(44 24 18 / 28%));
  opacity: 0;
  pointer-events: none;
  transform-origin: left bottom;
  animation: paper-corner-lift 760ms cubic-bezier(0.2, 0.75, 0.24, 1) 460ms both;
}

.bingo-paper--fluttering-bingo {
  animation-name: paper-flutter-bingo;
  animation-duration: 1180ms;
}

.bingo-paper--fluttering-bingo::after {
  animation-name: paper-corner-lift-bingo;
  animation-duration: 940ms;
}

.card-no {
  margin-top: 2.5%;
  overflow-wrap: anywhere;
  color: #fff4df;
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: clamp(11px, 3cqw, 14px);
  font-variant-numeric: tabular-nums;
  font-weight: 600;
  line-height: 1.2;
  text-align: right;
}

.bingo-placeholder-message {
  position: absolute;
  top: 50%;
  left: 50%;
  width: 82%;
  padding: 10px 12px;
  transform: translate(-50%, -50%);
  border-radius: 8px;
  background: #fffaf0;
  box-shadow: 0 6px 20px rgb(0 0 0 / 16%);
  color: #24364d;
  font-size: clamp(14px, 4.3cqw, 20px);
  font-weight: 800;
  line-height: 1.4;
  text-align: center;
}

@keyframes ball-land {
  0% {
    transform: translateY(-22%) scale(0.8);
  }
  70% {
    transform: translateY(2%) scale(1.08);
  }
  100% {
    transform: translateY(0) scale(1);
  }
}

@keyframes cell-settle {
  0% {
    opacity: 0;
    transform: scale(0.7);
  }
  42% {
    opacity: 0.48;
    transform: scale(1.04);
  }
  100% {
    opacity: 0;
    transform: scale(1);
  }
}

@keyframes line-draw {
  0% {
    opacity: 0;
    transform: translateY(-50%) rotate(var(--line-angle)) scaleX(0);
  }
  12% {
    opacity: 1;
  }
  78% {
    opacity: 1;
    transform: translateY(-50%) rotate(var(--line-angle)) scaleX(1);
  }
  100% {
    opacity: 0;
    transform: translateY(-50%) rotate(var(--line-angle)) scaleX(1);
  }
}

@keyframes bingo-stamp {
  0% {
    opacity: 0;
    transform: translate(-50%, -50%) scale(0.72) rotate(-5deg);
  }
  35% {
    opacity: 1;
    transform: translate(-50%, -50%) scale(1.05) rotate(1deg);
  }
  62% {
    opacity: 1;
    transform: translate(-50%, -50%) scale(1) rotate(0);
  }
  100% {
    opacity: 0;
    transform: translate(-50%, -50%) scale(0.98) rotate(0);
  }
}

@keyframes paper-flutter {
  0%,
  100% {
    transform: perspective(1100px) rotateX(0) rotateY(0) rotateZ(0) scale(1);
  }
  24% {
    transform: perspective(1100px) rotateX(2deg) rotateY(-3deg) rotateZ(-0.65deg)
      scale(0.996, 1.004);
  }
  52% {
    transform: perspective(1100px) rotateX(-1.8deg) rotateY(3.6deg) rotateZ(0.7deg)
      scale(1.004, 0.996);
  }
  77% {
    transform: perspective(1100px) rotateX(0.8deg) rotateY(-1.5deg) rotateZ(-0.25deg)
      scale(0.999, 1.001);
  }
}

@keyframes paper-flutter-bingo {
  0%,
  100% {
    transform: perspective(950px) rotateX(0) rotateY(0) rotateZ(0) scale(1);
  }
  16% {
    transform: perspective(950px) rotateX(3deg) rotateY(-5deg) rotateZ(-1deg) scale(0.994, 1.006);
  }
  34% {
    transform: perspective(950px) rotateX(-3.2deg) rotateY(5.4deg) rotateZ(1deg) scale(1.006, 0.994);
  }
  53% {
    transform: perspective(950px) rotateX(2deg) rotateY(-3.4deg) rotateZ(-0.65deg)
      scale(0.997, 1.003);
  }
  74% {
    transform: perspective(950px) rotateX(-1deg) rotateY(1.8deg) rotateZ(0.3deg) scale(1.002, 0.998);
  }
}

@keyframes paper-corner-lift {
  0%,
  100% {
    opacity: 0;
    transform: perspective(180px) rotateX(0) rotateY(0);
  }
  24%,
  62% {
    opacity: 0.96;
    transform: perspective(180px) rotateX(25deg) rotateY(-28deg) translate(2px, -2px);
  }
  82% {
    opacity: 0.55;
    transform: perspective(180px) rotateX(-8deg) rotateY(12deg) translate(-1px, 0);
  }
}

@keyframes paper-corner-lift-bingo {
  0%,
  100% {
    opacity: 0;
    transform: perspective(180px) rotateX(0) rotateY(0);
  }
  18%,
  46% {
    opacity: 1;
    transform: perspective(180px) rotateX(38deg) rotateY(-42deg) translate(3px, -4px);
  }
  68% {
    opacity: 0.82;
    transform: perspective(180px) rotateX(-22deg) rotateY(24deg) translate(-2px, 1px);
  }
  86% {
    opacity: 0.42;
    transform: perspective(180px) rotateX(10deg) rotateY(-12deg);
  }
}

@media (prefers-reduced-motion: reduce) {
  .grid-cell--active-line::before,
  .grid-cell--new-draw :deep(.ball),
  .line-effect,
  .bingo-celebration {
    animation: none !important;
  }

  .bingo-celebration {
    opacity: 0;
  }

  .bingo-paper--fluttering,
  .bingo-paper--fluttering::after {
    animation: none !important;
  }
}
</style>
