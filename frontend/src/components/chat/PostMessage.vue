<script setup lang="ts">
import { computed, ref, useId } from 'vue'
import { apiClient } from '@/api/apiClient'

const room = withDefaults(
  defineProps<{
    roomId: string
    variant?: 'default' | 'light'
  }>(),
  { variant: 'default' },
)

const newMessage = ref('')
const sending = ref(false)
const sendError = ref('')
const inputId = useId()
const canSend = computed(
  () => newMessage.value.trim().length > 0 && newMessage.value.length <= 500 && !sending.value,
)

const post = async () => {
  if (!canSend.value || sending.value) return
  sending.value = true
  sendError.value = ''
  try {
    const response = await apiClient.POST('/api/rooms/{roomId}/chats', {
      params: { path: { roomId: room.roomId } },
      body: { content: newMessage.value },
    })
    if (response.error) throw new Error(`HTTP Error: ${response.error}`)
    newMessage.value = ''
  } catch (error) {
    console.error('Error', error)
    sendError.value = '送信できませんでした。時間をおいて再度お試しください。'
  } finally {
    sending.value = false
  }
}
</script>

<template>
  <form
    class="wrapper"
    :class="{ 'wrapper--light': room.variant === 'light' }"
    @submit.prevent="post"
  >
    <label class="sr-only" :for="inputId">メッセージ</label>
    <input
      :id="inputId"
      v-model="newMessage"
      class="message-input"
      type="text"
      maxlength="500"
      autocomplete="off"
      placeholder="メッセージを書く"
      @input="sendError = ''"
    />
    <span class="message-count" :class="{ 'message-count--limit': newMessage.length > 450 }">
      {{ newMessage.length }}/500
    </span>
    <button
      class="send-button"
      type="submit"
      :disabled="!canSend"
      :aria-label="sending ? '送信中' : '送信'"
    >
      <svg aria-hidden="true" height="20" width="20" viewBox="0 0 24 24">
        <path d="M2.2 11.1 21 2.7l-5.9 18.6-4.1-7.2-8.8-3Zm9.8 1.3 3.9 2.5 2.7-8.4-6.6 5.9Z"></path>
      </svg>
    </button>
    <p v-if="sendError" class="send-error" role="status">{{ sendError }}</p>
  </form>
</template>

<style scoped>
.wrapper {
  display: flex;
  flex-wrap: wrap;
  box-sizing: border-box;
  min-height: 54px;
  flex: 0 0 auto;
  align-items: center;
  gap: 8px;
  padding: 8px 10px;
  border-top: 1px solid rgb(35 63 105 / 0.12);
  background: #fff;
}

.sr-only {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border: 0;
}

.message-input {
  box-sizing: border-box;
  width: 100%;
  min-width: 0;
  height: 44px;
  flex: 1 1 auto;
  padding: 8px 11px;
  border: 1px solid #d6d2c9;
  border-radius: 8px;
  background: #faf9f6;
  color: #28343d;
  font: inherit;
  font-size: 14px;
  outline: none;
  transition:
    border-color 140ms ease,
    box-shadow 140ms ease,
    background-color 140ms ease;
}

.message-input::placeholder {
  color: #79828a;
}
.message-input:hover {
  border-color: #aaa69d;
}
.message-input:focus-visible {
  border-color: #8c4338;
  background: #fff;
  box-shadow: 0 0 0 3px rgb(140 67 56 / 16%);
}

.message-count {
  flex: 0 0 auto;
  color: #626c69;
  font-size: 10px;
  font-variant-numeric: tabular-nums;
}
.message-count--limit {
  color: #9b443b;
}

.send-error {
  flex: 0 0 100%;
  margin: 0;
  color: #9b443b;
  font-size: 12px;
  line-height: 1.35;
}

.send-button {
  display: grid;
  width: 44px;
  height: 44px;
  flex: 0 0 auto;
  place-items: center;
  border: 1px solid #8d382e;
  border-radius: 8px;
  background: #a9362c;
  color: white;
  cursor: pointer;
  transition:
    background-color 130ms ease,
    transform 130ms ease,
    opacity 130ms ease;
}
.send-button svg {
  fill: currentColor;
}
.send-button:hover:not(:disabled) {
  background: #842d26;
}
.send-button:active:not(:disabled) {
  transform: translateY(1px) scale(0.97);
}
.send-button:focus-visible {
  outline: 3px solid rgb(140 67 56 / 25%);
  outline-offset: 2px;
}
.send-button:disabled {
  opacity: 0.42;
  cursor: not-allowed;
}

.wrapper:not(.wrapper--light) {
  border-color: rgb(255 255 255 / 0.12);
  background: #20262a;
}
.wrapper:not(.wrapper--light) .message-input {
  border-color: #485057;
  background: #2b3338;
  color: #f6f2e9;
}
.wrapper:not(.wrapper--light) .message-input::placeholder {
  color: #b1b8b7;
}
.wrapper:not(.wrapper--light) .message-count {
  color: #bac0be;
}

@media (prefers-reduced-motion: reduce) {
  .message-input,
  .send-button {
    transition: none;
  }
}
</style>
