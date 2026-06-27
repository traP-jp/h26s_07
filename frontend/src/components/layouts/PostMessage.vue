<script setup lang="ts">
import { ref } from 'vue'
const room = defineProps<{ roomId: string }>()
const newMessage = ref('')
type RequestBody = {
  content: string
}
const post = async () => {
  const data: RequestBody = {
    content: newMessage.value,
  }
  try {
    const response = await fetch(`/api/rooms/${room.roomId}/chats`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(data),
    })
    if (!response.ok) {
      throw new Error(`HTTP Error: ${response.status}`)
    }
    const result = await response.json()
    console.log('Success', result)
  } catch (error) {
    console.error('Error', error)
  }
}
</script>

<template>
  <div class="wrapper">
    <div class="newMessage">
      <input id="newMessage" v-model="newMessage" type="text" />
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
  gap: 6px;
  align-items: center;
  padding-left: 8px;
  padding-top: 8px;
  padding-right: 8px;
}
.newMessage {
  height: 20px;
  width: 80%;
}
.button {
  width: 20px;
  height: 20px;
}
path {
  fill: rgb(12, 185, 80);
}
#newMessage {
  height: 30px;
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
</style>
