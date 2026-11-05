import { http } from '../../http'
import { currentUser, sendEvent } from './core'
import { mockRooms } from './fixtures/index'

export const participantHandlers = [
  http.post('/api/rooms/{roomId}/participants', ({ params, request, response }) => {
    const room = mockRooms.find((room) => room.roomId === params.roomId)
    if (!room) return response(404).json({ message: 'Room not found' })
    const participant = currentUser(request)
    if (!room.participants.some((item) => item.user.userId === participant.userId)) {
      const joinedAt = new Date().toISOString()
      room.participants.push({ user: participant, joinedAt })
      sendEvent(
        {
          type: 'ParticipantJoined',
          body: { participant, participantCount: room.participants.length, joinedAt },
        },
        room.roomId,
      )
    }
    return response(204).empty()
  }),
]
