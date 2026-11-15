<script setup lang="ts">
import type { CardCell } from '@/api/schema'
import NumberBall from '@/components/common/NumberBall.vue'
import { computed } from 'vue'
const props = defineProps<{
  cell: CardCell
}>()

const ballColor = computed(() => {
  switch (props.cell.cellState) {
    case 'bingo':
      return '#9b6327'
    case 'reach':
      return '#a44350'
    case 'open':
      return '#49765d'
    case 'closed':
    default:
      return '#e7e0d2'
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
    :ball-color="ballColor"
    :text-color="textColor"
    :text="props.cell.displayText"
    :aria-label="props.cell.displayText === 'FREE' ? 'Free cell' : props.cell.displayText"
    :class="`ball--${props.cell.cellState}`"
  />
</template>
