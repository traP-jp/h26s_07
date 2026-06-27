<script setup lang="ts">
const props = defineProps<{
  ballColor: string
  textColor: string
  text: string
  size: number
}>()
</script>

<template>
  <div
    class="ball"
    :style="{
      backgroundColor: props.ballColor,
      color: props.textColor,
      width: `${props.size}px`,
      height: `${props.size}px`,
      fontSize: `${props.size * 0.42}px`,
    }"
  >
    {{ props.text }}
  </div>
</template>
<style scoped>
.ball {
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 900;
  aspect-ratio: 1 / 1;
  flex-shrink: 0;
  box-sizing: border-box;
  user-select: none;
}
</style>
