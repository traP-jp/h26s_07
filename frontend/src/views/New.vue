<script setup lang="ts">
import { reactive, watch, ref } from 'vue'
import type { FormError, FormSubmitEvent } from '@nuxt/ui'
import { useRouter } from 'vue-router'
import { apiClient } from '@/api/apiClient'
import { useCurrentUserStore } from '@/stores/currentUser'

const currentUserStore = useCurrentUserStore()
const router = useRouter()

const submitting = ref(false)

type FormState = {
  name: string
  description: string
  adminUserIds: string[]
}

const state = reactive<FormState>({
  name: '',
  description: '',
  adminUserIds: [],
})

watch(
  () => currentUserStore.userId,
  (userId) => {
    if (userId && state.adminUserIds.length === 0) {
      state.adminUserIds = [userId]
    }
  },
  { immediate: true },
)

function validate(formState: FormState): FormError[] {
  const errors: FormError[] = []

  if (!formState.name) {
    errors.push({ name: 'name', message: 'ルーム名を入力してください。' })
  }

  if (!formState.description) {
    errors.push({ name: 'description', message: '説明を入力してください。' })
  }

  if (formState.adminUserIds.length === 0) {
    errors.push({ name: 'admins', message: '管理者を1人以上設定してください。' })
  }

  return errors
}

async function onSubmit(event: FormSubmitEvent<FormState>) {
  submitting.value = true
  const { data, error } = await apiClient.POST('/api/rooms', {
    body: { settings: event.data },
  })

  submitting.value = false

  if (error) {
    console.error('ERROR: ', error)
    alert('新しいページの作成に失敗しました。')
    throw new Error('新しいページの作成に失敗しました。')
  } else {
    router.push(`/${data.roomCode}/settings`)
  }
}
</script>

<template>
  <UContainer class="pt-6">
    <h2 class="text-3xl font-bold mb-2">新しいルームを作成</h2>
    <UForm :validate="validate" :state="state" class="space-y-4 mt-6" @submit="onSubmit">
      <UFormField label="ルーム名" name="name" required>
        <UInput v-model="state.name" placeholder="イベントの名前・タイトルなど" class="w-full" />
      </UFormField>
      <UFormField label="説明" name="description" required>
        <UTextarea
          v-model="state.description"
          placeholder="イベントの内容・詳細など"
          class="w-full"
          :rows="4"
        />
      </UFormField>
      <UFormField label="管理者" name="admins" required>
        <UInputTags
          v-model="state.adminUserIds"
          placeholder="管理者のtraQ IDを入力、Enterで確定"
          class="w-full"
        />
      </UFormField>
      <div class="flex justify-end">
        <UButton type="submit" :loading="submitting">ルームを作成</UButton>
      </div>
    </UForm>
  </UContainer>
</template>
