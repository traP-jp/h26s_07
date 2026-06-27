import type { GameSettings, UpdateGameSettingsRequest } from '@/api/schema'

import { http } from '../../http'
import { readJson } from './core'

const mockSettings: GameSettings = {
  name: 'デモビンゴ',
  description: 'モック API で動かす待機中のビンゴルームです。',
  admins: [{ userId: 'mumumu' }],
}

export const settingHandlers = [
  http.put('/api/rooms/{roomId}/settings', async ({ request, response }) => {
    const body = await readJson<UpdateGameSettingsRequest>(request)
    const settings: GameSettings =
      body === undefined
        ? mockSettings
        : {
            name: body.settings.name,
            description: body.settings.description,
            admins: (body.settings.adminUserIds ?? ['mumumu']).map((userId) => ({ userId })),
          }

    return response(200).json(settings)
  }),
]
