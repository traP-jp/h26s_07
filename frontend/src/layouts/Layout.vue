<template>
  <div>
    <Header />
    <UMain>
      <RouterView />
    </UMain>
  </div>
</template>

<script setup lang="ts">
import Header from '@/components/layouts/Header.vue'
</script>
