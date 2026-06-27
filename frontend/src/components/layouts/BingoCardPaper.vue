<script setup lang="ts">
import type { Card } from '@/api/schema'
import CardCell from '@/components/layouts/CardCell.vue'

const props = withDefaults(
  defineProps<{
    card: Card
    cellSize?: number
  }>(),
  { cellSize: 48 },
)

const cardNo = props.card.cardNumber
const size = props.cellSize
const title = ['B', 'I', 'N', 'G', 'O']
</script>

<template>
  <div class="bingo-paper" :style="{ padding: `${size * 0.2}px ${size * 0.2}px  ${size * 0.1}px` }">
    <div class="bingo-title">
      <div
        v-for="letter in title"
        :key="letter"
        class="bingo-title-cell"
        :style="{ fontSize: `${size * 0.6}px` }"
      >
        {{ letter }}
      </div>
    </div>

    <div
      class="bingo-grid"
      :style="{
        gridTemplateColumns: `repeat(5, ${size}px)`,
        gridAutoRows: `${size}px`,
      }"
    >
      <div v-for="cell in props.card.cells" :key="cell.index" class="grid-cell">
        <CardCell :cell="cell" :size="size * 0.75" />
      </div>
    </div>

    <div class="bingo-footer">
      <span class="card-no" :style="{ fontSize: `${size * 0.2}px` }">Card No. {{ cardNo }}</span>
    </div>
  </div>
</template>

<style scoped>
.bingo-paper {
  display: inline-block;
  background: #c32020;
  box-shadow: 0 4px 14px rgb(0 0 0 / 18%);
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
}

.bingo-grid {
  display: grid;
  gap: 0;
  border-top: 4px solid #c32020;
  border-left: 4px solid #c32020;
  background: #c32020;
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
}

.card-no {
  line-height: 1.2;
}
</style>
