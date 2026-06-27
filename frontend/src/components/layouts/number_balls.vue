<script setup lang="ts">
const props = defineProps<{
  ballColor: string
  textColor: string
  text: string
  size?: string
}>()
</script>

<template>
  <div
    class="ball"
    :style="{
      backgroundColor: props.ballColor,
      color: props.textColor,
      '--size': props.size ?? '48px',
    }"
  >
    {{ props.text }}
  </div>
</template>
<style scoped>
.ball {
  width: var(--size);
  height: var(--size);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: bold;
  font-size: calc(var(--size) * 0.45);
}
</style>
