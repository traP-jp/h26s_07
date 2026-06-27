import type {
    AllPickedBody,
    BingoSummary,
    BingoUpdate,
    Card,
    CardChanges,
    CardCell,
    CardCellState,
    CreateMessageRequest,
    CreateRoomRequest,
    DisplayGameFinishedBody,
    DisplayGameSettingsUpdatedBody,
    DisplayGameStartedBody,
    DisplayInitializedBody,
    DisplayPickFinishedBody,
    GameSettings,
    GameSettingsInput,
    HideQrCodeBody,
    Line,
    Message,
    MessageCreatedBody,
    PickedBall,
    PickCanceledBody,
    PickStartedBody,
    ParticipantGameFinishedBody,
    ParticipantGameSettingsUpdatedBody,
    ParticipantGameStartedBody,
    ParticipantInitializedBody,
    ParticipantPickFinishedBody,
    ReachUpdate,
    Room,
    RoomId,
    ShowQrCodeBody,
    UpdateGameSettingsRequest,
    User,
    UserId,
    WebSocketEventType,
    WebSocketMode,
} from '@/api/schema'
import { ws } from 'msw'

import { http } from '../http'

type MockRoom = {
    room: Room
    messages: Message[]
    pickedBalls: PickedBall[]
    cards: Map<UserId, Card>
    reachedUserIds: Set<UserId>
    bingoRecords: Array<{
        userId: UserId
        lineKey: string
        order: number
    }>
}

type ApiError = {
    code: string
    message: string
    description: string
}

type MockUser = User & {
    name: string
}

type MockSocketConnection = {
    roomId: RoomId
    mode: WebSocketMode
    userId: UserId
    send(data: string): void
    close(code?: number, reason?: string): void
}

type MockWebSocketEvent<TBody> = {
    type: WebSocketEventType
    body: TBody
}

type CardSnapshot = {
    openCellIndices: Set<number>
    reachLineKeys: Set<string>
    bingoLineKeys: Set<string>
}

type FinishPickResult = {
    pickedBall: PickedBall
    cardChangesByUserId: Map<UserId, CardChanges>
    newBingos: BingoUpdate[]
    newReaches: ReachUpdate[]
    allPicked: boolean
}

type JsonRequest = {
    json(): Promise<unknown>
}

const LINES: Line[] = [
    [0, 1, 2, 3, 4],
    [5, 6, 7, 8, 9],
    [10, 11, 12, 13, 14],
    [15, 16, 17, 18, 19],
    [20, 21, 22, 23, 24],
    [0, 5, 10, 15, 20],
    [1, 6, 11, 16, 21],
    [2, 7, 12, 17, 22],
    [3, 8, 13, 18, 23],
    [4, 9, 14, 19, 24],
    [0, 6, 12, 18, 24],
    [4, 8, 12, 16, 20],
]

const COLUMN_RANGES: Array<[number, number]> = [
    [1, 15],
    [16, 30],
    [31, 45],
    [46, 60],
    [61, 75],
]

const state = {
    rooms: new Map<RoomId, MockRoom>(),
}

const roomSocket = ws.link('*/api/rooms/:roomId/ws')
const socketConnections = new Set<MockSocketConnection>()

let idCounter = 1

function createId(): string {
    const suffix = String(idCounter).padStart(12, '0')
    idCounter += 1
    return `00000000-0000-4000-8000-${suffix}`
}

function now(): string {
    return new Date().toISOString()
}

function createUser(userId: UserId): MockUser {
    return {
        userId,
        name: userId,
    }
}

function currentUser(request: Request): MockUser {
    const forwardedUser = request.headers.get('X-Forwarded-User')?.trim()
    return createUser(forwardedUser || 'mumumu')
}

async function readJson<T>(request: JsonRequest): Promise<T | undefined> {
    try {
        return (await request.json()) as T
    } catch {
        return undefined
    }
}

function apiError(code: string, message: string, description: string): ApiError {
    return { code, message, description }
}

function badRequest(description: string): ApiError {
    return apiError('BadRequest', 'リクエスト body が不正です。', description)
}

function notFound(): ApiError {
    return apiError(
        'RoomNotFound',
        '指定したルームが存在しません。',
        'roomId に対応するルームが見つかりません。',
    )
}

function forbidden(description: string): ApiError {
    return apiError('Forbidden', '権限がありません。', description)
}

function conflict(description: string): ApiError {
    return apiError('Conflict', '現在の状態では操作できません。', description)
}

