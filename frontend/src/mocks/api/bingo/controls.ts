import { http } from '../../http'

export const controlHandlers = [
  http.post('/api/rooms/{roomId}/control/start', ({ response }) => {
    return response(204).empty()
  }),

  http.post('/api/rooms/{roomId}/control/finish', ({ response }) => {
    return response(204).empty()
  }),

  http.post('/api/rooms/{roomId}/control/pick/start', ({ response }) => {
    return response(204).empty()
  }),

  http.post('/api/rooms/{roomId}/control/pick/cancel', ({ response }) => {
    return response(204).empty()
  }),

  http.post('/api/rooms/{roomId}/control/pick/finish', ({ response }) => {
    return response(204).empty()
  }),

  http.post('/api/rooms/{roomId}/control/qrcode/show', ({ response }) => {
    return response(204).empty()
  }),

  http.post('/api/rooms/{roomId}/control/qrcode/hide', ({ response }) => {
    return response(204).empty()
  }),
]
