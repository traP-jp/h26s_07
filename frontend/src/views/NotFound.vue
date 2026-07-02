<template>
  <UContainer class="not-found">
    <h1 class="status-code">404 Not Found</h1>

    <section class="message-card">
      <h2 class="message-title">ページを表示できません</h2>
      <p class="message-text">{{ errorMessage }}</p>
    </section>

    <UButton to="/" color="primary" size="xl">Homeに戻る</UButton>
  </UContainer>
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

<style scoped>
.not-found {
  display: flex;
  height: 100vh;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  background: #f8fbff;
  text-align: center;
}

.status-code {
  color: #1476d4;
  font-size: 6rem;
  font-weight: 800;
}

.message-card {
  width: 60%;
  margin-bottom: 1.5rem;
  padding: 3%;
  border: 4px solid #e5e7eb;
  border-radius: 0.75rem;
  background: #fff;
}

.message-title {
  margin-bottom: 0.75rem;
  color: #1f2937;
  font-size: 1.875rem;
  font-weight: 700;
}

.message-text {
  color: #6b7280;
  font-weight: 700;
}
</style>
