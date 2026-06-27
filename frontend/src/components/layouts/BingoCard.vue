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

const size = props.cellSize
</script>

<template>
  <div
    class="bingo-card"
    :style="{
      gridTemplateColumns: `repeat(5, ${size}px)`,
      gridAutoRows: `${size}px`,
    }"
  >
    <CardCell v-for="cell in props.card.cells" :key="cell.index" :cell="cell" :size="size" />
  </div>
</template>

<style scoped>
.bingo-card {
  display: grid;
  gap: 8px;
}
</style>
