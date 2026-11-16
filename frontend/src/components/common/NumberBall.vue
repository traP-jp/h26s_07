<script setup lang="ts">
const props = defineProps<{
  ballColor: string
  textColor: string
  text: string
}>()
</script>

<template>
  <div
    class="ball"
    :style="{
      backgroundColor: props.ballColor,
      color: props.textColor,
    }"
  >
    <span class="ball-content">{{ props.text }}</span>
  </div>
</template>
<style scoped>
.ball {
  width: 100%;
  height: auto;
  aspect-ratio: 1 / 1;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow:
    0 1px 2px rgb(31 38 42 / 0.12),
    inset 0 1px 0 rgb(255 255 255 / 0.34);
  container-type: inline-size;
  transition:
    background-color 180ms ease,
    color 180ms ease,
    box-shadow 180ms ease;
  user-select: none;
}

.ball-content {
  font-size: clamp(14px, 42cqw, 48px);
  font-weight: 850;
  font-variant-numeric: tabular-nums;
  line-height: 1;
}

.ball--closed {
  box-shadow: inset 0 1px 0 rgb(255 255 255 / 0.52);
}

.ball--open {
  box-shadow:
    0 2px 5px rgb(34 65 48 / 0.18),
    inset 0 1px 0 rgb(255 255 255 / 0.3);
}

@media (prefers-reduced-motion: reduce) {
  .ball {
    transition: none;
  }
}
</style>
