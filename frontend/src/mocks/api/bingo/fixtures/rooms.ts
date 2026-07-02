import type { Room } from '@/api/schema'

export const mockRooms: Room[] = [
  {
    roomId: '00000000-0000-4000-8000-000000000011',
    roomCode: '111111',
    state: 'waiting',
    pickState: 'idle',
    qrCodeVisible: true,
    participants: [
      {
        user: { userId: 'mumumu' },
        joinedAt: '2026-06-28T10:01:00.000Z',
      },
      {
        user: { userId: 'rurun' },
        joinedAt: '2026-06-28T10:02:00.000Z',
      },
      {
        user: { userId: 'kurao' },
        joinedAt: '2026-06-28T10:03:00.000Z',
      },
    ],
    bingoSummaries: [],
    reachSummaries: [],
    settings: {
      name: '未開始表示確認ビンゴ',
      description: 'display 画面でゲーム未開始状態を確認するためのモックルームです。',
      admins: [{ userId: 'mumumu' }],
    },
    createdAt: '2026-06-28T10:00:00.000Z',
    updatedAt: '2026-06-28T10:03:00.000Z',
  },
  {
    roomId: '00000000-0000-4000-8000-000000000001',
    roomCode: '123456',
    state: 'waiting',
    pickState: 'idle',
    qrCodeVisible: true,
    participants: [
      {
        user: { userId: 'kurosaki' },
        joinedAt: '2026-06-27T10:01:00.000Z',
      },
      {
        user: { userId: 'rurun' },
        joinedAt: '2026-06-27T10:02:00.000Z',
      },
      {
        user: { userId: 'howard127' },
        joinedAt: '2026-06-27T10:03:00.000Z',
      },
      {
        user: { userId: 'kurao' },
        joinedAt: '2026-06-27T10:04:00.000Z',
      },
      {
        user: { userId: 'minami' },
        joinedAt: '2026-06-27T10:05:00.000Z',
      },
      {
        user: { userId: 'yamada' },
        joinedAt: '2026-06-27T10:06:00.000Z',
      },
    ],
    bingoSummaries: [],
    reachSummaries: [],
    settings: {
      name: 'デモビンゴ',
      description: 'モック API で動かす待機中のビンゴルームです。',
      admins: [{ userId: 'mumumu' }],
    },
    createdAt: '2026-06-27T10:00:00.000Z',
    updatedAt: '2026-06-27T10:00:00.000Z',
  },
  {
    roomId: '00000000-0000-4000-8000-000000000002',
    roomCode: '234567',
    state: 'playing',
    pickState: 'idle',
    qrCodeVisible: false,
    participants: [
      {
        user: { userId: 'mumumu' },
        joinedAt: '2026-06-27T11:00:00.000Z',
      },
      {
        user: { userId: 'rurun' },
        joinedAt: '2026-06-27T11:01:00.000Z',
      },
      {
        user: { userId: 'howard127' },
        joinedAt: '2026-06-27T11:02:00.000Z',
      },
    ],
    bingoSummaries: [],
    reachSummaries: [
      {
        user: { userId: 'howard127' },
        createdAt: '2026-06-27T11:06:00.000Z',
      },
    ],
    settings: {
      name: '進行中ビンゴ',
      description: 'playing 状態の表示確認用モックルームです。',
      admins: [{ userId: 'mumumu' }],
    },
    createdAt: '2026-06-27T11:00:00.000Z',
    updatedAt: '2026-06-27T11:10:00.000Z',
  },
  {
    roomId: '00000000-0000-4000-8000-000000000003',
    roomCode: '345678',
    state: 'finished',
    pickState: 'idle',
    qrCodeVisible: false,
    participants: [
      {
        user: { userId: 'mumumu' },
        joinedAt: '2026-06-27T12:00:00.000Z',
      },
      {
        user: { userId: 'kurao' },
        joinedAt: '2026-06-27T12:01:00.000Z',
      },
    ],
    bingoSummaries: [
      {
        user: { userId: 'kurao' },
        bingoOrders: [1, 2],
        createdAt: '2026-06-27T12:20:00.000Z',
      },
    ],
    reachSummaries: [],
    settings: {
      name: '終了済みビンゴ',
      description: 'finished 状態の表示確認用モックルームです。',
      admins: [{ userId: 'rurun' }],
    },
    createdAt: '2026-06-27T12:00:00.000Z',
    updatedAt: '2026-06-27T12:30:00.000Z',
  },
  {
    roomId: '00000000-0000-4000-8000-000000000004',
    roomCode: '456789',
    state: 'waiting',
    pickState: 'idle',
    qrCodeVisible: true,
    participants: [
      {
        user: { userId: 'rurun' },
        joinedAt: '2026-06-27T13:01:00.000Z',
      },
      {
        user: { userId: 'kurao' },
        joinedAt: '2026-06-27T13:03:00.000Z',
      },
    ],
    bingoSummaries: [],
    reachSummaries: [],
    settings: {
      name: '少人数ビンゴ',
      description: '参加者はいるが管理者は未参加の待機中ルームです。',
      admins: [{ userId: 'kurosaki' }],
    },
    createdAt: '2026-06-27T13:00:00.000Z',
    updatedAt: '2026-06-27T13:03:00.000Z',
  },
  {
    roomId: '00000000-0000-4000-8000-000000000005',
    roomCode: '567890',
    state: 'playing',
    pickState: 'picking',
    qrCodeVisible: true,
    participants: [
      {
        user: { userId: 'kurosaki' },
        joinedAt: '2026-06-27T14:00:00.000Z',
      },
      {
        user: { userId: 'mumumu' },
        joinedAt: '2026-06-27T14:01:00.000Z',
      },
      {
        user: { userId: 'kurao' },
        joinedAt: '2026-06-27T14:02:00.000Z',
      },
    ],
    bingoSummaries: [
      {
        user: { userId: 'kurosaki' },
        bingoOrders: [],
        createdAt: '2026-06-27T14:07:00.000Z',
      },
      {
        user: { userId: 'mumumu' },
        bingoOrders: [1],
        createdAt: '2026-06-27T14:10:00.000Z',
      },
      {
        user: { userId: 'kurao' },
        bingoOrders: [],
        createdAt: '2026-06-27T14:09:00.000Z',
      },
    ],
    reachSummaries: [
      {
        user: { userId: 'kurosaki' },
        createdAt: '2026-06-27T14:07:00.000Z',
      },
      {
        user: { userId: 'kurao' },
        createdAt: '2026-06-27T14:09:00.000Z',
      },
    ],
    settings: {
      name: '抽選中ビンゴ',
      description: 'pickState が picking の進行中ルームです。',
      admins: [{ userId: 'kurosaki' }, { userId: 'howard127' }],
    },
    createdAt: '2026-06-27T14:00:00.000Z',
    updatedAt: '2026-06-27T14:15:00.000Z',
  },
  {
    roomId: '00000000-0000-4000-8000-000000000006',
    roomCode: '678901',
    state: 'finished',
    pickState: 'exhausted',
    qrCodeVisible: false,
    participants: [
      {
        user: { userId: 'kurosaki' },
        joinedAt: '2026-06-27T15:00:00.000Z',
      },
      {
        user: { userId: 'rurun' },
        joinedAt: '2026-06-27T15:02:00.000Z',
      },
      {
        user: { userId: 'howard127' },
        joinedAt: '2026-06-27T15:04:00.000Z',
      },
      {
        user: { userId: 'kurao' },
        joinedAt: '2026-06-27T15:05:00.000Z',
      },
    ],
    bingoSummaries: [
      {
        user: { userId: 'kurosaki' },
        bingoOrders: [3],
        createdAt: '2026-06-27T15:35:00.000Z',
      },
      {
        user: { userId: 'rurun' },
        bingoOrders: [],
        createdAt: '2026-06-27T15:20:00.000Z',
      },
      {
        user: { userId: 'howard127' },
        bingoOrders: [1, 2],
        createdAt: '2026-06-27T15:25:00.000Z',
      },
      {
        user: { userId: 'kurao' },
        bingoOrders: [],
        createdAt: '2026-06-27T15:23:00.000Z',
      },
    ],
    reachSummaries: [
      {
        user: { userId: 'rurun' },
        createdAt: '2026-06-27T15:20:00.000Z',
      },
      {
        user: { userId: 'kurao' },
        createdAt: '2026-06-27T15:23:00.000Z',
      },
    ],
    settings: {
      name: '全番号終了ビンゴ',
      description: 'pickState が exhausted の終了済みルームです。',
      admins: [{ userId: 'howard127' }],
    },
    createdAt: '2026-06-27T15:00:00.000Z',
    updatedAt: '2026-06-27T15:45:00.000Z',
  },
  {
    roomId: '00000000-0000-4000-8000-000000000007',
    roomCode: '789012',
    state: 'waiting',
    pickState: 'idle',
    qrCodeVisible: false,
    participants: [],
    bingoSummaries: [],
    reachSummaries: [],
    settings: {
      name: '空の待機ルーム',
      description: '参加者がまだいない状態の確認用ルームです。',
      admins: [{ userId: 'kurao' }, { userId: 'mumumu' }],
    },
    createdAt: '2026-06-27T16:00:00.000Z',
    updatedAt: '2026-06-27T16:00:00.000Z',
  },
  {
    roomId: '00000000-0000-4000-8000-000000000008',
    roomCode: '890123',
    state: 'playing',
    pickState: 'picking',
    qrCodeVisible: false,
    participants: [
      {
        user: { userId: 'howard127' },
        joinedAt: '2026-06-27T17:00:00.000Z',
      },
      {
        user: { userId: 'mumumu' },
        joinedAt: '2026-06-27T17:02:00.000Z',
      },
    ],
    bingoSummaries: [
      {
        user: { userId: 'howard127' },
        bingoOrders: [],
        createdAt: '2026-06-27T17:12:00.000Z',
      },
      {
        user: { userId: 'mumumu' },
        bingoOrders: [1],
        createdAt: '2026-06-27T17:16:00.000Z',
      },
    ],
    reachSummaries: [
      {
        user: { userId: 'howard127' },
        createdAt: '2026-06-27T17:12:00.000Z',
      },
    ],
    settings: {
      name: '二人進行ビンゴ',
      description: '管理者が参加者に含まれていない進行中ルームです。',
      admins: [{ userId: 'rurun' }],
    },
    createdAt: '2026-06-27T17:00:00.000Z',
    updatedAt: '2026-06-27T17:20:00.000Z',
  },
  {
    roomId: '00000000-0000-4000-8000-000000000009',
    roomCode: '901234',
    state: 'finished',
    pickState: 'idle',
    qrCodeVisible: true,
    participants: [
      {
        user: { userId: 'kurosaki' },
        joinedAt: '2026-06-27T18:00:00.000Z',
      },
      {
        user: { userId: 'mumumu' },
        joinedAt: '2026-06-27T18:01:00.000Z',
      },
      {
        user: { userId: 'rurun' },
        joinedAt: '2026-06-27T18:02:00.000Z',
      },
      {
        user: { userId: 'howard127' },
        joinedAt: '2026-06-27T18:03:00.000Z',
      },
      {
        user: { userId: 'kurao' },
        joinedAt: '2026-06-27T18:04:00.000Z',
      },
    ],
    bingoSummaries: [
      {
        user: { userId: 'kurosaki' },
        bingoOrders: [4],
        createdAt: '2026-06-27T18:35:00.000Z',
      },
      {
        user: { userId: 'mumumu' },
        bingoOrders: [],
        createdAt: '2026-06-27T18:18:00.000Z',
      },
      {
        user: { userId: 'rurun' },
        bingoOrders: [1, 3],
        createdAt: '2026-06-27T18:20:00.000Z',
      },
      {
        user: { userId: 'howard127' },
        bingoOrders: [],
        createdAt: '2026-06-27T18:22:00.000Z',
      },
      {
        user: { userId: 'kurao' },
        bingoOrders: [2],
        createdAt: '2026-06-27T18:27:00.000Z',
      },
    ],
    reachSummaries: [
      {
        user: { userId: 'mumumu' },
        createdAt: '2026-06-27T18:18:00.000Z',
      },
      {
        user: { userId: 'howard127' },
        createdAt: '2026-06-27T18:22:00.000Z',
      },
    ],
    settings: {
      name: '全員参加ビンゴ',
      description: '5人全員が参加している終了済みルームです。',
      admins: [{ userId: 'kurosaki' }, { userId: 'rurun' }],
    },
    createdAt: '2026-06-27T18:00:00.000Z',
    updatedAt: '2026-06-27T18:40:00.000Z',
  },
  {
    roomId: '00000000-0000-4000-8000-000000000010',
    roomCode: '012345',
    state: 'waiting',
    pickState: 'idle',
    qrCodeVisible: true,
    participants: [
      {
        user: { userId: 'howard127' },
        joinedAt: '2026-06-27T19:01:00.000Z',
      },
    ],
    bingoSummaries: [],
    reachSummaries: [],
    settings: {
      name: '招待中ビンゴ',
      description: 'QR 表示中で参加者が少ない待機中ルームです。',
      admins: [{ userId: 'howard127' }],
    },
    createdAt: '2026-06-27T19:00:00.000Z',
    updatedAt: '2026-06-27T19:01:00.000Z',
  },
]
export const fallbackRoom = mockRooms[0] as Room
