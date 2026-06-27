import { http } from '../../http'

export const participantHandlers = [
  http.post('/api/rooms/{roomId}/participants', ({ response }) => {
    return response(204).empty()
  }),
]
