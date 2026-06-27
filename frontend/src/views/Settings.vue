<script setup lang="ts">
import { reactive, ref, onMounted, watch } from 'vue'
import type { FormError, FormSubmitEvent } from '@nuxt/ui'
import { apiClient } from '@/api/apiClient'
import { useCurrentUserStore } from '@/stores/currentUser'
import { useRoomsStore } from '@/stores/rooms'
import { useClipboard } from '@vueuse/core'
import { useToast } from '@nuxt/ui/composables'

const roomsStore = useRoomsStore()

const form = ref()

const submitting = ref(false)
const loading = ref(false)
const errorMessage = ref('')
const editing = ref(false)

const currentUserStore = useCurrentUserStore()
const userId = currentUserStore.userId

const props = defineProps<{
  roomCode: string
}>()

const toast = useToast()

let roomId = ''

onMounted(async () => {
  loading.value = true

  roomId = (await roomsStore.getRoomIdByCode(props.roomCode)) as string

  if (!roomId) {
    loading.value = false
    errorMessage.value = 'コードに対応するルームが見つかりませんでした。'
    return
  }

  const { data, error } = await apiClient.GET('/api/rooms/{roomId}', {
    params: {
      path: {
        roomId,
      },
    },
  })

  if (error) {
    errorMessage.value = error.message
  } else {
    state.name = data.settings.name
    state.description = data.settings.description
    state.adminUserIds = data.settings.admins.map((admin) => admin.userId)
  }

  loading.value = false
})

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

  if (!formState.adminUserIds.some((id) => userId === id)) {
    errors.push({ name: 'admins', message: '自身を管理者から外すことはできません。' })
  }

  return errors
}

async function onSubmit(event: FormSubmitEvent<FormState>) {
  submitting.value = true
  const { error } = await apiClient.PUT('/api/rooms/{roomId}/settings', {
    params: {
      path: {
        roomId,
      },
    },
    body: { settings: event.data },
  })

  submitting.value = false
  editing.value = false

  if (error) {
    console.error('ERROR: ', error)
    toast.add({ title: '設定の変更に失敗しました。' })
    throw new Error('設定の変更に失敗しました。')
  } else {
    toast.add({ title: '設定を更新しました。' })
  }
}

const beforeEdit = ref<FormState>()

watch(
  editing,
  (newValue) => {
    if (newValue) {
      beforeEdit.value = {
        name: state.name,
        description: state.description,
        adminUserIds: [...state.adminUserIds],
      }
    }
  },
  { immediate: true },
)

const baseUrl = window.location.origin

const participantrUrl = ref(`${baseUrl}/${props.roomCode}/participant`)
const displayUrl = ref(`${baseUrl}/${props.roomCode}/display`)
const controlerUrl = ref(`${baseUrl}/${props.roomCode}/controler`)
const { copy } = useClipboard()

const copyUrl = (text: string) => {
  copy(text)
  toast.add({ title: 'URLをコピーしました。', description: text })
}

const cancelEdit = () => {
  editing.value = false
  Object.assign(state, {
    name: (beforeEdit.value as FormState).name,
    description: (beforeEdit.value as FormState).description,
    adminUserIds: [...(beforeEdit.value as FormState).adminUserIds],
  })

  form.value?.clear()
}
</script>

