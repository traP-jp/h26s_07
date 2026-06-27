<script setup lang="ts">
import { ref } from 'vue'
interface Message {
  messageId: string
  userId: string
  content: string
}
const messages = ref<Message[]>([
  { messageId: '0', userId: '', content: '球がなくなりました！' },
  { messageId: '1', userId: 'HokubuRailway', content: 'おい' },
  { messageId: '5', userId: 'BOT_kinano', content: 'おい！' },
])
</script>

<template>
  <div>
    <div v-for="message in messages" :key="message.messageId">
      <Message :user-id="message.userId" :content="message.content"></Message>
    </div>
  </div>
</template>
