import { fileURLToPath, URL } from 'node:url'

import { defineConfig, loadEnv } from 'vite'
import vue from '@vitejs/plugin-vue'
import ui from '@nuxt/ui/vite'
import vueDevTools from 'vite-plugin-vue-devtools'

// https://vite.dev/config/
export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '')

  return {
    plugins: [
      vue(),
      ui({
        colorMode: false,
        ui: {
          colors: {
            primary: 'rose',
            secondary: 'purple',
            tertiary: 'indigo',
          },
        },
      }),
      vueDevTools(),
    ],
    resolve: {
      alias: {
        '@': fileURLToPath(new URL('./src', import.meta.url)),
      },
    },
    server:
      env.VITE_API_MOCK === 'true'
        ? undefined
        : {
            proxy: {
              '/api': {
                target: env.VITE_BACKEND_BASE_URL || 'http://localhost:8080',
                changeOrigin: true,
                ws: true,
                configure: (proxy) => {
                  proxy.on('proxyReq', (proxyReq) => {
                    const traqId = env.MY_TRAQ_ID
                    if (traqId) {
                      proxyReq.setHeader('X-Forwarded-User', traqId)
                    }
                  })
                },
              },
            },
          },
  }
})
