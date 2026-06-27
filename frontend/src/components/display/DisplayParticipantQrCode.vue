<script setup lang="ts">
import { computed } from 'vue'
import { useQRCode } from '@vueuse/integrations/useQRCode'

const props = defineProps<{
  roomCode: string
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
    default-open
    :close="false"
    :dismissible="false"
    :ui="{ overlay: 'bg-elevated/25', content: 'w-[40vw] p-8' }"
  >
    <template #content>
      <div class="display-participant-qr">
        <p class="display-participant-qr__title">参加はこちら</p>
        <img class="display-participant-qr__image" :src="qrCode" alt="参加画面へのQRコード" />
        <p class="display-participant-qr__url">{{ participantUrl }}</p>
      </div>
    </template>
  </UModal>
</template>

<style scoped>
.display-participant-qr {
  display: grid;
  justify-items: center;
  gap: 16px;
}

.display-participant-qr__title {
  font-size: 32px;
  font-weight: 900;
  color: #173f75;
}

.display-participant-qr__image {
  width: 80%;
}

.display-participant-qr__url {
  width: 100%;
  margin: 0;
  font-size: 1em;
  font-weight: 700;
  text-align: center;
  color: #395875;
}
</style>
