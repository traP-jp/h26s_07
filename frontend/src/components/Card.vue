<script setup lang="ts">
import { computed } from 'vue'
interface Props {
  title: string
  description: string
  roomCode: string
  isAdmin?: boolean
}
const props = withDefaults(defineProps<Props>(), {
  isAdmin: false,
})
const joinUrl = computed(() => `./${props.roomCode}/participant`)
const displayUrl = computed(() => `./${props.roomCode}/display`)
const settingsUrl = computed(() => `./${props.roomCode}/settings`)
const controlerUrl = computed(() => `./${props.roomCode}/controler`)
</script>

<template>
  <UCard class="w-full p-1" variant="outline" :ui="{ body: 'p-2 sm:p-2' }">
    <h3 class="font-bold text-2xl">{{ title }}</h3>
    <p>{{ description }}</p>
    <div class="flex gap-2 mt-2 flex-wrap">
      <UButton
        color="primary"
        variant="solid"
        icon="i-lucide-rocket"
        size="md"
        :to="joinUrl"
        target="_blank"
        >ゲームに参加</UButton
      >
      <UButton
        color="primary"
        variant="soft"
        icon="i-lucide-monitor"
        size="md"
        :to="displayUrl"
        target="_blank"
        >表示用画面</UButton
      >
      <UButton
        color="primary"
        variant="outline"
        icon="i-lucide-settings"
        size="md"
        :to="settingsUrl"
        target="_blank"
        v-if="isAdmin"
        >設定</UButton
      >
      <UButton
        color="primary"
        variant="outline"
        icon="i-lucide-gamepad"
        size="md"
        :to="controlerUrl"
        target="_blank"
        v-if="isAdmin"
        >操作</UButton
      >
    </div>
  </UCard>
</template>
