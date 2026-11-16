<script setup lang="ts">
import { computed, ref } from 'vue'
import { apiClient } from '@/api/apiClient'

const { roomId, variant = 'default' } = defineProps<{
  roomId: string
  variant?: 'default' | 'light'
}>()

const newMessage = ref('')
const sending = ref(false)
const sendError = ref('')
const messageLength = computed(() => Array.from(newMessage.value).length)
const canSend = computed(
  () => newMessage.value.trim().length > 0 && messageLength.value <= 500 && !sending.value,
)

const post = async () => {
  if (!canSend.value || sending.value) return
  sending.value = true
  sendError.value = ''
  try {
    const response = await apiClient.POST('/api/rooms/{roomId}/chats', {
      params: { path: { roomId } },
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
  <div class="wrapper" :class="{ 'wrapper--light': variant === 'light' }">
    <div class="newMessage">
      <input
        v-model="newMessage"
        type="text"
        placeholder="1 文字以上 500 文字以下で入力"
        @keydown.enter.prevent="post()"
      />
    </div>
    <div class="button" @click="post()">
      <svg height="30" width="30" view-box="0 0 30 30">
        <path d="M0 0v12.75l30 2.25-30 2.25v12.75l30-15z"></path>
      </svg>
    </div>
    <p v-if="sendError" role="status">{{ sendError }}</p>
  </div>
</template>

<style scoped>
.wrapper {
  display: flex;
  box-sizing: border-box;
  height: 48px;
  flex: 0 0 48px;
  gap: 6px;
  align-items: center;
  padding: 8px;
  border-top: 2px solid #cfcfcf;
}
.newMessage {
  height: 32px;
  min-width: 0;
  flex: 1 1 auto;
}
.button {
  display: flex;
  width: 32px;
  height: 32px;
  flex: 0 0 auto;
  align-items: center;
  justify-content: center;
  cursor: pointer;
}
path {
  fill: rgb(12, 185, 80);
}
#newMessage {
  box-sizing: border-box;
  height: 32px;
  color: #cfcfcf;
  width: 100%;
  padding: 8px 10px;
  border: none;
  border-radius: 7px;
  background-color: #272727;
  font-size: 1em;
  line-height: 1.5;
}
#newMessage:focus {
  outline: 1px solid rgb(12, 185, 80);
}

.wrapper--light {
  border-top-color: rgb(35 63 105 / 0.1);
  background: #ffffff;
}

.wrapper--light #newMessage {
  border: 1px solid rgb(35 63 105 / 0.16);
  background-color: #f8fafc;
  color: #1f3556;
}

.wrapper--light #newMessage::placeholder {
  color: #8a9ab0;
}
</style>
