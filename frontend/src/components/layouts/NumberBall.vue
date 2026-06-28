<script setup lang="ts">
const props = defineProps<{
  ballColor: string
  textColor: string
  text: string
  size?: number
}>()
</script>

<template>
  <div
    class="ball"
    :style="{
      backgroundColor: props.ballColor,
      color: props.textColor,
      width: props.size === undefined ? undefined : `${props.size}px`,
      height: props.size === undefined ? undefined : `${props.size}px`,
      fontSize: props.size === undefined ? undefined : `${props.size * 0.42}px`,
    }"
    v-if="props.text == 'FREE'"
  >
    <UIcon name="i-lucide-star" />
  </div>
  <div
    class="ball"
    :style="{
      backgroundColor: props.ballColor,
      color: props.textColor,
      width: props.size === undefined ? undefined : `${props.size}px`,
      height: props.size === undefined ? undefined : `${props.size}px`,
      fontSize: props.size === undefined ? undefined : `${props.size * 0.42}px`,
    }"
    v-else
  >
    {{ props.text }}
  </div>
</template>
<style scoped>
.ball {
  width: calc(var(--cell-size, 64px) * 0.75);
  height: calc(var(--cell-size, 64px) * 0.75);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: calc(var(--cell-size, 64px) * 0.315);
  font-weight: 900;
  aspect-ratio: 1 / 1;
  flex-shrink: 0;
  box-sizing: border-box;
  user-select: none;
}
</style>
