<script setup lang="ts">
import { ref } from 'vue'
import { apiClient } from '@/api/apiClient'
const room = withDefaults(
  defineProps<{
    roomId: string
    variant?: 'default' | 'light'
  }>(),
  {
    variant: 'default',
  },
)

const newMessage = ref('')
type RequestBody = {
  content: string
}
const post = async () => {
  if (newMessage.value.length == 0) return

  const data: RequestBody = {
    content: newMessage.value,
  }
  if (newMessage.value.length > 500) {
    throw new Error('Error 400 : Message Invalid')
  }
  try {
    const response = await apiClient.POST('/api/rooms/{roomId}/chats', {
      params: { path: { roomId: room.roomId } },
      body: data,
    })
    if (response.error) {
      throw new Error(`HTTP Error: ${response.error}`)
    } else {
      newMessage.value = ''
    }
  } catch (error) {
    console.error('Error', error)
  }
}
</script>

<template>
  <div class="wrapper" :class="{ 'wrapper--light': room.variant === 'light' }">
    <div class="newMessage">
      <input
        id="newMessage"
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