function nextRoomCode(): string {
    return String(100000 + ((idCounter * 137) % 900000)).padStart(6, '0')
}

function uniqueUserIds(userIds: UserId[]): UserId[] {
    return [...new Set(userIds.map((userId) => userId.trim()).filter(Boolean))]
}

function settingsFromInput(
    input: GameSettingsInput,
    fallbackAdminIds: UserId[],
): GameSettings | undefined {
    if (input.name.trim() === '') {
        return undefined
    }

    const adminIds =
        input.adminUserIds === undefined
            ? uniqueUserIds(fallbackAdminIds)
            : uniqueUserIds(input.adminUserIds)

    if (adminIds.length === 0) {
        return undefined
    }

    return {
        name: input.name,
        description: input.description,
        admins: adminIds.map(createUser),
    }
}

function isAdmin(room: Room, userId: UserId): boolean {
    return room.settings.admins.some((admin) => admin.userId === userId)
}

function isParticipant(room: Room, userId: UserId): boolean {
    return room.participants.some((participant) => participant.user.userId === userId)
}

function addParticipant(roomState: MockRoom, user: User): void {
    if (isParticipant(roomState.room, user.userId)) {
        return
    }

    roomState.room.participants.push({
        user,
        joinedAt: now(),
    })
    touch(roomState.room)
}

function touch(room: Room): void {
    room.updatedAt = now()
}

function getRoom(roomId: RoomId): MockRoom | undefined {
    return state.rooms.get(roomId)
}

function pathParam(value: unknown): string | undefined {
    if (Array.isArray(value)) {
        return typeof value[0] === 'string' ? value[0] : undefined
    }

    return typeof value === 'string' ? value : undefined
}

function sendEvent<TBody>(
    connection: MockSocketConnection,
    type: WebSocketEventType,
    body: TBody,
): void {
    const event = {
        type,
        body,
    } satisfies MockWebSocketEvent<TBody>

    connection.send(JSON.stringify(event))
}

function roomConnectionList(roomId: RoomId): MockSocketConnection[] {
    return [...socketConnections].filter((connection) => connection.roomId === roomId)
}

function pickedNumberSet(roomState: MockRoom): Set<number> {
    return new Set(roomState.pickedBalls)
}

function openCellIndices(card: Card, pickedNumbers: Set<number>): Set<number> {
    return new Set(
        card.cells
            .filter((cell) => cell.number === null || pickedNumbers.has(cell.number))
            .map((cell) => cell.index),
    )
}

function snapshotCard(card: Card, pickedNumbers: Set<number>): CardSnapshot {
    return {
        openCellIndices: openCellIndices(card, pickedNumbers),
        reachLineKeys: new Set(card.reachLines.map(lineKey)),
        bingoLineKeys: new Set(card.bingoLines.map(lineKey)),
    }
}

function emptyCardChanges(): CardChanges {
    return {
        openedCellIndices: [],
        newReachLines: [],
        newBingoLines: [],
    }
}

function hashUserId(userId: UserId): number {
    return [...userId].reduce((sum, char) => sum + char.charCodeAt(0), 0)
}

function createCard(ownerUserId: UserId): Card {
    const offset = hashUserId(ownerUserId) % 15
    const cells: CardCell[] = []

    for (let index = 0; index < 25; index += 1) {
        const row = Math.floor(index / 5)
        const column = index % 5

        if (index === 12) {
            cells.push({
                index,
                number: null,
                displayText: 'FREE',
                cellState: 'open',
            })
            continue
        }

        const [start, end] = COLUMN_RANGES[column] ?? [1, 15]
        const rangeSize = end - start + 1
        const number = start + ((offset + column * 5 + row * 3) % rangeSize)

        cells.push({
            index,
            number,
            displayText: String(number),
            cellState: 'closed',
        })
    }

    return {
        cardId: createId(),
        ownerUserId,
        cells,
        bingoLines: [],
        reachLines: [],
    }
}

function lineCells(card: Card, line: Line): CardCell[] {
    const cellsByIndex = new Map(card.cells.map((cell) => [cell.index, cell]))
    return line
        .map((index) => cellsByIndex.get(index))
        .filter((cell): cell is CardCell => cell !== undefined)
}

function lineKey(line: Line): string {
    return line.join(',')
}

