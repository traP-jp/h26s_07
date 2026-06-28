<script setup lang="ts">
import { computed } from 'vue'
import { useQRCode } from '@vueuse/integrations/useQRCode'

const props = defineProps<{
  roomCode: string
  open: boolean
}>()

const participantUrl = computed(() => {
  return `${window.location.origin}/${props.roomCode}/participant`
})

const qrCode = useQRCode(participantUrl, {
  errorCorrectionLevel: 'M',
  margin: 2,
  width: 520,
})
</script>

<template>
  <UModal
    :open="props.open"
    :overlay="false"
    :close="false"
    :dismissible="false"
    fullscreen
    :ui="{
      content: 'w-full h-full bg-transparent  p-8',
    }"
  >
    <template #content>
      <div class="display-participant-qr">
        <div class="display-participant-qr__panel">
          <p class="display-participant-qr__title">参加はこちら</p>
          <img class="display-participant-qr__image" :src="qrCode" />
          <p class="display-participant-qr__url">{{ participantUrl }}</p>
        </div>
      </div>
    </template>
  </UModal>
</template>

<style scoped>
.display-participant-qr {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: flex-start;
  justify-content: flex-start;
}

.display-participant-qr__panel {
  width: 34%;
  aspect-ratio: 1 / 1;
  height: auto;
  display: grid;
  grid-template-rows: auto 1fr auto;
  justify-items: center;
  border-radius: 12px;
  background: rgb(255 255 255 / 0.92);
  box-shadow: 0 12px 32px rgb(14 39 78 / 0.16);
}

.display-participant-qr__title {
  margin-top: 16px;
  font-size: 24px;
  font-weight: 900;
  color: #173f75;
}

.display-participant-qr__image {
  width: 67%;
}

.display-participant-qr__url {
  width: 100%;
  margin-bottom: 24px;
  font-size: 0.8em;
  font-weight: 700;
  text-align: center;
  color: #395875;
}
</style>
