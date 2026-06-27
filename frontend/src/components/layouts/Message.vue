<script setup lang="ts">
import type { UserId } from '@/api/schema'
defineProps<{
  userId: UserId
  content: string
}>()
</script>

<template>
  <div class="message">
    <UserIcon v-if="userId != ''" :user-id="userId" class="icon"></UserIcon>
    <div class="nakami" :class="{ special: userId == '' }">
      {{ content }}
    </div>
  </div>
</template>

<style scoped>
.message {
  display: flex;
  align-items: center;
  gap: 8px;
  padding-left: 8px;
  padding-top: 8px;
  padding-right: 8px;
}
.icon {
  width: 36px;
  height: 36px;
  margin-right: 6px;
  flex-shrink: 0;
  border-radius: 18px;
  overflow: hidden;
}
.nakami.special {
  width: 100%;
  max-width: 100%;
  background: rgb(250, 226, 22);
  text-align: center;
  border-radius: 10px;
  color: #272727;
  font-weight: 600;
  font-size: 1.4em;
  line-height: 1;
}
.nakami {
  max-width: 75%;
  padding: 10px 14px;
  background: #424242;
  border: 1px solid rgb(12, 54, 203);
  border-radius: 14px;
  color: #cfcfcf;
  line-height: 1.6;
  white-space: pre-wrap;
  word-break: break-word;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.08);
}
</style>