function updateBingoSummaries(roomState: MockRoom): void {
    const summariesByUserId = new Map<UserId, BingoSummary>()

    for (const record of roomState.bingoRecords) {
        const participant = roomState.room.participants.find(
            (entry) => entry.user.userId === record.userId,
        )
        if (participant === undefined) {
            continue
        }

        const summary =
            summariesByUserId.get(record.userId) ??
            ({
                user: participant.user,
                bingoOrders: [],
            } satisfies BingoSummary)

        summary.bingoOrders.push(record.order)
        summariesByUserId.set(record.userId, summary)
    }

    roomState.room.bingoSummaries = [...summariesByUserId.values()]
}

function participantInitializedBody(
    roomState: MockRoom,
    userId: UserId,
): ParticipantInitializedBody {
    const card = roomState.cards.get(userId)
    const body: ParticipantInitializedBody = {
        state: roomState.room.state,
        settings: roomState.room.settings,
        pickState: roomState.room.pickState,
        pickedBalls: roomState.pickedBalls,
        bingoSummaries: roomState.room.bingoSummaries,
    }

    if (card !== undefined && roomState.room.state !== 'waiting') {
        body.card = card
    }

    return body
}

function displayInitializedBody(roomState: MockRoom): DisplayInitializedBody {
    return {
        state: roomState.room.state,
        settings: roomState.room.settings,
        pickState: roomState.room.pickState,
        participantCount: roomState.room.participants.length,
        pickedBalls: roomState.pickedBalls,
        qrCodeVisible: roomState.room.qrCodeVisible,
        bingoSummaries: roomState.room.bingoSummaries,
    }
}

function sendInitialized(connection: MockSocketConnection, roomState: MockRoom): void {
    if (connection.mode === 'participant') {
        sendEvent<ParticipantInitializedBody>(
            connection,
            'Initialized',
            participantInitializedBody(roomState, connection.userId),
        )
        return
    }

    sendEvent<DisplayInitializedBody>(connection, 'Initialized', displayInitializedBody(roomState))
}

function broadcastRoom(
    roomState: MockRoom,
    send: (connection: MockSocketConnection) => void,
): void {
    for (const connection of roomConnectionList(roomState.room.roomId)) {
        send(connection)
    }
}

function broadcastGameStarted(roomState: MockRoom): void {
    broadcastRoom(roomState, (connection) => {
        if (connection.mode === 'participant') {
            const card = roomState.cards.get(connection.userId)
            if (card !== undefined) {
                sendEvent<ParticipantGameStartedBody>(connection, 'GameStarted', { card })
            }
            return
        }

        sendEvent<DisplayGameStartedBody>(connection, 'GameStarted', {
            participantCount: roomState.room.participants.length,
        })
    })
}

function broadcastPickStarted(roomState: MockRoom): void {
    broadcastRoom(roomState, (connection) => {
        sendEvent<PickStartedBody>(connection, 'PickStarted', {})
    })
}

function broadcastPickCanceled(roomState: MockRoom): void {
    broadcastRoom(roomState, (connection) => {
        sendEvent<PickCanceledBody>(connection, 'PickCanceled', {})
    })
}

function broadcastPickFinished(roomState: MockRoom, result: FinishPickResult): void {
    broadcastRoom(roomState, (connection) => {
        if (connection.mode === 'participant') {
            const card = roomState.cards.get(connection.userId)
            if (card === undefined) {
                return
            }

            sendEvent<ParticipantPickFinishedBody>(connection, 'PickFinished', {
                pickedBall: result.pickedBall,
                pickState: roomState.room.pickState,
                card,
                cardChanges:
                    result.cardChangesByUserId.get(connection.userId) ?? emptyCardChanges(),
                pickedBalls: roomState.pickedBalls,
                bingoSummaries: roomState.room.bingoSummaries,
                newBingos: result.newBingos,
                newReaches: result.newReaches,
            })
            return
        }

        sendEvent<DisplayPickFinishedBody>(connection, 'PickFinished', {
            pickedBall: result.pickedBall,
            pickState: roomState.room.pickState,
            participantCount: roomState.room.participants.length,
            bingoSummaries: roomState.room.bingoSummaries,
            newBingos: result.newBingos,
            newReaches: result.newReaches,
            pickedBalls: roomState.pickedBalls,
        })
    })
}

