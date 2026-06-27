import { ref } from 'vue'
import { defineStore } from 'pinia'

import { apiClient } from '@/api/apiClient'

import type { UserId } from '@/api/schema'

export const useCurrentUserStore = defineStore('currentUser', () => {
  const userId = ref<UserId | null>(null)

  async function fetchUserId() {
    const { data, error } = await apiClient.GET('/api/me')

    if (error !== undefined) {
      userId.value = null
      throw new Error('Failed to fetch current user')
    }

    userId.value = data.userId
  }

  async function ensureUserId() {
    if (userId.value !== null) {
      return
    }

    await fetchUserId()
  }

  function clear() {
    userId.value = null
  }

  return {
    userId,
    fetchUserId,
    ensureUserId,
    clear,
  }
})
