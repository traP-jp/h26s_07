<script setup lang="ts">
import type { CardCell } from '@/api/schema'
import NumberBall from '@/components/common/NumberBall.vue'
import { computed } from 'vue'
const props = defineProps<{
  cell: CardCell
  size?: number
}>()

const ballColor = computed(() => {
  switch (props.cell.cellState) {
    case 'bingo':
      return '#ff9800'
    case 'reach':
      return '#f06fb1'
    case 'open':
      return '#4caf50'
    case 'closed':
    default:
      return '#dddddd'
  }
})

const textColor = computed(() => {
  switch (props.cell.cellState) {
    case 'bingo':
    case 'reach':
    case 'open':
      return '#ffffff'
    case 'closed':
    default:
      return '#333333'
  }
})
</script>

<template>
  <NumberBall
    :key="props.cell.cellState"
    :ball-color="ballColor"
    :text-color="textColor"
    :text="props.cell.displayText"
    :size="props.size"
  />
</template>
