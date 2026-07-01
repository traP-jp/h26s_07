<script setup lang="ts">
const props = defineProps<{
  ballColor: string
  textColor: string
  text: string
}>()
</script>

<template>
  <div
    class="ball number-ball"
    :style="{
      backgroundColor: props.ballColor,
      color: props.textColor,
    }"
    v-if="props.text == 'FREE'"
  >
    <UIcon name="i-lucide-star" />
  </div>
  <div
    class="ball number-ball"
    :style="{
      backgroundColor: props.ballColor,
      color: props.textColor,
    }"
    v-else
  >
    {{ props.text }}
  </div>
</template>
<style scoped>
.ball {
  width: 100%;
  max-width: 100%;
  height: auto;
  aspect-ratio: 1 / 1;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: calc(var(--cell-size, 64px) * 0.42);
  font-weight: 900;
  flex-shrink: 0;
  box-sizing: border-box;
  box-shadow: 0 3px 9px rgb(0 0 0 / 0.08);
  transition:
    background-color 260ms ease,
    color 260ms ease,
    box-shadow 260ms ease,
    transform 260ms ease;
  user-select: none;
  animation: ball-state-pop 560ms cubic-bezier(0.18, 0.85, 0.28, 1.28) both;
}

@keyframes ball-state-pop {
  0% {
    filter: brightness(1);
    transform: scale(0.86);
  }
  48% {
    filter: brightness(1.2);
    box-shadow:
      0 0 0 calc(var(--cell-size, 64px) * 0.07) rgb(255 255 255 / 0.92),
      0 0 24px rgb(255 220 242 / 0.62);
    transform: scale(1.12);
  }
  100% {
    filter: brightness(1);
    transform: scale(1);
  }
}

@media (prefers-reduced-motion: reduce) {
  .ball {
    transition-duration: 1ms;
    animation-duration: 1ms;
  }
}
</style>
