import { createRouter, createWebHistory } from 'vue-router'
import type { RouteLocationNormalized } from 'vue-router'

import type { RoomCode } from '@/api/schema'
import DefaultLayout from '@/layouts/Layout.vue'
import Home from '@/views/Home.vue'
import { useRoomsStore } from '@/stores/rooms'

async function validateDisplayRoomCode(to: RouteLocationNormalized) {
  const roomCode = to.params.roomCode

  if (typeof roomCode !== 'string') {
    return {
      name: 'NotFound',
      query: { message: '存在しないroomcodeです。' },
    }
  }

  const roomsStore = useRoomsStore()
  const room = await roomsStore.getRoomByCode(roomCode as RoomCode)

  if (!room) {
    return {
      name: 'NotFound',
      query: { message: '存在しないroomcodeです。' },
    }
  }

  return true
}

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      component: DefaultLayout,
      children: [
        {
          path: '',
          name: 'Home',
          component: Home,
        },
        {
          path: '/rurun',
          name: 'rurun',
          component: () => import('@/views/Rurun.vue'),
        },
        {
          path: '/:roomCode',
          redirect: (to) => `/${String(to.params.roomCode)}/participants`,
        },
        {
          path: '/:pathMatch(.*)*',
          redirect: {
            name: 'NotFound',
          },
        },
      ],
    },
    {
      path: '/:roomCode/display',
      name: 'display',
      component: () => import('@/views/Display.vue'),
      beforeEnter: validateDisplayRoomCode,
    },
    {
      path: '/404',
      name: 'NotFound',
      component: () => import('@/views/NotFound.vue'),
      props: (route) => ({
        errorMessage:
          typeof route.query.message === 'string'
            ? route.query.message
            : 'ページが見つかりませんでした。',
      }),
    },
  ],
})

export default router
