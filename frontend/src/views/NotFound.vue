<template>
  <div
    class="grid min-h-screen content-center justify-items-center gap-6 bg-[#f8fbff] px-4 py-6 text-center text-[#17324d]"
  >
    <h1 class="text-6xl font-extrabold leading-none text-[#1476d4] md:text-8xl">404 Not Found</h1>
    <UCard
      class="w-full max-w-155 border-4 border-gray-200 bg-white"
      variant="outline"
      :ui="{ body: 'p-8 sm:p-8' }"
    >
      <h2 class="mb-3 text-2xl font-bold text-gray-800 md:text-3xl">ページを表示できません</h2>
      <p class="text-base font-bold leading-7 text-gray-500">{{ errorMessage }}</p>
    </UCard>
    <UButton to="/" color="primary" variant="solid" size="xl">Homeに戻る</UButton>
  </div>
</template>

<script setup lang="ts">
withDefaults(
  defineProps<{
    errorMessage?: string
  }>(),
  {
    errorMessage: 'ページが見つかりませんでした。',
  },
)
</script>