<template>
  <UContainer class="pt-6">
    <div v-if="loading">読み込み中...</div>
    <div v-else-if="errorMessage">{{ errorMessage }}</div>
    <template v-else>
      <h2 class="text-3xl font-bold mb-6">ルーム{{ props.roomCode }}の設定</h2>

      <h3 class="text-2xl font-bold mb-2">情報</h3>
      <UFormField label="参加画面">
        <div class="flex flex-wrap gap-2">
          <UIcon name="i-lucide-rocket" class="size-5 m-1.5" />
          <UFieldGroup class="">
            <UInput
              color="neutral"
              readonly
              v-model="participantrUrl"
              :ui="{
                base: '[direction:rtl] text-left !rounded-r-none !rounded-l-md',
              }"
            />
            <UButton
              color="neutral"
              variant="subtle"
              class="hidden sm:inline-flex rounded-r-md!"
              icon="i-lucide-clipboard"
              @click="copyUrl(participantrUrl)"
            >
              参加画面のURLをコピー
            </UButton>
            <UButton
              color="neutral"
              variant="subtle"
              class="sm:hidden"
              icon="i-lucide-clipboard"
              @click="copyUrl(participantrUrl)"
            />
          </UFieldGroup>
          <UButton
            variant="soft"
            :to="participantrUrl"
            target="_blank"
            icon="i-lucide-square-arrow-out-up-right"
            >参加画面へ移動</UButton
          >
        </div>
      </UFormField>

      <UFormField label="表示画面" class="mt-4">
        <div class="flex flex-wrap gap-2">
          <UIcon name="i-lucide-monitor" class="size-5 m-1.5" />
          <UFieldGroup>
            <UInput
              color="neutral"
              readonly
              v-model="displayUrl"
              :ui="{
                base: '[direction:rtl] text-left !rounded-r-none !rounded-l-md',
              }"
            />
            <UButton
              color="neutral"
              variant="subtle"
              class="hidden sm:inline-flex rounded-r-md!"
              icon="i-lucide-clipboard"
              @click="copyUrl(displayUrl)"
            >
              表示画面のURLをコピー
            </UButton>
            <UButton
              color="neutral"
              variant="subtle"
              class="sm:hidden"
              icon="i-lucide-clipboard"
              @click="copyUrl(displayUrl)"
            />
          </UFieldGroup>
          <UButton
            variant="soft"
            :to="displayUrl"
            target="_blank"
            icon="i-lucide-square-arrow-out-up-right"
            >表示画面へ移動</UButton
          >
        </div>
      </UFormField>

      <UFormField label="操作画面（管理者限定）" class="mt-4">
        <div class="flex flex-wrap gap-2">
          <UIcon name="i-lucide-gamepad" class="size-5 m-1.5" />
          <UFieldGroup>
            <UInput
              color="neutral"
              readonly
              v-model="controlerUrl"
              :ui="{
                base: '[direction:rtl] text-left !rounded-r-none !rounded-l-md',
              }"
            />
            <UButton
              color="neutral"
              variant="subtle"
              class="hidden sm:inline-flex rounded-r-md!"
              icon="i-lucide-clipboard"
              @click="copyUrl(controlerUrl)"
            >
              操作画面のURLをコピー
            </UButton>
            <UButton
              color="neutral"
              variant="subtle"
              class="sm:hidden"
              icon="i-lucide-clipboard"
              @click="copyUrl(controlerUrl)"
            />
          </UFieldGroup>
          <UButton
            variant="soft"
            :to="controlerUrl"
            target="_blank"
            icon="i-lucide-square-arrow-out-up-right"
            >操作画面へ移動</UButton
          >
        </div>
      </UFormField>

      <h3 class="text-2xl font-bold mb-2 mt-6">設定の変更</h3>
      <UForm ref="form" :validate="validate" :state="state" class="space-y-4" @submit="onSubmit">
        <UFormField label="ルーム名" name="name" required>
          <UInput
            v-model="state.name"
            placeholder="イベントの名前・タイトルなど"
            class="w-full"
            :disabled="!editing"
          />
        </UFormField>
        <UFormField label="説明" name="description" required>
          <UTextarea
            v-model="state.description"
            placeholder="イベントの内容・詳細など"
            class="w-full"
            :rows="4"
            :disabled="!editing"
          />
        </UFormField>
        <UFormField label="管理者" name="admins" required>
          <UInputTags
            v-model="state.adminUserIds"
            placeholder="管理者のtraQ IDを入力、Enterで確定"
            class="w-full"
            :disabled="!editing"
          />
        </UFormField>
        <div class="flex justify-end">
          <UButton @click.prevent="editing = true" :loading="submitting" v-if="!editing"
            >設定を編集</UButton
          >
          <div class="flex flex-wrap gap-2" v-else>
            <UButton
              type="button"
              variant="outline"
              :loading="submitting"
              @click.prevent="cancelEdit()"
              >編集をキャンセル</UButton
            >
            <UButton type="submit" variant="solid" :loading="submitting" icon="i-lucide-save"
              >編集を保存</UButton
            >
          </div>
        </div>
      </UForm>
    </template>
  </UContainer>
</template>
