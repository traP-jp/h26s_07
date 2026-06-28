import { createApp } from 'vue'
import { createPinia } from 'pinia'
import ui from '@nuxt/ui/vue-plugin'

import './assets/css/main.css'
import App from './App.vue'
import router from './router'

// mockを起動する
if (import.meta.env.DEV && import.meta.env.VITE_API_MOCK === 'true') {
  const { worker } = await import('./mocks/server')
  await worker.start({ onUnhandledRequest: 'warn' })
}

const app = createApp(App)
const pinia = createPinia()

app.use(pinia)
app.use(router)
app.use(ui)

await router.isReady()

app.mount('#app')
