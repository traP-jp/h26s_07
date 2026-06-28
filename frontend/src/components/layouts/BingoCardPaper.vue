<script setup lang="ts">
import type { Card, CardCell as CardCellType } from '@/api/schema'
import CardCell from '@/components/layouts/CardCell.vue'
import { computed } from 'vue'

const props = withDefaults(
  defineProps<{
    card?: Card | null
    cellSize?: number | string
    placeholder?: boolean
  }>(),
  { cellSize: 48, placeholder: false },
)

const cardNo = computed(() => props.card?.cardNumber ?? 'not assigned')
const size = computed(() => props.cellSize)
const cellSizeStyle = computed(() =>
  typeof size.value === 'number' ? `${size.value}px` : size.value,
)
const ballSize = computed(() => (typeof size.value === 'number' ? size.value * 0.75 : undefined))
const title = ['B', 'I', 'N', 'G', 'O']
const placeholderCells: CardCellType[] = Array.from({ length: 25 }, (_, index) => ({
  index,
  number: index === 12 ? null : 0,
  displayText: index === 12 ? 'FREE' : '?',
  cellState: index === 12 ? 'open' : 'closed',
}))
const cells = computed(() => props.card?.cells ?? placeholderCells)
</script>

<template>
  <div
    class="bingo-paper"
    :class="{ 'bingo-paper--placeholder': props.placeholder || !props.card }"
    :style="{ '--cell-size': cellSizeStyle }"
  >
    <div class="bingo-title">
      <div v-for="letter in title" :key="letter" class="bingo-title-cell">
        {{ letter }}
      </div>
    </div>

    <div class="bingo-grid">
      <div v-for="cell in cells" :key="cell.index" class="grid-cell">
        <CardCell :cell="cell" :size="ballSize" />
      </div>
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
  display: inline-block;
  width: fit-content;
  max-width: 100%;
  padding: calc(var(--cell-size) * 0.2) calc(var(--cell-size) * 0.2) calc(var(--cell-size) * 0.1);
  overflow: hidden;
  background: #c32020;
  box-shadow: 0 4px 14px rgb(0 0 0 / 18%);
}

.bingo-paper--placeholder .grid-cell {
  background: #f3f5f8;
}

.bingo-paper--placeholder :deep(.number-ball) {
  opacity: 0.35;
}

.bingo-title {
  display: grid;
  grid-template-columns: repeat(5, 1fr);
  margin-bottom: 0.5em;
}

.bingo-title-cell {
  text-align: center;
  color: #ffffff;
  font-weight: 900;
  letter-spacing: 0.05em;
  line-height: 1;
  font-size: calc(var(--cell-size) * 0.6);
}

.bingo-grid {
  display: grid;
  grid-template-rows: repeat(5, var(--cell-size));
  grid-auto-columns: var(--cell-size);
  gap: 0;
  border-top: 4px solid #c32020;
  border-left: 4px solid #c32020;
  background: #c32020;
  place-content: center;
  grid-auto-flow: column;
}

.grid-cell {
  background: #ffffff;
  border-right: 4px solid #c32020;
  border-bottom: 4px solid #c32020;

  display: flex;
  align-items: center;
  justify-content: center;
}

.bingo-footer {
  margin-top: 8px;
  display: flex;
  justify-content: end;
  align-items: center;
  color: #ffffff;
  font-style: italic;
  min-width: 0;
}

.card-no {
  display: block;
  width: 100%;
  min-width: 0;
  max-width: 100%;
  overflow-wrap: anywhere;
  word-break: break-all;
  line-height: 1.2;
  font-size: calc(var(--cell-size) * 0.16);
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
</style>
