import { http } from '../../http'
import { sendEvent } from './core'

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
    sendEvent({ type: 'ShowQRCode', body: {} })
    return response(204).empty()
  }),

  http.post('/api/rooms/{roomId}/control/qrcode/hide', ({ response }) => {
    sendEvent({ type: 'HideQRCode', body: {} })
    return response(204).empty()
  }),
]