function broadcastGameFinished(roomState: MockRoom): void {
    broadcastRoom(roomState, (connection) => {
        if (connection.mode === 'participant') {
            const card = roomState.cards.get(connection.userId)
            if (card !== undefined) {
                sendEvent<ParticipantGameFinishedBody>(connection, 'GameFinished', {
                    state: 'finished',
                    pickState: 'idle',
                    card,
                    bingoSummaries: roomState.room.bingoSummaries,
                })
            }
            return
        }

        sendEvent<DisplayGameFinishedBody>(connection, 'GameFinished', {
            state: 'finished',
            pickState: 'idle',
            participantCount: roomState.room.participants.length,
            bingoSummaries: roomState.room.bingoSummaries,
        })
    })
}

function broadcastMessageCreated(roomState: MockRoom, message: Message): void {
    broadcastRoom(roomState, (connection) => {
        sendEvent<MessageCreatedBody>(connection, 'MessageCreated', { message })
    })
}

function broadcastAllPicked(roomState: MockRoom): void {
    const body: AllPickedBody = { pickedBalls: roomState.pickedBalls }
    broadcastRoom(roomState, (connection) => {
        sendEvent<AllPickedBody>(connection, 'AllPicked', body)
    })
}

function broadcastGameSettingsUpdated(roomState: MockRoom): void {
    broadcastRoom(roomState, (connection) => {
        if (connection.mode === 'participant') {
            sendEvent<ParticipantGameSettingsUpdatedBody>(connection, 'GameSettingsUpdated', {
                settings: roomState.room.settings,
            })
            return
        }

        sendEvent<DisplayGameSettingsUpdatedBody>(connection, 'GameSettingsUpdated', {
            settings: roomState.room.settings,
        })
    })
}

function broadcastShowQRCode(roomState: MockRoom): void {
    broadcastRoom(roomState, (connection) => {
        if (connection.mode === 'display') {
            sendEvent<ShowQrCodeBody>(connection, 'ShowQRCode', {})
        }
    })
}

function broadcastHideQRCode(roomState: MockRoom): void {
    broadcastRoom(roomState, (connection) => {
        if (connection.mode === 'display') {
            sendEvent<HideQrCodeBody>(connection, 'HideQRCode', {})
        }
    })
}

function recomputeCards(roomState: MockRoom): void {
    const pickedNumbers = new Set(roomState.pickedBalls)
    const knownBingoKeys = new Set(
        roomState.bingoRecords.map((record) => `${record.userId}:${record.lineKey}`),
    )

    for (const [userId, card] of roomState.cards) {
        const bingoLines: Line[] = []
        const reachLines: Line[] = []

        for (const line of LINES) {
            const cells = lineCells(card, line)
            if (cells.length !== 5) {
                continue
            }

            const openCount = cells.filter(
                (cell) => cell.number === null || pickedNumbers.has(cell.number),
            ).length
            if (openCount === 5) {
                bingoLines.push([...line])

                const key = `${userId}:${lineKey(line)}`
                if (!knownBingoKeys.has(key)) {
                    knownBingoKeys.add(key)
                    roomState.bingoRecords.push({
                        userId,
                        lineKey: lineKey(line),
                        order: roomState.bingoRecords.length + 1,
                    })
                }
            } else if (openCount === 4) {
                reachLines.push([...line])
            }
        }

        const bingoCellIndices = new Set(bingoLines.flat())
        const reachCellIndices = new Set(reachLines.flat())

        card.cells = card.cells.map((cell) => {
            const opened = cell.number === null || pickedNumbers.has(cell.number)
            let cellState: CardCellState = opened ? 'open' : 'closed'

            if (bingoCellIndices.has(cell.index)) {
                cellState = 'bingo'
            } else if (reachCellIndices.has(cell.index)) {
                cellState = 'reach'
            }

            return {
                ...cell,
                cellState,
            }
        })
        card.bingoLines = bingoLines
        card.reachLines = reachLines
    }

    updateBingoSummaries(roomState)
}

function createMockRoom(settingsInput: GameSettingsInput, creator: User): MockRoom | undefined {
    const settings = settingsFromInput(
        settingsInput,
        uniqueUserIds([creator.userId, ...(settingsInput.adminUserIds ?? [])]),
    )
    if (settings === undefined) {
        return undefined
    }

    const createdAt = now()
    const room: Room = {
        roomId: createId(),
        roomCode: nextRoomCode(),
        state: 'waiting',
        pickState: 'idle',
        qrCodeVisible: true,
        participants: [],
        bingoSummaries: [],
        settings,
        createdAt,
        updatedAt: createdAt,
    }

    return {
        room,
        messages: [],
        pickedBalls: [],
        cards: new Map(),
        reachedUserIds: new Set(),
        bingoRecords: [],
    }
}

