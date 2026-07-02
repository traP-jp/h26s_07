<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { apiClient } from '@/api/apiClient'
import { useRoomsStore } from '@/stores/rooms'
import { useRouter } from 'vue-router'

const router = useRouter()
const roomsStore = useRoomsStore()

const loading = ref(false)
const errorMessage = ref('')

const props = defineProps<{
  roomCode: string
}>()

let roomId = ''

const isDisabled = ref(false)

onMounted(async () => {
  loading.value = true

  roomId = (await roomsStore.getRoomIdByCode(props.roomCode)) as string

  if (!roomId) {
    loading.value = false
    errorMessage.value = 'コードに対応するルームが見つかりませんでした。'
    return
  }

  await getLatestInfo()
  loading.value = false
})

const gameStarted = ref(true)
const pickStarted = ref(false)
const pickExhausted = ref(true)
const qrCodeShowed = ref(false)
const startModal = ref(false)
const finishModal = ref(false)
const finishedModal = ref(false)

const getLatestInfo = async () => {
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
    if (data.state === 'finished') {
      finishedModal.value = true
    }
    gameStarted.value = data.state !== 'waiting'
    pickStarted.value = data.pickState === 'picking'
    pickExhausted.value = data.pickState === 'exhausted'
    qrCodeShowed.value = data.qrCodeVisible
  }
}

const clickGameStart = () => {
  isDisabled.value = true
  startModal.value = true
}

const clickGameFinish = () => {
  isDisabled.value = true
  finishModal.value = true
}

const clickStartModal = async () => {
  const { error } = await apiClient.POST('/api/rooms/{roomId}/control/start', {
    params: {
      path: {
        roomId,
      },
    },
  })
  if (error) {
    errorMessage.value = error.message
  }
  await getLatestInfo()
  startModal.value = false
  isDisabled.value = false
}

const clickFinishModal = async () => {
  const { error } = await apiClient.POST('/api/rooms/{roomId}/control/finish', {
    params: {
      path: {
        roomId,
      },
    },
  })
  if (error) {
    errorMessage.value = error.message
  }
  await getLatestInfo()
  finishModal.value = false
  isDisabled.value = false
}

const clickPickStart = async () => {
  isDisabled.value = true
  const { error } = await apiClient.POST('/api/rooms/{roomId}/control/pick/start', {
    params: {
      path: {
        roomId,
      },
    },
  })
  if (error) {
    errorMessage.value = error.message
  }
  await getLatestInfo()
  isDisabled.value = false
}

const clickPickFinish = async () => {
  isDisabled.value = true
  const { error } = await apiClient.POST('/api/rooms/{roomId}/control/pick/finish', {
    params: {
      path: {
        roomId,
      },
    },
  })
  if (error) {
    errorMessage.value = error.message
  }
  await getLatestInfo()
  isDisabled.value = false
}

const clickPickCancel = async () => {
  isDisabled.value = true
  const { error } = await apiClient.POST('/api/rooms/{roomId}/control/pick/cancel', {
    params: {
      path: {
        roomId,
      },
    },
  })
  if (error) {
    errorMessage.value = error.message
  }
  await getLatestInfo()
  isDisabled.value = false
}

const clickShowQr = async () => {
  isDisabled.value = true
  const { error } = await apiClient.POST('/api/rooms/{roomId}/control/qrcode/show', {
    params: {
      path: {
        roomId,
      },
    },
  })
  if (error) {
    errorMessage.value = error.message
  }
  await getLatestInfo()
  isDisabled.value = false
}

const clickHideQr = async () => {
  isDisabled.value = true
  const { error } = await apiClient.POST('/api/rooms/{roomId}/control/qrcode/hide', {
    params: {
      path: {
        roomId,
      },
    },
  })
  if (error) {
    errorMessage.value = error.message
  }
  await getLatestInfo()
  isDisabled.value = false
}
</script>

