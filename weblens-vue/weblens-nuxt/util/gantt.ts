import type { TaskInfo } from '@ethanrous/weblens-api'

export type GanttBar = {
    task: TaskInfo
    startMs: number
    endMs: number
    running: boolean
}

export type GanttLane = {
    key: string
    label: string
    bars: GanttBar[]
}

export type TimeDomain = {
    minMs: number
    maxMs: number
}

export type GanttModel = {
    lanes: GanttLane[]
    domain: TimeDomain
    queuedCount: number
}

const MIN_BAR_PX = 6
const AXIS_TARGET_PX = 120
const TICK_STEPS_MS = [
    1_000,
    5_000,
    15_000,
    30_000,
    60_000,
    5 * 60_000,
    15 * 60_000,
    30 * 60_000,
    60 * 60_000,
    3 * 60 * 60_000,
    6 * 60 * 60_000,
    12 * 60 * 60_000,
    24 * 60 * 60_000,
]

// parseTimeMs returns epoch ms for an ISO string, or 0 for missing/zero/invalid times.
export function parseTimeMs(t?: string): number {
    if (!t) {
        return 0
    }
    const ms = Date.parse(t)
    if (isNaN(ms) || ms <= 0) {
        return 0
    }
    return ms
}

// isQueued is true for tasks waiting for a worker (no assignment yet).
export function isQueued(task: TaskInfo): boolean {
    return task.State === 'InQueue' || task.State === 'Created'
}

// isRunning is true for tasks actively executing (not queued, not finished).
export function isRunning(task: TaskInfo): boolean {
    return !task.Completed && (task.State === 'Executing' || task.State === 'Sleeping')
}

// stateColorClass maps a task to a Tailwind background token for its bar.
export function stateColorClass(task: TaskInfo): string {
    if (isQueued(task)) {
        return 'bg-warn/50'
    }
    if (isRunning(task)) {
        return 'bg-theme-primary'
    }
    switch (task.status) {
        case 'success':
            return 'bg-valid'
        case 'error':
            return 'bg-danger'
        case 'canceled':
            return 'bg-graphite-400'
        default:
            // Finished but final status not captured (non-persistent task evicted before a poll saw it).
            return 'bg-graphite-600'
    }
}

// mergeTaskPoll folds a fresh poll into the accumulated session history. Polled tasks
// overwrite their prior copy; tasks that vanished while still open are closed out at nowMs
// (the backend evicts non-persistent tasks on completion, so this is our only finish signal).
export function mergeTaskPoll(prev: Map<string, TaskInfo>, polled: TaskInfo[], nowMs: number): Map<string, TaskInfo> {
    const next = new Map(prev)
    const polledIds = new Set<string>()

    for (const task of polled) {
        polledIds.add(task.taskID)
        next.set(task.taskID, task)
    }

    for (const [id, task] of prev) {
        if (polledIds.has(id) || task.Completed || parseTimeMs(task.finishTime) > 0) {
            continue
        }
        next.set(id, {
            ...task,
            Completed: true,
            State: 'Exited',
            finishTime: new Date(nowMs).toISOString(),
        })
    }

    return next
}

// barTiming returns the [start, end] window for a task's bar: start..(finish or now).
function barTiming(task: TaskInfo, nowMs: number): { startMs: number; endMs: number } {
    const start = parseTimeMs(task.startTime) || parseTimeMs(task.queueTime) || nowMs
    const finish = parseTimeMs(task.finishTime)
    return { startMs: start, endMs: finish > 0 ? finish : nowMs }
}

// buildGantt groups tasks into lanes (queued first, then worker lanes by id) and computes
// the time domain spanning all bars and now. minSpanMs extends the domain left so the
// timeline always covers at least that window back from now (keeps "now" pinned right).
export function buildGantt(tasks: TaskInfo[], nowMs: number, minSpanMs = 0): GanttModel {
    const laneMap = new Map<string, GanttLane>()

    let minMs = Number.POSITIVE_INFINITY
    let maxMs = nowMs
    let queuedCount = 0

    for (const task of tasks) {
        // Queued tasks have no meaningful timeline span; surface them as a count, not bars.
        if (isQueued(task)) {
            queuedCount++
            continue
        }

        const { startMs, endMs } = barTiming(task, nowMs)
        const key = `worker-${task.workerID}`

        let lane = laneMap.get(key)
        if (!lane) {
            lane = { key, label: `Worker ${task.workerID}`, bars: [] }
            laneMap.set(key, lane)
        }
        lane.bars.push({ task, startMs, endMs, running: isRunning(task) })

        minMs = Math.min(minMs, startMs)
        maxMs = Math.max(maxMs, endMs)
    }

    if (!isFinite(minMs)) {
        minMs = nowMs - 60_000
    }

    if (minSpanMs > 0) {
        minMs = Math.min(minMs, nowMs - minSpanMs)
    }

    const lanes = Array.from(laneMap.values()).sort((a, b) => workerNum(a.key) - workerNum(b.key))

    for (const lane of lanes) {
        lane.bars.sort((a, b) => a.startMs - b.startMs)
    }

    return { lanes, domain: { minMs, maxMs }, queuedCount }
}

function workerNum(key: string): number {
    const n = Number.parseInt(key.replace('worker-', ''), 10)
    return isNaN(n) ? 0 : n
}

export function totalWidthPx(domain: TimeDomain, pxPerMs: number): number {
    return Math.max(0, (domain.maxMs - domain.minMs) * pxPerMs)
}

export function barLeftPx(startMs: number, domain: TimeDomain, pxPerMs: number): number {
    return (startMs - domain.minMs) * pxPerMs
}

export function barWidthPx(startMs: number, endMs: number, pxPerMs: number): number {
    return Math.max(MIN_BAR_PX, (endMs - startMs) * pxPerMs)
}

// chooseTickStepMs picks a "nice" axis interval so ticks land roughly AXIS_TARGET_PX apart.
export function chooseTickStepMs(pxPerMs: number): number {
    const targetMs = AXIS_TARGET_PX / pxPerMs
    for (const step of TICK_STEPS_MS) {
        if (step >= targetMs) {
            return step
        }
    }
    return TICK_STEPS_MS[TICK_STEPS_MS.length - 1]
}

export function axisTicks(domain: TimeDomain, stepMs: number): number[] {
    const ticks: number[] = []
    const first = Math.ceil(domain.minMs / stepMs) * stepMs
    for (let t = first; t <= domain.maxMs; t += stepMs) {
        ticks.push(t)
    }
    return ticks
}

export function formatClock(ms: number): string {
    return new Date(ms).toLocaleTimeString()
}