function createMessage(author: User, content: string): Message {
    return {
        messageId: createId(),
        content,
        author,
        createdAt: now(),
    }
}

function validateSettingsInput(input: GameSettingsInput | undefined): input is GameSettingsInput {
    if (
        input === undefined ||
        typeof input.name !== 'string' ||
        input.name.trim() === '' ||
        typeof input.description !== 'string' ||
        input.description.trim() === ''
    ) {
        return false
    }

    if (input.adminUserIds === undefined) {
        return true
    }

    return Array.isArray(input.adminUserIds) && uniqueUserIds(input.adminUserIds).length > 0
}

function startRoom(roomState: MockRoom): ApiError | undefined {
    if (roomState.room.state !== 'waiting') {
        return conflict('ルームが waiting ではありません。')
    }
    if (roomState.room.participants.length === 0) {
        return conflict('参加者がいないため開始できません。')
    }

    roomState.cards.clear()
    roomState.bingoRecords = []
    roomState.pickedBalls = []
    roomState.reachedUserIds.clear()

    for (const participant of roomState.room.participants) {
        roomState.cards.set(participant.user.userId, createCard(participant.user.userId))
    }

    roomState.room.state = 'playing'
    roomState.room.pickState = 'idle'
    roomState.room.bingoSummaries = []
    touch(roomState.room)

    return undefined
}

function finishPick(roomState: MockRoom): ApiError | FinishPickResult {
    if (roomState.room.state !== 'playing' || roomState.room.pickState !== 'picking') {
        return conflict('現在抽選中ではない、またはゲーム状態が不正です。')
    }

    const previousPickedNumbers = pickedNumberSet(roomState)
    const beforeSnapshots = new Map<UserId, CardSnapshot>()
    for (const [userId, card] of roomState.cards) {
        beforeSnapshots.set(userId, snapshotCard(card, previousPickedNumbers))
    }

    const previousBingoRecordCount = roomState.bingoRecords.length
    const nextNumber = Array.from({ length: 75 }, (_, index) => index + 1).find(
        (number) => !previousPickedNumbers.has(number),
    )

    if (nextNumber === undefined) {
        roomState.room.pickState = 'exhausted'
        touch(roomState.room)
        return conflict('抽選可能な球がありません。')
    }

    const pickedBall: PickedBall = nextNumber

    roomState.pickedBalls.push(pickedBall)
    recomputeCards(roomState)

    roomState.room.pickState = roomState.pickedBalls.length >= 75 ? 'exhausted' : 'idle'
    touch(roomState.room)

    const nextPickedNumbers = pickedNumberSet(roomState)
    const cardChangesByUserId = new Map<UserId, CardChanges>()
    const newReaches: ReachUpdate[] = []

    for (const [userId, card] of roomState.cards) {
        const before = beforeSnapshots.get(userId)
        const afterOpenCellIndices = openCellIndices(card, nextPickedNumbers)
        const openedCellIndices = [...afterOpenCellIndices].filter(
            (index) => !before?.openCellIndices.has(index),
        )
        const newReachLines = card.reachLines.filter(
            (line) => !before?.reachLineKeys.has(lineKey(line)),
        )
        const newBingoLines = card.bingoLines.filter(
            (line) => !before?.bingoLineKeys.has(lineKey(line)),
        )

        cardChangesByUserId.set(userId, {
            openedCellIndices,
            newReachLines,
            newBingoLines,
        })

        if (card.reachLines.length > 0 && !roomState.reachedUserIds.has(userId)) {
            roomState.reachedUserIds.add(userId)
            const participant = roomState.room.participants.find(
                (entry) => entry.user.userId === userId,
            )
            if (participant !== undefined) {
                newReaches.push({ user: participant.user })
            }
        }
    }

    const newRecords = roomState.bingoRecords.slice(previousBingoRecordCount)
    const newBingosByUserId = new Map<UserId, number[]>()
    for (const record of newRecords) {
        const orders = newBingosByUserId.get(record.userId) ?? []
        orders.push(record.order)
        newBingosByUserId.set(record.userId, orders)
    }

    const newBingos: BingoUpdate[] = [...newBingosByUserId.entries()].flatMap(
        ([userId, newBingoOrders]) => {
            const summary = roomState.room.bingoSummaries.find(
                (entry) => entry.user.userId === userId,
            )
            if (summary === undefined) {
                return []
            }

            return [
                {
                    user: summary.user,
                    newBingoOrders,
                    bingoOrders: summary.bingoOrders,
                },
            ]
        },
    )

    return {
        pickedBall,
        cardChangesByUserId,
        newBingos,
        newReaches,
        allPicked: roomState.room.pickState === 'exhausted',
    }
}

