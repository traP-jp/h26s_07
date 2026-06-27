<script setup lang="ts">
import { ref } from 'vue'
import type { Uuid, Message, DateTime } from '@/api/schema'
const room = defineProps<{ roomId: string; textarea: boolean }>()
const messages = ref<Message[]>([])

const addUserMessage = (m: Message) => {
  messages.value.push(m)
}

const addSpecialMessage = (id: Uuid, content: string, createdAt: DateTime) => {
  messages.value.push({
    messageId: id,
    content: content,
    author: { userId: '' },
    createdAt: createdAt,
  })
}

defineExpose({ addUserMessage, addSpecialMessage })
</script>

<template>
  <div>
    <div v-for="message in messages" :key="message.messageId">
      <Message :user-id="message.author.userId" :content="message.content"></Message>
    </div>
  </div>
  <div v-if="room.textarea">
    <PostMessage :room-id="room.roomId"></PostMessage>
  </div>
</template>
