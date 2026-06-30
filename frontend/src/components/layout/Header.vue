<script setup lang="ts">
import { useCurrentUserStore } from '@/stores/currentUser'
import { computed } from 'vue'
import { useRouter } from 'vue-router'

const currentUserStore = useCurrentUserStore()
const iconUrl = computed(() =>
  currentUserStore.userId ? `https://q.trap.jp/api/v3/public/icon/${currentUserStore.userId}` : '',
)

const router = useRouter()
</script>

<template>
  <UHeader :toggle="false">
    <template #left>
      <img
        src="/logo.svg"
        alt="完璧で究極のBINGO"
        class="h-6 w-auto cursor-pointer"
        @click="router.push('/')"
      />
    </template>
    <template #right>
      <UButton
        icon="i-lucide-plus"
        to="/new"
        target="_blank"
        aria-label="新しいルームを作成"
        class="sm:hidden rounded-full"
      ></UButton>
      <UButton icon="i-lucide-plus" to="/new" target="_blank" class="hidden sm:inline-flex"
        >新しいルームを作成</UButton
      >
      <UAvatar v-if="currentUserStore.userId" :src="iconUrl" loading="lazy" />
      <div v-if="currentUserStore.userId">{{ currentUserStore.userId }}</div>
    </template>
  </UHeader>
</template>
