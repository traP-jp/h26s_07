import { computed, ref } from 'vue'
import { defineStore } from 'pinia'

import type { User, UserId } from '@/api/schema'

export const useCurrentUserStore = defineStore('currentUser', () => {
  const currentUserId = ref<UserId | null>(null)
  const initialized = ref(false)
  const loginRequired = ref(false)
  let initPromise: Promise<void> | null = null

  async function init(): Promise<void> {
    if (initialized.value) {
      return
    }
    if (initPromise !== null) {
      return initPromise
    }

    initPromise = fetchCurrentUser()
    await initPromise
  }

  async function getUserId(): Promise<UserId> {
    await init()
    if (currentUserId.value === null) {
      throw new Error('Failed to fetch current user')
    }
    return currentUserId.value
  }

  async function fetchCurrentUser(): Promise<void> {
    loginRequired.value = false

    const response = await fetch('/api/me', {
      headers: {
        Accept: 'application/json',
      },
      redirect: 'manual',
    })

    if (response.type === 'opaqueredirect' || response.status === 0) {
      loginRequired.value = true
      initialized.value = true
      return
    }

    if (!response.ok) {
      throw new Error('Failed to fetch current user')
    }

    const data = (await response.json()) as User
    currentUserId.value = data.userId
    initialized.value = true
  }

  return {
    initialized,
    loginRequired,
    userId: computed(() => currentUserId.value),
    getUserId,
  }
})