function assertVisible(roomState: MockRoom, user: User): ApiError | undefined {
    if (isParticipant(roomState.room, user.userId) || isAdmin(roomState.room, user.userId)) {
        return undefined
    }

    return forbidden('このルームを閲覧する権限がありません。')
}

function assertAdmin(roomState: MockRoom, user: User): ApiError | undefined {
    if (isAdmin(roomState.room, user.userId)) {
        return undefined
    }

    return forbidden('admin ではありません。')
}

function isApiError(value: ApiError | FinishPickResult): value is ApiError {
    return 'code' in value
}

function sleep(ms: number): Promise<void> {
    return new Promise((resolve) => {
        window.setTimeout(resolve, ms)
    })
}

async function runMockPickCycle(roomState: MockRoom): Promise<void> {
    if (roomState.room.state === 'waiting') {
        const startError = startRoom(roomState)
        if (startError !== undefined) {
            return
        }
        broadcastGameStarted(roomState)
    }

    if (roomState.room.state !== 'playing' || roomState.room.pickState !== 'idle') {
        return
    }

    roomState.room.pickState = 'picking'
    touch(roomState.room)
    broadcastPickStarted(roomState)

    await sleep(800)

    const result = finishPick(roomState)
    if (isApiError(result)) {
        return
    }

    broadcastPickFinished(roomState, result)
    if (result.allPicked) {
        broadcastAllPicked(roomState)
    }
}

function handleSocketMessage(connection: MockSocketConnection, data: unknown): void {
    if (typeof data !== 'string') {
        return
    }

    let payload: { type?: string } | undefined
    try {
        payload = JSON.parse(data) as { type?: string }
    } catch {
        return
    }

    const roomState = getRoom(connection.roomId)
    if (roomState === undefined) {
        return
    }

    if (payload.type === 'mock:start') {
        const startError = startRoom(roomState)
        if (startError === undefined) {
            broadcastGameStarted(roomState)
        }
        return
    }

    if (payload.type === 'mock:pick') {
        void runMockPickCycle(roomState)
        return
    }

    if (payload.type === 'mock:finish' && roomState.room.state === 'playing') {
        const wasPicking = roomState.room.pickState === 'picking'
        roomState.room.state = 'finished'
        roomState.room.pickState = 'idle'
        touch(roomState.room)

        if (wasPicking) {
            broadcastPickCanceled(roomState)
        }
        broadcastGameFinished(roomState)
    }
}

const roomWebSocketHandler = roomSocket.addEventListener('connection', ({ client, params }) => {
    const roomId = pathParam(params.roomId)
    const roomState = roomId === undefined ? undefined : getRoom(roomId)
    const mode = client.url.searchParams.get('mode')
    const userId = client.url.searchParams.get('userId')?.trim() || 'mumumu'

    if (roomId === undefined || roomState === undefined) {
        client.close(1008, 'Room not found')
        return
    }

    if (mode !== 'participant' && mode !== 'display') {
        client.close(1008, 'Invalid mode')
        return
    }

    if (mode === 'participant' && !isParticipant(roomState.room, userId)) {
        client.close(1008, 'Participant required')
        return
    }

    const connection: MockSocketConnection = {
        roomId,
        mode,
        userId,
        send(data) {
            client.send(data)
        },
        close(code, reason) {
            client.close(code, reason)
        },
    }

    socketConnections.add(connection)
    client.addEventListener(
        'close',
        () => {
            socketConnections.delete(connection)
        },
        { once: true },
    )
    client.addEventListener('message', (event) => {
        handleSocketMessage(connection, event.data)
    })

    sendInitialized(connection, roomState)
})

function seedRooms(): void {
    const admin = createUser('mumumu')
    const roomState = createMockRoom(
        {
            name: 'デモビンゴ',
            description: 'モック API で動かす待機中のビンゴルームです。',
            adminUserIds: ['mumumu'],
        },
        admin,
    )

    if (roomState === undefined) {
        return
    }

    addParticipant(roomState, admin)
    addParticipant(roomState, createUser('saba'))
    roomState.messages.push(createMessage(admin, 'モックルームへようこそ'))
    state.rooms.set(roomState.room.roomId, roomState)
}

