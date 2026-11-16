<script setup lang="ts">
import { computed } from 'vue'

interface Props {
  title: string
  description: string
  roomCode: string
  isAdmin?: boolean
  isJoinable?: boolean
}

const { title, description, roomCode, isAdmin = false, isJoinable = false } = defineProps<Props>()

const joinUrl = computed(() => `/${roomCode}/participant`)
const displayUrl = computed(() => `/${roomCode}/display`)
const settingsUrl = computed(() => `/${roomCode}/settings`)
const controllerUrl = computed(() => `/${roomCode}/controller`)

const emit = defineEmits<{
  join: []
}>()
</script>

<template>
  <UCard class="w-full p-1" variant="outline" :ui="{ body: 'p-2 sm:p-2' }">
    <h3 class="font-bold text-2xl">{{ title }}</h3>
    <p>{{ description }}</p>

    <div class="flex gap-2 mt-2 flex-wrap">
      <UButton
        v-if="isJoinable"
        @click.stop.prevent="emit('join')"
        color="primary"
        variant="solid"
        icon="i-lucide-rocket"
        size="md"
      >
        ゲームに参加
      </UButton>

      <UButton
        v-else
        color="primary"
        variant="solid"
        icon="i-lucide-rocket"
        size="md"
        :to="joinUrl"
        target="_blank"
      >
        ゲームに参加
      </UButton>

      <UButton
        color="primary"
        variant="soft"
        icon="i-lucide-monitor"
        size="md"
        :to="displayUrl"
        target="_blank"
      >
        表示用画面
      </UButton>

      <UButton
        v-if="isAdmin"
        color="primary"
        variant="outline"
        icon="i-lucide-settings"
        size="md"
        :to="settingsUrl"
        target="_blank"
      >
        設定
      </UButton>

      <UButton
        v-if="isAdmin"
        color="primary"
        variant="outline"
        icon="i-lucide-gamepad"
        size="md"
        :to="controllerUrl"
        target="_blank"
      >
        操作
      </UButton>
    </div>
  </UCard>
</template>
