<script setup lang="ts">
import { computed } from 'vue'
import type { UserId } from '@/api/schema'
const userId = defineProps<{
  userId: UserId
}>()
const url = computed(() => `https://q.trap.jp/api/v3/public/icon/${userId.userId}`)
</script>

<template>
  <div>
    <img class="icon" :src="url" />
  </div>
</template>

<style scoped>
.icon {
  width: 100%;
  height: 100%;
  border-radius: 50%;
}
</style>
