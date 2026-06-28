<template>
  <UError
    :error="error"
    :clear="false"
    class="min-h-screen bg-[#f8fbff] px-4 text-center text-[#17324d]"
    :ui="{
      root: 'grid content-center justify-items-center',
      statusCode: 'text-6xl font-extrabold leading-none text-[#1476d4] md:text-8xl',
      statusMessage: 'mt-6 text-2xl font-bold text-gray-800 md:text-3xl',
      message: 'mt-3 max-w-155 text-base font-bold leading-7 text-gray-500',
      links: 'mt-8',
    }"
  >
    <template #statusCode>404 Not Found</template>

    <template #links>
      <UButton to="/" color="primary" variant="solid" size="xl">Homeに戻る</UButton>
    </template>
  </UError>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = withDefaults(
  defineProps<{
    errorMessage?: string
  }>(),
  {
    errorMessage: 'ページが見つかりませんでした。',
  },
)

const error = computed(() => ({
  statusCode: 404,
  statusMessage: 'ページを表示できません',
  message: props.errorMessage,
}))
</script>
