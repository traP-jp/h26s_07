import { computed, ref } from 'vue'
import { defineStore } from 'pinia'

import { apiClient } from '@/api/apiClient'

import type { UserId } from '@/api/schema'

export const useCurrentUserStore = defineStore('currentUser', () => {
  const currentUserId = ref<UserId | null>(null)

  async function init() {
    if (currentUserId.value !== null) {
      return
    }

    const { data, error } = await apiClient.GET('/api/me')

    if (error !== undefined || data === undefined) {
      throw new Error('Failed to fetch current user')
    }

    currentUserId.value = data.userId
  }

  return {
    userId: computed(() => {
      if (currentUserId.value === null) {
        throw new Error('ユーザー情報が初期化されていません')
      }

      return currentUserId.value
    }),
    init,
  }
})
