import type { Message } from '@/api/schema'

export const mockMessages: Message[] = [
  {
    messageId: '00000000-0000-4000-8000-000000000101',
    content: 'モックルームへようこそ',
    author: { userId: 'mumumu' },
    createdAt: '2026-06-27T10:02:00.000Z',
  },
  {
    messageId: '00000000-0000-4000-8000-000000000102',
    content: '準備できています',
    author: { userId: 'rurun' },
    createdAt: '2026-06-27T10:03:00.000Z',
  },
]

const chatScrollSamples = [
  { author: { userId: 'mumumu' }, content: 'みなさん、こんにちは！' },
  { author: { userId: 'rurun' }, content: 'チャットを開けました。' },
  {
    author: { userId: 'kurao' },
    content:
      'スマホから参加しています。長めのメッセージも、チャット欄の幅に合わせて折り返して読めるか確認中です。',
  },
  { author: { userId: 'rurun' }, content: 'こちらからも見えています 🙌' },
  { author: { userId: 'mumumu' }, content: '先に送られたメッセージも読めました。' },
  { author: { userId: 'kurao' }, content: '準備できています！' },
] satisfies Pick<Message, 'author' | 'content'>[]

export const chatScrollHistory: Message[] = Array.from({ length: 10 }, (_, batch) =>
  chatScrollSamples.map((sample, index) => {
    const number = batch * chatScrollSamples.length + index + 1
    return {
      ...sample,
      messageId: `00000000-0000-4000-8000-${String(number + 1000).padStart(12, '0')}`,
      content: `履歴 ${number} · ${sample.content}`,
      createdAt: new Date(Date.UTC(2026, 5, 28, 11, 3, number)).toISOString(),
    }
  }),
).flat()

export const chatScrollLiveMessages = Array.from({ length: 30 }, (_, index) => ({
  author: { userId: 'rurun' },
  content: `新着 ${index + 1} · スクロール位置を確認するためのメッセージです。`,
}))