<template>
  <UModal
    v-model:open="startModal"
    :dismissible="false"
    title="本当にゲームを開始しますか？"
    description="ゲームを一度開始すると、以降、新たにプレイヤーがゲームに参加することはできません。"
    :ui="{ footer: 'justify-end' }"
    :close="false"
  >
    <template #footer>
      <UButton
        label="やっぱ開始しない"
        color="neutral"
        variant="outline"
        @click="
          startModal = false
          isDisabled = false
        "
      />
      <UButton label="本当に開始する！" color="neutral" @click="clickStartModal()" />
    </template>
  </UModal>

  <UModal
    v-model:open="finishModal"
    :dismissible="false"
    title="本当にゲームを終了しますか？"
    description="ゲームを一度終了すると、再開することはできません。新たな球の抽選やチャットへの投稿はできなくなります。"
    :ui="{ footer: 'justify-end' }"
    :close="false"
  >
    <template #footer>
      <UButton
        label="やっぱ終了しない"
        color="neutral"
        variant="outline"
        @click="
          finishModal = false
          isDisabled = false
        "
      />
      <UButton label="本当に終了する！" color="neutral" @click="clickFinishModal()" />
    </template>
  </UModal>

  <UModal
    v-model:open="finishedModal"
    :dismissible="false"
    title="このゲームは終了しました。"
    :ui="{ footer: 'justify-end' }"
    :close="false"
  >
    <template #footer>
      <UButton label="トップへ戻る" color="neutral" to="/" />
    </template>
  </UModal>

  <UContainer class="pt-6">
    <div v-if="loading">読み込み中...</div>
    <div v-else-if="errorMessage">
      {{ errorMessage }}　ゲームの操作を続けるには再読み込みしてください。
    </div>
    <template v-else>
      <h2 class="text-3xl font-bold mb-6">ルーム{{ props.roomCode }}の操作</h2>

      <div class="grid h-[70dvh] grid-rows-5 overflow-hidden gap-4">
        <template v-if="!gameStarted">
          <UButton
            variant="solid"
            class="row-span-4 min-h-0 h-full w-full overflow-hidden rounded-2xl text-6xl font-extrabold leading-none grid place-items-center"
            @click="clickGameStart()"
            :disabled="isDisabled"
          >
            <span class="flex flex-col items-center justify-center gap-4">
              <UIcon name="i-lucide-power" class="size-32" />
              <div>
                <span class="whitespace-nowrap">ゲーム</span>
                <span class="whitespace-nowrap">を</span>
                <span class="whitespace-nowrap">開始</span>
              </div>
            </span>
          </UButton>
        </template>

        <template v-else>
          <template v-if="!pickStarted">
            <UButton
              variant="solid"
              class="row-span-3 min-h-0 h-full w-full overflow-hidden rounded-2xl text-6xl font-extrabold leading-none grid place-items-center"
              @click="clickPickStart()"
              v-if="!pickExhausted"
              :disabled="isDisabled"
            >
              <span class="flex flex-col items-center justify-center gap-4">
                <UIcon name="i-lucide-circle-play" class="size-32" />
                <div>
                  <span class="whitespace-nowrap">抽選</span>
                  <span class="whitespace-nowrap">を</span>
                  <span class="whitespace-nowrap">開始</span>
                </div>
              </span>
            </UButton>
            <UButton
              variant="solid"
              class="row-span-3 min-h-0 h-full w-full overflow-hidden rounded-2xl text-6xl font-extrabold leading-none grid place-items-center"
              disabled
              v-else
            >
              <span class="flex flex-col items-center justify-center gap-4">
                <UIcon name="i-lucide-x" class="size-32" />
                <div>
                  <span class="whitespace-nowrap">抽選</span>
                  <span class="whitespace-nowrap">を</span>
                  <span class="whitespace-nowrap">開始</span>
                </div>
              </span>
            </UButton>
          </template>

          <template v-else>
            <UButton
              variant="solid"
              class="row-span-2 min-h-0 h-full w-full overflow-hidden rounded-2xl text-[32px] font-extrabold leading-none grid place-items-center"
              @click="clickPickFinish()"
              :disabled="isDisabled"
            >
              <span class="flex flex-col items-center justify-center gap-4">
                <UIcon name="i-lucide-circle-pause" class="size-24" />
                <div>
                  <span class="whitespace-nowrap">抽選</span>
                  <span class="whitespace-nowrap">を</span>
                  <span class="whitespace-nowrap">終了</span>
                </div>
              </span>
            </UButton>

            <UButton
              variant="outline"
              class="min-h-0 h-full w-full overflow-hidden rounded-2xl text-[32px] font-extrabold leading-none grid place-items-center"
              @click="clickPickCancel()"
              :disabled="isDisabled"
            >
              <span class="flex flex-col items-center justify-center gap-4">
                <div>
                  <span class="whitespace-nowrap">抽選</span>
                  <span class="whitespace-nowrap">を</span>
                  <span class="whitespace-nowrap">キャンセル</span>
                </div>
              </span>
            </UButton>
          </template>

          <UButton
            variant="outline"
            class="min-h-0 h-full w-full overflow-hidden rounded-2xl text-[32px] font-extrabold leading-none grid place-items-center"
            @click="clickGameFinish()"
            :disabled="isDisabled"
          >
            <span class="flex flex-col items-center justify-center gap-4">
              <div>
                <span class="whitespace-nowrap">ゲーム</span>
                <span class="whitespace-nowrap">を</span>
                <span class="whitespace-nowrap">終了</span>
              </div>
            </span>
          </UButton>
        </template>

        <template v-if="qrCodeShowed">
          <UButton
            variant="soft"
            class="min-h-0 h-full w-full overflow-hidden rounded-2xl text-[32px] font-extrabold leading-none grid place-items-center"
            @click="clickHideQr()"
            :disabled="isDisabled"
          >
            <span class="flex flex-col items-center justify-center gap-4">
              <div>
                <UIcon
                  name="i-lucide-qr-code"
                  class="size-8 inline-flex shrink-0 translate-y-[-0.1em] mr-1"
                />
                <span class="whitespace-nowrap">QR</span>
                <span class="whitespace-nowrap">コード</span>
                <span class="whitespace-nowrap">を</span>
                <span class="whitespace-nowrap">非表示</span>
              </div>
            </span>
          </UButton>
        </template>
        <template v-else>
          <UButton
            variant="soft"
            class="min-h-0 h-full w-full overflow-hidden rounded-2xl text-[32px] font-extrabold leading-none grid place-items-center"
            @click="clickShowQr()"
            :disabled="isDisabled"
          >
            <span class="flex flex-col items-center justify-center gap-4">
              <div>
                <UIcon
                  name="i-lucide-qr-code"
                  class="size-8 inline-flex shrink-0 translate-y-[-0.1em] mr-1"
                />
                <span class="whitespace-nowrap">QR</span>
                <span class="whitespace-nowrap">コード</span>
                <span class="whitespace-nowrap">を</span>
                <span class="whitespace-nowrap">表示</span>
              </div>
            </span>
          </UButton>
        </template>
      </div>
    </template>
  </UContainer>
</template>
