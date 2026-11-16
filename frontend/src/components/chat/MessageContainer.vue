<script setup lang="ts">
import type { UserId } from '@/api/schema'
import UserIcon from '@/components/common/UserIcon.vue'

defineProps<{
  userId: UserId
  content: string
  notificationType?: 'bingo' | 'reach' | 'exhausted' | 'notice'
}>()
</script>

<template>
  <div class="message" :class="{ 'message--system': userId === '' }">
    <UserIcon v-if="userId !== ''" :user-id="userId" class="icon"></UserIcon>
    <div v-if="userId !== ''" class="message__body">
      <span class="message__author">{{ userId }}</span>
      <div class="nakami">{{ content }}</div>
    </div>
    <div
      v-else
      class="nakami special"
      :class="{
        'special--bingo': notificationType === 'bingo',
        'special--reach': notificationType === 'reach',
        'special--exhausted': notificationType === 'exhausted',
        'special--notice': notificationType === 'notice',
      }"
      role="status"
    >
      {{ content }}
    </div>
  </div>
</template>

<style scoped>
.message {
  display: flex;
  align-items: flex-start;
  gap: 9px;
  padding: 9px 12px 2px;
}
.icon {
  width: 30px;
  height: 30px;
  margin-top: 17px;
  flex-shrink: 0;
  border-radius: 50%;
  overflow: hidden;
}
.message__body {
  min-width: 0;
  max-width: 82%;
}
.message__author {
  display: block;
  margin: 0 0 3px 2px;
  color: #56616b;
  font-size: 11px;
  font-weight: 700;
  line-height: 1.2;
  overflow-wrap: anywhere;
}
.nakami.special {
  width: auto;
  max-width: 100%;
  margin: 4px auto;
  padding: 8px 12px;
  border: 1px solid #d8d1c4;
  background: #f5f0e5;
  color: #4a4b45;
  text-align: center;
  border-radius: 999px;
  font-weight: 650;
  font-size: 12px;
  line-height: 1.4;
  box-shadow: none;
}
.nakami.special--bingo {
  border-color: #ddc77c;
  background: #f5ebc8;
  color: #654d19;
  font-weight: 800;
}
.nakami.special--reach {
  border-color: #d8b1a5;
  background: #f7e9e4;
  color: #78433b;
  font-weight: 600;
}
.nakami.special--exhausted {
  border-color: #a9bcb4;
  background: #eaf0ed;
  color: #315447;
  font-weight: 750;
}
.nakami.special--notice {
  border-color: #dedbd4;
  background: #f4f3ef;
  color: #535b61;
}
.nakami {
  width: fit-content;
  max-width: 100%;
  padding: 8px 11px;
  background: #ffffff;
  border: 1px solid #e7e1d7;
  border-radius: 3px 12px 12px 12px;
  color: #28343d;
  line-height: 1.5;
  white-space: pre-wrap;
  word-break: break-word;
  box-shadow: 0 1px 2px rgb(39 42 46 / 5%);
}
</style>
