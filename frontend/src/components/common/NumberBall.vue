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
    v-if="props.text == 'FREE'"
  >
    <span class="ball-content">
      <UIcon name="i-lucide-star" />
    </span>
  </div>
  <div
    class="ball"
    :style="{
      backgroundColor: props.ballColor,
      color: props.textColor,
    }"
    v-else
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
  box-shadow: 0 3px 9px rgb(0 0 0 / 0.08);
  container-type: inline-size;
  transition:
    background-color 260ms ease,
    color 260ms ease,
    box-shadow 260ms ease,
    transform 260ms ease;
  user-select: none;
  animation: ball-state-pop 560ms cubic-bezier(0.18, 0.85, 0.28, 1.28) both;
}

.ball-content {
  font-size: 42cqw;
  font-weight: 900;
}

@keyframes ball-state-pop {
  0% {
    filter: brightness(1);
    transform: scale(0.86);
  }
  48% {
    filter: brightness(1.2);
    box-shadow:
      0 0 0 0.16em rgb(255 255 255 / 0.92),
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
