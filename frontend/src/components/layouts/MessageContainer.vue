<script setup lang="ts">
import type { UserId } from '@/api/schema'
defineProps<{
  userId: UserId
  content: string
  notificationType?: 'bingo' | 'reach' | 'notice'
}>()
</script>

<template>
  <div class="message">
    <UserIcon v-if="userId != ''" :user-id="userId" class="icon"></UserIcon>
    <div
      class="nakami"
      :class="{
        special: userId == '',
        'special--bingo': notificationType === 'bingo',
        'special--reach': notificationType === 'reach',
        'special--notice': notificationType === 'notice',
      }"
    >
      {{ content }}
    </div>
  </div>
</template>

<style scoped>
.message {
  display: flex;
  align-items: start;
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
  text-align: center;
  border-radius: 10px;
  font-weight: 600;
  font-size: 1.4em;
  line-height: 1;
}
.nakami.special--bingo {
  background: #3972b8;
  color: #ffffff;
}
.nakami.special--reach {
  background: #ffffff;
  color: #185aa9;
}
.nakami.special--notice {
  border-color: #f1d529;
  background: #fff2a8;
  color: #37506f;
}
.nakami {
  max-width: 75%;
  padding: 10px 14px;
  background: #ffffff;
  border: 1px solid rgb(56 114 177 / 0.28);
  border-radius: 0 14px 14px 14px;
  color: #24364d;
  line-height: 1.6;
  white-space: pre-wrap;
  word-break: break-word;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.08);
}
</style>
