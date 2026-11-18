import { ref } from 'vue'
import { defineStore } from 'pinia'

import type { User, UserId } from '@/api/schema'

export const useCurrentUserStore = defineStore('currentUser', () => {
  const userId = ref<UserId | null>(null)

  async function init(): Promise<void> {
    const response = await fetch('/api/me', {
      headers: {
        Accept: 'application/json',
      },
      redirect: 'manual',
    })

    if (response.type === 'opaqueredirect' || response.status === 0) {
      return
    }

    if (!response.ok) {
      throw new Error('Failed to fetch current user')
    }

    const data = (await response.json()) as User
    userId.value = data.userId
  }

  return {
    userId,
    init,
  }
})