seedRooms()

export const bingoHandlers = [
    roomWebSocketHandler,

    http.get('/api/me', ({ request, response }) => {
        return response(200).json(currentUser(request))
    }),

    http.get('/api/rooms', ({ response }) => {
        return response(200).json([...state.rooms.values()].map((roomState) => roomState.room))
    }),

    http.post('/api/rooms', async ({ request, response }) => {
        const body = await readJson<CreateRoomRequest>(request)
        const user = currentUser(request)

        if (!validateSettingsInput(body?.settings)) {
            return response(400).json(
                badRequest('settings.name と settings.description が必要です。'),
            )
        }

        const roomState = createMockRoom(body.settings, user)
        if (roomState === undefined) {
            return response(400).json(badRequest('adminUserIds を指定する場合は空にできません。'))
        }

        state.rooms.set(roomState.room.roomId, roomState)
        return response(200).json(roomState.room)
    }),

    http.get('/api/rooms/{roomId}', ({ params, request, response }) => {
        const roomState = getRoom(params.roomId)
        if (roomState === undefined) {
            return response(404).json(notFound())
        }

        const visibilityError = assertVisible(roomState, currentUser(request))
        if (visibilityError !== undefined) {
            return response(403).json(visibilityError)
        }

        return response(200).json(roomState.room)
    }),

    http.post('/api/rooms/{roomId}/participants', ({ params, request, response }) => {
        const roomState = getRoom(params.roomId)
        if (roomState === undefined) {
            return response(404).json(notFound())
        }
        if (roomState.room.state !== 'waiting') {
            return response(409).json(conflict('ルームが waiting ではないため参加できません。'))
        }

        addParticipant(roomState, currentUser(request))
        return response(204).empty()
    }),

    http.get('/api/rooms/{roomId}/chats', ({ params, request, response }) => {
        const roomState = getRoom(params.roomId)
        if (roomState === undefined) {
            return response(404).json(notFound())
        }

        const visibilityError = assertVisible(roomState, currentUser(request))
        if (visibilityError !== undefined) {
            return response(403).json(visibilityError)
        }

        return response(200).json(roomState.messages)
    }),

    http.post('/api/rooms/{roomId}/chats', async ({ params, request, response }) => {
        const roomState = getRoom(params.roomId)
        if (roomState === undefined) {
            return response(404).json(notFound())
        }

        const user = currentUser(request)
        if (!isParticipant(roomState.room, user.userId) && !isAdmin(roomState.room, user.userId)) {
            return response(403).json(forbidden('このルームにチャット投稿する権限がありません。'))
        }
        if (roomState.room.state === 'finished') {
            return response(409).json(conflict('finished のルームには投稿できません。'))
        }

        const body = await readJson<CreateMessageRequest>(request)
        const content = body?.content.trim()
        if (content === undefined || content.length === 0 || content.length > 500) {
            return response(400).json(
                badRequest('content は 1 文字以上 500 文字以下で指定してください。'),
            )
        }

        const message = createMessage(user, content)
        roomState.messages.push(message)
        touch(roomState.room)
        broadcastMessageCreated(roomState, message)

        return response(200).json(message)
    }),

    http.post('/api/rooms/{roomId}/control/start', ({ params, request, response }) => {
        const roomState = getRoom(params.roomId)
        if (roomState === undefined) {
            return response(404).json(notFound())
        }

        const adminError = assertAdmin(roomState, currentUser(request))
        if (adminError !== undefined) {
            return response(403).json(adminError)
        }

        const startError = startRoom(roomState)
        if (startError !== undefined) {
            return response(409).json(startError)
        }

        broadcastGameStarted(roomState)

        return response(204).empty()
    }),

    http.post('/api/rooms/{roomId}/control/finish', ({ params, request, response }) => {
        const roomState = getRoom(params.roomId)
        if (roomState === undefined) {
            return response(404).json(notFound())
        }

        const adminError = assertAdmin(roomState, currentUser(request))
        if (adminError !== undefined) {
            return response(403).json(adminError)
        }
        if (roomState.room.state !== 'playing') {
            return response(409).json(
                conflict('ルームが playing ではない、または既に finished です。'),
            )
        }

        const wasPicking = roomState.room.pickState === 'picking'
        roomState.room.state = 'finished'
        roomState.room.pickState = 'idle'
        touch(roomState.room)

        if (wasPicking) {
            broadcastPickCanceled(roomState)
        }
        broadcastGameFinished(roomState)

        return response(204).empty()
    }),

    http.post('/api/rooms/{roomId}/control/pick/start', ({ params, request, response }) => {
        const roomState = getRoom(params.roomId)
        if (roomState === undefined) {
            return response(404).json(notFound())
        }

        const adminError = assertAdmin(roomState, currentUser(request))
        if (adminError !== undefined) {
            return response(403).json(adminError)
        }
        if (roomState.room.state !== 'playing' || roomState.room.pickState !== 'idle') {
            return response(409).json(conflict('playing + idle ではありません。'))
        }
        if (roomState.pickedBalls.length >= 75) {
            roomState.room.pickState = 'exhausted'
            touch(roomState.room)
            return response(409).json(conflict('抽選可能な球がありません。'))
        }

        roomState.room.pickState = 'picking'
        touch(roomState.room)
        broadcastPickStarted(roomState)

        return response(204).empty()
    }),

    http.post('/api/rooms/{roomId}/control/pick/cancel', ({ params, request, response }) => {
        const roomState = getRoom(params.roomId)
        if (roomState === undefined) {
            return response(404).json(notFound())
        }

        const adminError = assertAdmin(roomState, currentUser(request))
        if (adminError !== undefined) {
            return response(403).json(adminError)
        }
        if (roomState.room.state !== 'playing' || roomState.room.pickState !== 'picking') {
            return response(409).json(conflict('現在抽選中ではない、またはゲーム状態が不正です。'))
        }

        roomState.room.pickState = 'idle'
        touch(roomState.room)
        broadcastPickCanceled(roomState)

        return response(204).empty()
    }),

    http.post('/api/rooms/{roomId}/control/pick/finish', ({ params, request, response }) => {
        const roomState = getRoom(params.roomId)
        if (roomState === undefined) {
            return response(404).json(notFound())
        }

        const adminError = assertAdmin(roomState, currentUser(request))
        if (adminError !== undefined) {
            return response(403).json(adminError)
        }

        const result = finishPick(roomState)
        if (isApiError(result)) {
            return response(409).json(result)
        }

        broadcastPickFinished(roomState, result)
        if (result.allPicked) {
            broadcastAllPicked(roomState)
        }

        return response(204).empty()
    }),

    http.post('/api/rooms/{roomId}/control/qrcode/show', ({ params, request, response }) => {
        const roomState = getRoom(params.roomId)
        if (roomState === undefined) {
            return response(404).json(notFound())
        }

        const adminError = assertAdmin(roomState, currentUser(request))
        if (adminError !== undefined) {
            return response(403).json(adminError)
        }

        roomState.room.qrCodeVisible = true
        touch(roomState.room)
        broadcastShowQRCode(roomState)

        return response(204).empty()
    }),

    http.post('/api/rooms/{roomId}/control/qrcode/hide', ({ params, request, response }) => {
        const roomState = getRoom(params.roomId)
        if (roomState === undefined) {
            return response(404).json(notFound())
        }

        const adminError = assertAdmin(roomState, currentUser(request))
        if (adminError !== undefined) {
            return response(403).json(adminError)
        }

        roomState.room.qrCodeVisible = false
        touch(roomState.room)
        broadcastHideQRCode(roomState)

        return response(204).empty()
    }),

    http.put('/api/rooms/{roomId}/settings', async ({ params, request, response }) => {
        const roomState = getRoom(params.roomId)
        if (roomState === undefined) {
            return response(404).json(notFound())
        }

        const user = currentUser(request)
        const adminError = assertAdmin(roomState, user)
        if (adminError !== undefined) {
            return response(403).json(adminError)
        }
        if (roomState.room.state === 'finished') {
            return response(409).json(conflict('finished のルームは設定変更できません。'))
        }

        const body = await readJson<UpdateGameSettingsRequest>(request)
        if (!validateSettingsInput(body?.settings)) {
            return response(400).json(
                badRequest('settings.name と settings.description が必要です。'),
            )
        }

        const fallbackAdminIds = roomState.room.settings.admins.map((admin) => admin.userId)
        const settings = settingsFromInput(body.settings, fallbackAdminIds)
        if (settings === undefined) {
            return response(400).json(badRequest('adminUserIds を指定する場合は空にできません。'))
        }

        roomState.room.settings = settings
        touch(roomState.room)
        broadcastGameSettingsUpdated(roomState)

        return response(200).json(settings)
    }),
]
